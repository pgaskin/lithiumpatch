package dict

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed dict.js
var dictJS []byte

// Entry contains a single result.
//
//	Name • Pronunciation
//	...
//	Info
//	Source
type Entry struct {
	Terms         []string       // matched terms (will be normalized)
	Name          string         //
	Pronunciation string         // optional
	MeaningGroups []EntryMeaning //
	Info          string         // optional; e.g., etymology
	Source        string         // optional
}

// EntryMeaning contains the definitions for one sub-form of a word.
//
// Info[0] — Info[*] <-- like parts of speech, word forms, etc
//  1. [Meanings[0].Tags[0]] [Meanings[0].Tags[*]] Meanings[0].Text
//     Meanings[0].Example <-- can be disabled
//  2. Meanings[*]
type EntryMeaning struct {
	Info         []string           // optional
	Meanings     []EntryMeaningItem //
	WordVariants []string           // for sorting results by relevance; not used for matching or display, so it can be imperfect (note that the headword is also checked first)
}

// EntryMeaningItem contains a single definition.
type EntryMeaningItem struct {
	Tags     []string // optional
	Text     string   //
	Examples []string // optional
}

// JS gets the javascript for reading the parsed dictionaries.
func JS() []byte {
	return slices.Clone(dictJS)
}

// ParseFunc parses a dictionary.
type ParseFunc func() ([]Entry, error)

var dict = map[string]ParseFunc{}

// Register adds a dictionary to be parsed when [Parse] is called.
func Register(name string, parse ParseFunc) {
	if _, exists := dict[name]; exists {
		panic("dict: " + name + " already exists")
	}
	dict[name] = parse
}

// Dicts gets the registered dictionary names.
func Dicts() []string {
	ds := make([]string, len(dict))[:0]
	for k := range dict {
		ds = append(ds, k)
	}
	slices.Sort(ds)
	return ds
}

var dictParsed = map[string][]Entry{}

// Parse parses dictionaries.
func Parse(verbose bool) error {
	for _, d := range Dicts() {
		if _, done := dictParsed[d]; !done {
			p, err := dict[d]()
			if err != nil {
				return fmt.Errorf("%s: %w", d, err)
			}
			dictParsed[d] = p
		}
		if verbose {
			seen := map[string]struct{}{}
			for _, x := range dictParsed[d] {
				for _, t := range x.Terms {
					seen[Normalize(t)] = struct{}{}
				}
			}
			fmt.Printf("... %s (%d terms, %d entries)\n", d, len(seen), len(dictParsed[d]))
		}
	}
	return nil
}

// Build builds all dictionaries into subdirectories of the provided path, which
// should be empty.
func Build(path string) error {
	for _, d := range Dicts() {
		if _, done := dictParsed[d]; !done {
			return fmt.Errorf("build %s: not parsed yet", d)
		}
		if err := BuildDict(filepath.Join(path, d), dictParsed[d]); err != nil {
			return fmt.Errorf("build %s: %w", d, err)
		}
	}
	return nil
}

type builder struct {
	output       string
	entries      []Entry
	indexBuckets [][]builderIndexEntry
	shardSize    int
}

type builderIndexEntry struct {
	Term  string
	Entry int
}

// BuildDict builds a single dictionary into the provided path.
func BuildDict(path string, dict []Entry) error {
	return (&builder{
		output:  path,
		entries: dict,
	}).run()
}

func (b *builder) run() error {
	b.shardSize = 512

	// normalize, sort, and deduplicate the lookup terms
	for xi := range b.entries {
		ts := make([]string, len(b.entries[xi].Terms))[:0]
		for _, t := range b.entries[xi].Terms {
			if t = Normalize(t); t != "" {
				ts = append(ts, t)
			}
		}
		slices.Sort(ts)
		ts = slices.Compact(ts)
		b.entries[xi].Terms = ts
	}

	// build the term index buckets
	for xi, x := range b.entries {
		for _, y := range x.Terms {
			e := builderIndexEntry{
				Term:  y,
				Entry: xi,
			}
			for len(y) > len(b.indexBuckets) {
				b.indexBuckets = append(b.indexBuckets, nil)
			}
			b.indexBuckets[len(y)-1] = append(b.indexBuckets[len(y)-1], e)
		}
	}

	// sort the term index buckets (for binary searches)
	// note: there may be multiple entries for a term, so we don't dedupe
	for _, x := range b.indexBuckets {
		sort.Slice(x, func(i, j int) bool {
			return bytes.Compare([]byte(x[i].Term), []byte(x[j].Term)) < 0
		})
	}

	// write the index buckets
	if err := b.create("index", func(w *bytes.Buffer) error {
		// shard size
		binary.Write(w, binary.BigEndian, uint32(b.shardSize))

		// number of buckets (i.e., max word length)
		binary.Write(w, binary.BigEndian, uint32(len(b.indexBuckets)))

		// bucket sizes
		for _, x := range b.indexBuckets {
			binary.Write(w, binary.BigEndian, uint32(len(x)))
		}

		// bucket words
		for _, x := range b.indexBuckets {
			for _, y := range x {
				w.WriteString(y.Term)
			}
		}

		// bucket indexes
		for _, x := range b.indexBuckets {
			for _, y := range x {
				binary.Write(w, binary.BigEndian, uint32(y.Entry))
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	// write the shards
	for shard := 0; shard < (len(b.entries)+b.shardSize-1)/b.shardSize; shard++ {
		if err := b.create(fmt.Sprintf("%03x", shard), func(w *bytes.Buffer) error {
			buf := make([]byte, (b.shardSize+1)*4)

			for idx, e := range b.entries[shard*b.shardSize:] {
				// offset
				binary.BigEndian.PutUint32(buf[idx*4:], uint32(len(buf)))
				if idx >= b.shardSize {
					break
				}

				// Name
				buf = binary.BigEndian.AppendUint32(buf, uint32(len(e.Name)))
				buf = append(buf, e.Name...)

				// Pronunciation
				buf = binary.BigEndian.AppendUint32(buf, uint32(len(e.Pronunciation)))
				buf = append(buf, e.Pronunciation...)

				// MeaningGroups
				buf = binary.BigEndian.AppendUint32(buf, uint32(len(e.MeaningGroups)))
				for _, mg := range e.MeaningGroups {

					// Info
					buf = binary.BigEndian.AppendUint32(buf, uint32(len(mg.Info)))
					for _, v := range mg.Info {

						// item
						buf = binary.BigEndian.AppendUint32(buf, uint32(len(v)))
						buf = append(buf, v...)
					}

					// Meanings
					buf = binary.BigEndian.AppendUint32(buf, uint32(len(mg.Meanings)))
					for _, m := range mg.Meanings {

						// Tags
						buf = binary.BigEndian.AppendUint32(buf, uint32(len(m.Tags)))
						for _, v := range m.Tags {

							// item
							buf = binary.BigEndian.AppendUint32(buf, uint32(len(v)))
							buf = append(buf, v...)
						}

						// Text
						buf = binary.BigEndian.AppendUint32(buf, uint32(len(m.Text)))
						buf = append(buf, m.Text...)

						// Examples
						buf = binary.BigEndian.AppendUint32(buf, uint32(len(m.Examples)))
						for _, v := range m.Examples {

							// item
							buf = binary.BigEndian.AppendUint32(buf, uint32(len(v)))
							buf = append(buf, v...)
						}
					}

					// WordVariants
					buf = binary.BigEndian.AppendUint32(buf, uint32(len(mg.WordVariants)))
					for _, v := range mg.WordVariants {

						// item
						buf = binary.BigEndian.AppendUint32(buf, uint32(len(v)))
						buf = append(buf, v...)
					}
				}

				// Info
				buf = binary.BigEndian.AppendUint32(buf, uint32(len(e.Info)))
				buf = append(buf, e.Info...)

				// Source
				buf = binary.BigEndian.AppendUint32(buf, uint32(len(e.Source)))
				buf = append(buf, e.Source...)
			}

			w.Write(buf)

			return nil
		}); err != nil {
			return fmt.Errorf("write shard %d: %w", shard, err)
		}
	}

	return nil
}

func (b *builder) create(name string, fn func(w *bytes.Buffer) error) error {
	var w bytes.Buffer
	if err := fn(&w); err != nil {
		return fmt.Errorf("generate %s: %w", name, err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(b.output, name)), 0777); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(b.output, name), w.Bytes(), 0666)
}

// Normalize reduces term to a limited set of ASCII characters.
func Normalize(term string) string {
	// decompose accents and stuff
	// convert similar characters with only stylistic differences
	// other unicode normalization stuff
	term = norm.NFKC.String(term)

	// to lowercase (unicode-aware)
	term = strings.ToLower(term)

	// normalize whitespace
	term = strings.Join(strings.FieldsFunc(term, unicode.IsSpace), " ")

	// replace smart punctuation
	term = strings.Map(func(r rune) rune {
		switch r {
		case '\u00ab':
			return '"'
		case '\u00bb':
			return '"'
		case '\u2010':
			return '-'
		case '\u2011':
			return '-'
		case '\u2012':
			return '-'
		case '\u2013':
			return '-'
		case '\u2014':
			return '-'
		case '\u2015':
			return '-'
		case '\u2018':
			return '\''
		case '\u2019':
			return '\''
		case '\u201a':
			return '\''
		case '\u201b':
			return '\''
		case '\u201c':
			return '"'
		case '\u201d':
			return '"'
		case '\u201e':
			return '"'
		case '\u201f':
			return '"'
		case '\u2024':
			return '.'
		case '\u2032':
			return '\''
		case '\u2033':
			return '"'
		case '\u2035':
			return '\''
		case '\u2036':
			return '"'
		case '\u2038':
			return '^'
		case '\u2039':
			return '\''
		case '\u203a':
			return '\''
		case '\u204f':
			return ';'
		default:
			return r
		}
	}, term)

	// expand ligatures
	term = strings.NewReplacer(
		"\ua74f", `oo`,
		"\u00df", `ss`,
		"\u00e6", `ae`,
		"\u0153", `oe`,
		"\ufb00", `ff`,
		"\ufb01", `fi`,
		"\ufb02", `fl`,
		"\ufb03", `ffi`,
		"\ufb04", `ffl`,
		"\ufb05", `ft`,
		"\ufb06", `st`,
		"\u2025", `..`,
		"\u2026", `...`,
		"\u2042", `***`,
		"\u2047", `??`,
		"\u2048", `?!`,
		"\u2049", `!?`,
	).Replace(term)

	// normalize dashes
	term = strings.Join(strings.FieldsFunc(term, func(r rune) bool {
		return r == '-'
	}), "-")

	// remove unknown characters/diacritics
	// note: since we decomposed diacritics, this will leave the base char
	term = strings.Map(func(r rune) rune {
		if 'a' <= r && r <= 'z' {
			return r
		}
		if '0' <= r && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '\'' || r == '_' || r == '.' || r == ',' {
			return r
		}
		return -1
	}, term)

	return term
}
