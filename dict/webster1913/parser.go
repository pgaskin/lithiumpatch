package webster1913

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"
)

// Dict represents the parsed dictionary.
type Dict []*Entry

// Entry is a single dictionary entry.
type Entry struct {
	Headword    string
	Variant     []string
	Info        string
	Etymology   string
	Meanings    []*EntryMeaning
	Synonyms    []string
	PhraseDefns []string
	Extra       string // unparseable text
}

// EntryMeaning is a meaning for a dictionary entry.
type EntryMeaning struct {
	Text    string
	Example string
}

var (
	entryWordRe         = regexp.MustCompile(`^[A-Z_ ;-]+$`)
	numberedDefnStartRe = regexp.MustCompile(`^[0-9]+\.\s*`)
	singleDefnStartRe   = regexp.MustCompile(`^Defn:\s+`)
	noteStartRe         = regexp.MustCompile(`^\s*Note:\s+`)
	synStartRe          = regexp.MustCompile(`^Syn.\s*$`)
	synItemStartRe      = regexp.MustCompile(`^\s+--\s+`)
	phraseDefnStartRe   = regexp.MustCompile(`^\s+--\s+([A-Za-z ]+?[A-Za-z])\s*(\([^)]+\))?[,.]\s*`)
	wordInfoFormRe      = regexp.MustCompile(`(?:p\. p\.|vb\. n\.|p\. pr\.) +([A-Z][a-z]+)[:;.,]`)
)

type state int

const (
	// StateNone is before the first entry.
	StateNone state = iota
	// StateEntryInfo is at the beginning of the entry.
	StateEntryInfo
	// StateEntryExtra is unclassified text in the entry.
	StateEntryExtra
	// StateEntryMeaningText is inside an entry's meaning's text.
	StateEntryMeaningText
	// StateEntryMeaningExample is inside an entry's meaning's example.
	StateEntryMeaningExample
	// StateEntrySynonym is inside an entry's synonym list.
	StateEntrySynonym
	// StateEntryPhraseDefn is inside an entry's phrase definition list.
	StateEntryPhraseDefn
)

// ParseDict parses Project Gutenberg's Webster's Unabridged Dictionary.
//
// http://www.gutenberg.org/ebooks/29765.txt.utf-8
func ParseDict(r io.Reader) (Dict, error) {
	var wd Dict
	var perr error
	sc := bufio.NewScanner(r)

	var state state
	var entry *Entry
	var meaning *EntryMeaning
	var i int
	for sc.Scan() {
		ln := sc.Bytes()
		lnt := bytes.TrimSpace(ln)
		blankLine := len(lnt) == 0

		if bytes.HasPrefix(lnt, []byte("*** END")) {
			break
		}

		if entryWordRe.Match(ln) {
			if state == StateNone {
				// skip the file header(up to the word "A")
				if !bytes.Equal(lnt, []byte{'A'}) {
					continue
				}
			}
			if bytes.Count(lnt, []byte{'-'}) != len(lnt) {
				// ^ if all dashes, it is a false positive
				spl := strings.Split(string(bytes.ToLower(ln)), ";")
				entry = &Entry{Headword: strings.TrimSpace(spl[0])}
				if len(spl) > 1 {
					for _, v := range spl[1:] {
						if w := strings.TrimSpace(v); w != "" {
							entry.Variant = append(entry.Variant, w)
						}
					}
				}
				meaning = nil
				wd = append(wd, entry)
				state = StateEntryInfo
				continue
			}
		}

		switch state {
		case StateNone:
			// ignore any text before the first entry
		case StateEntryInfo:
			switch {
			case blankLine:
				for _, m := range wordInfoFormRe.FindAllStringSubmatch(entry.Info, -1) {
					entry.Variant = append(entry.Variant, strings.ToLower(m[1]))
				}
				// attempt to split into etymology
				if spl := strings.SplitN(entry.Info, " Etym: ", 2); len(spl) == 2 {
					entry.Info = strings.TrimSpace(spl[0])
					entry.Etymology = strings.TrimSpace(spl[1])
				}
				state = StateEntryExtra
			default:
				entry.Info += " " + string(lnt)
			}
		case StateEntryExtra:
			switch {
			case singleDefnStartRe.Match(ln):
				meaning = &EntryMeaning{Text: string(singleDefnStartRe.ReplaceAllLiteral(ln, nil))}
				entry.Meanings = append(entry.Meanings, meaning)
				state = StateEntryMeaningText
			case numberedDefnStartRe.Match(ln):
				meaning = &EntryMeaning{Text: string(numberedDefnStartRe.ReplaceAllLiteral(ln, nil))}
				entry.Meanings = append(entry.Meanings, meaning)
				state = StateEntryMeaningText
			case phraseDefnStartRe.Match(ln):
				meaning = nil
				entry.PhraseDefns = append(entry.PhraseDefns, string(bytes.TrimSpace(bytes.Replace(lnt, []byte("--"), nil, 1))))
				entry.Variant = append(entry.Variant, string(bytes.ToLower(phraseDefnStartRe.FindSubmatch(ln)[1])))
				state = StateEntryPhraseDefn
			case blankLine:
				// ignore
			default:
				entry.Extra += " " + string(lnt)
			}
		case StateEntryMeaningText:
			switch {
			case synStartRe.Match(ln):
				meaning = nil
				state = StateEntrySynonym
			case singleDefnStartRe.Match(ln):
				// if it is in any kind of definition (single/numbered), it is part of it.
				meaning.Text += " " + string(singleDefnStartRe.ReplaceAllLiteral(lnt, nil))
			case numberedDefnStartRe.Match(ln):
				meaning = &EntryMeaning{Text: string(numberedDefnStartRe.ReplaceAllLiteral(ln, nil))}
				entry.Meanings = append(entry.Meanings, meaning)
				state = StateEntryMeaningText
			case phraseDefnStartRe.Match(ln):
				meaning = nil
				entry.PhraseDefns = append(entry.PhraseDefns, string(bytes.TrimSpace(bytes.Replace(lnt, []byte("--"), nil, 1))))
				entry.Variant = append(entry.Variant, string(bytes.ToLower(phraseDefnStartRe.FindSubmatch(ln)[1])))
				state = StateEntryPhraseDefn
			case len(meaning.Text) > 5 && len(lnt) < 55 && bytes.HasSuffix(lnt, []byte{'.'}) && !noteStartRe.Match(ln):
				// if there is already some body text, it is not a hard-wrapped
				// line, and it ends with a period, and is not a note, then it's
				// the last line of the text before the example.
				meaning.Text += " " + string(lnt)
				state = StateEntryMeaningExample
			case blankLine:
				// ignore
			default:
				meaning.Text += " " + string(lnt)
			}
		case StateEntryMeaningExample:
			switch {
			case synStartRe.Match(ln):
				meaning = nil
				state = StateEntrySynonym
			case singleDefnStartRe.Match(ln):
				meaning = &EntryMeaning{Text: string(singleDefnStartRe.ReplaceAllLiteral(ln, nil))}
				entry.Meanings = append(entry.Meanings, meaning)
				state = StateEntryMeaningText
			case numberedDefnStartRe.Match(ln):
				meaning = &EntryMeaning{Text: string(numberedDefnStartRe.ReplaceAllLiteral(ln, nil))}
				entry.Meanings = append(entry.Meanings, meaning)
				state = StateEntryMeaningText
			case phraseDefnStartRe.Match(ln):
				meaning = nil
				entry.PhraseDefns = append(entry.PhraseDefns, string(bytes.TrimSpace(bytes.Replace(lnt, []byte("--"), nil, 1))))
				entry.Variant = append(entry.Variant, string(bytes.ToLower(phraseDefnStartRe.FindSubmatch(ln)[1])))
				state = StateEntryPhraseDefn
			case blankLine:
				// ignore
			default:
				if meaning.Example != "" {
					meaning.Example += " "
				}
				meaning.Example += string(lnt)
			}
		case StateEntrySynonym:
			switch {
			case blankLine:
				state = StateEntryExtra
			case synItemStartRe.Match(ln):
				entry.Synonyms = append(entry.Synonyms, string(synItemStartRe.ReplaceAllLiteral(ln, nil)))
			case len(entry.Synonyms) == 0:
				// there was a "Syn." without any valid synonyms under it
				state = StateEntryExtra
			case phraseDefnStartRe.Match(ln):
				meaning = nil
				entry.PhraseDefns = append(entry.PhraseDefns, string(bytes.TrimSpace(bytes.Replace(lnt, []byte("--"), nil, 1))))
				entry.Variant = append(entry.Variant, string(bytes.ToLower(phraseDefnStartRe.FindSubmatch(ln)[1])))
				state = StateEntryPhraseDefn
			default:
				entry.Synonyms[len(entry.Synonyms)-1] += " " + string(lnt)
			}
		case StateEntryPhraseDefn:
			switch {
			case phraseDefnStartRe.Match(ln):
				meaning = nil
				entry.PhraseDefns = append(entry.PhraseDefns, string(bytes.TrimSpace(bytes.Replace(lnt, []byte("--"), nil, 1))))
				entry.Variant = append(entry.Variant, string(bytes.ToLower(phraseDefnStartRe.FindSubmatch(ln)[1])))
				state = StateEntryPhraseDefn
			case blankLine:
				// allow a blank line to end it for reducing the chance of bugs.
				state = StateEntryExtra
			default:
				// phrase definitions are always last, so no need for checking
				// for any other state changes (e.g. the start of a numbered
				// definition) (and the previous case should deal with any
				// edge-cases).
				entry.PhraseDefns[len(entry.PhraseDefns)-1] += " " + string(lnt)
			}
		}
		i++
	}

	if serr := sc.Err(); serr != nil {
		return nil, serr
	}
	if perr != nil {
		return nil, perr
	}
	return wd, nil
}
