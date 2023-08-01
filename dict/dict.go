package dict

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"maps"
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
	Terms         []string       `json:"-"` // matched terms (will be normalized)
	Name          string         `json:"w"`
	Pronunciation string         `json:"p,omitempty"`
	MeaningGroups []EntryMeaning `json:"m"`
	Info          string         `json:"i,omitempty"` // e.g., etymology
	Source        string         `json:"s,omitempty"`
}

// EntryMeaning contains the definitions for one sub-form of a word.
//
// Info[0] — Info[*] <-- like parts of speech, word forms, etc
//  1. [Meanings[0].Tags[0]] [Meanings[0].Tags[*]] Meanings[0].Text
//     Meanings[0].Example <-- can be disabled
//  2. Meanings[*]
type EntryMeaning struct {
	Info         []string           `json:"i,omitempty"`
	Meanings     []EntryMeaningItem `json:"m"`
	WordVariants []string           `json:"v"` // for sorting results by relevance; not used for matching or display, so it can be imperfect (note that the headword is also checked first)
}

// EntryMeaningItem contains a single definition.
type EntryMeaningItem struct {
	Tags     []string `json:"t,omitempty"`
	Text     string   `json:"x"`
	Examples []string `json:"s,omitempty"`
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
	ds := maps.Keys(dict)
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
			// offsets
			for index, offset := 0, 4*b.shardSize+4; index <= b.shardSize; index++ {
				binary.Write(w, binary.BigEndian, uint32(offset))

				if shard*b.shardSize+index < len(b.entries) {
					buf, err := json.Marshal(b.entries[shard*b.shardSize+index])
					if err != nil {
						return err
					}
					offset += len(buf)
				}
			}

			// data
			for index := 0; index < b.shardSize; index++ {
				if shard*b.shardSize+index < len(b.entries) {
					buf, err := json.Marshal(b.entries[shard*b.shardSize+index])
					if err != nil {
						return err
					}
					w.Write(buf)
				}
			}

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
