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

//go:embed lib/dict.js
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
			buf := make([]byte, b.shardSize*4)

			for idx, e := range b.entries[shard*b.shardSize:] {
				if idx >= b.shardSize {
					break
				}

				// offset
				binary.BigEndian.PutUint32(buf[idx*4:], uint32(len(buf)))

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
	var (
		n     strings.Builder
		lastS = true // trim leading whitespace
		lastD = false
	)
	n.Grow(len(term))

	// decompose accents and stuff
	// convert similar characters with only stylistic differences
	// convert all whitespace to the ascii equivalent (incl nbsp,em-space,en-space,etc->space)
	// other unicode normalization stuff
	// to lowercase (unicode-aware)
	for _, r := range norm.NFKD.String(term) {
		r = unicode.ToLower(r)

		// replace smart punctuation
		switch r {
		case 0x00ab:
			r = '"'
		case 0x00bb:
			r = '"'
		case 0x2010:
			r = '-'
		case 0x2011:
			r = '-'
		case 0x2012:
			r = '-'
		case 0x2013:
			r = '-'
		case 0x2014:
			r = '-'
		case 0x2015:
			r = '-'
		case 0x2018:
			r = '\''
		case 0x2019:
			r = '\''
		case 0x201a:
			r = '\''
		case 0x201b:
			r = '\''
		case 0x201c:
			r = '"'
		case 0x201d:
			r = '"'
		case 0x201e:
			r = '"'
		case 0x201f:
			r = '"'
		case 0x2024:
			r = '.'
		case 0x2032:
			r = '\''
		case 0x2033:
			r = '"'
		case 0x2035:
			r = '\''
		case 0x2036:
			r = '"'
		case 0x2038:
			r = '^'
		case 0x2039:
			r = '\''
		case 0x203a:
			r = '\''
		case 0x204f:
			r = ';'
		}

		// collapse whitespace
		if r == 32 || (r >= 9 && r <= 12) {
			if lastS {
				continue
			}
			lastS = true
			r = 32
		} else {
			lastS = false
		}

		// collapse dashes
		if r == 45 {
			if lastD {
				continue
			}
			lastD = true
		} else {
			lastD = false
		}

		// expand ligatures
		// remove unknown characters/diacritics
		switch r {
		case 0xa74f:
			n.WriteString(`oo`)
		case 0x00df:
			n.WriteString(`ss`)
		case 0x00e6:
			n.WriteString(`ae`)
		case 0x0153:
			n.WriteString(`oe`)
		case 0xfb00:
			n.WriteString(`ff`)
		case 0xfb01:
			n.WriteString(`fi`)
		case 0xfb02:
			n.WriteString(`fl`)
		case 0xfb03:
			n.WriteString(`ffi`)
		case 0xfb04:
			n.WriteString(`ffl`)
		case 0xfb05:
			n.WriteString(`ft`)
		case 0xfb06:
			n.WriteString(`st`)
		default:
			switch {
			case 'a' <= r && r <= 'z':
			case '0' <= r && r <= '9':
			case r == ' ' || r == '-' || r == '\'' || r == '_' || r == '.' || r == ',':
			default:
				continue
			}
			n.WriteRune(r)
		}
	}
	if lastS && n.Len() != 0 {
		// trim trailing whitespace
		return n.String()[:n.Len()-1]
	}
	return n.String()
}
