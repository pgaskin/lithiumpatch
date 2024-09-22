package edgedict

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/pgaskin/edgedict"
	"github.com/pgaskin/lithiumpatch/dict"

	_ "github.com/pgaskin/xmlwriter" // needed for edgedict-fetch
)

//go:generate go run github.com/pgaskin/edgedict/cmd/edgedict-fetch -f en-us -f en-gb

//go:embed *
var assets embed.FS

func exists(name string) bool {
	_, err := fs.Stat(assets, name)
	return err == nil
}

func init() {
	if us, gb := exists("Dictionary_EN_US.db"), exists("Dictionary_EN_GB.db"); us || gb {
		dict.Register("oxford_en", 50, func() ([]dict.Entry, error) {
			var sources []io.ReaderAt
			var names []string
			if gb {
				f, err := assets.Open("Dictionary_EN_GB.db")
				if err != nil {
					return nil, err
				}
				defer f.Close()

				sources = append(sources, f.(io.ReaderAt))
				names = append(names, "en-GB")
			}
			if us {
				f, err := assets.Open("Dictionary_EN_US.db")
				if err != nil {
					return nil, err
				}
				defer f.Close()

				sources = append(sources, f.(io.ReaderAt))
				names = append(names, "en-US")
			}
			return Parse(sources, names)
		})
	}
}

func Parse(sources []io.ReaderAt, names []string) ([]dict.Entry, error) {
	if len(sources) != len(names) {
		panic("edgedict: length of sources and names must match")
	}

	type Lref struct {
		Source int
		Ref    edgedict.Ref
	}

	var (
		entries           []dict.Entry
		dictionaries      = make([]*edgedict.Dictionary, len(sources))
		oxSeenHeadword    = map[string]int{}   // map of headword to which dictionary it came from
		oxEntryMap        = map[Lref]string{}  // map of edgedict term to headword
		oxHeadwordEntries = map[string][]int{} // map of headword to entries matching it
		oxSourceTerms     = map[string]int{}   // map of source where we first saw a term (lowercased) better dict entry deduplication
	)
	for sourceIndex, source := range names {
		edict, err := edgedict.New(sources[sourceIndex])
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", source, err)
		}
		defer edict.Close()

		dictionaries[sourceIndex] = edict
	}
	for sourceIndex, source := range names {
		if err := dictionaries[sourceIndex].WalkRefs(func(term string, _ edgedict.Ref) error {
			if _, seen := oxSourceTerms[term]; !seen {
				oxSourceTerms[strings.ToLower(term)] = sourceIndex
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("resolve terms for %s: %w", source, err)
		}
	}
	for sourceIndex, source := range names {
		if err := dictionaries[sourceIndex].Walk(func(ref edgedict.Ref, buf []byte) error {
			var e edgedict.Entry
			if err := json.Unmarshal(buf, &e); err != nil {
				return fmt.Errorf("parse entry %s: %w", ref, err)
			}
			oxEntryMap[Lref{sourceIndex, ref}] = e.Name

			if ss, seen := oxSourceTerms[strings.ToLower(e.Name)]; seen && ss != sourceIndex {
				// we've already added an entry matching this term in an earlier dictionary
				return nil
			}
			if ss, seen := oxSeenHeadword[e.Name]; !seen {
				oxSeenHeadword[e.Name] = sourceIndex
			} else if ss == sourceIndex {
				// there may be multiple entries for a headword
			} else {
				// we could attempt to merge the entries from different languages, but that's likely to result in duplicates, so just take the first seen entry
				return nil
			}

			var ew dict.Entry
			ew.Terms = append(ew.Terms, e.Name)
			ew.Name = e.Name
			ew.Pronunciation = e.Pronunciation
			ew.Info = e.WordOrigin
			ew.Source = "Oxford (" + source + ")"
			for _, g := range e.MeaningGroups {
				var ewm dict.EntryMeaning

				var t bytes.Buffer
				for i, p := range g.PartsOfSpeech {
					if i != 0 {
						t.WriteString(", ")
					}
					t.WriteString(p.Name)
				}
				if t.Len() != 0 {
					ewm.Info = append(ewm.Info, t.String())
				}

				t.Reset()
				for i, f := range g.WordForms {
					if i != 0 {
						t.WriteString(", ")
					}
					t.WriteString(f.Word.Name)

					ew.Terms = append(ew.Terms, f.Word.Name)
					ewm.WordVariants = append(ewm.WordVariants, f.Word.Name)
				}
				if t.Len() != 0 {
					ewm.Info = append(ewm.Info, t.String())
				}

				for _, m := range g.Meanings {
					for _, d := range m.RichDefinitions {
						var ewmi dict.EntryMeaningItem

						t.Reset()
						for i, x := range d.Fragments {
							if i != 0 {
								t.WriteByte(' ')
							}
							t.WriteString(x.Text)
						}
						ewmi.Text = t.String()

						ewmi.Examples = append(ewmi.Tags, d.Examples...)
						ewmi.Tags = append(ewmi.Tags, d.LabelTags...)
						ewm.Meanings = append(ewm.Meanings, ewmi)
					}
				}
				ew.MeaningGroups = append(ew.MeaningGroups, ewm)
			}
			oxHeadwordEntries[e.Name] = append(oxHeadwordEntries[e.Name], len(entries))
			entries = append(entries, ew)

			return nil
		}); err != nil {
			return nil, fmt.Errorf("add entries from %s: %w", source, err)
		}
	}
	for sourceIndex, source := range names {
		if err := dictionaries[sourceIndex].WalkRefs(func(term string, ref edgedict.Ref) error {
			if n, ok := oxEntryMap[Lref{sourceIndex, ref}]; !ok {
				panic("wtf: we should have seen this ref before...")
			} else {
				for _, e := range oxHeadwordEntries[n] {
					entries[e].Terms = append(entries[e].Terms, term)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("add term mappings from %s: %w", source, err)
		}
	}
	return entries, nil
}
