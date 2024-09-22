package webster1913

import (
	"embed"
	"io"
	"io/fs"
	"regexp"
	"strings"

	"github.com/pgaskin/lithiumpatch/dict"
)

//go:embed *
var assets embed.FS

func exists(name string) bool {
	_, err := fs.Stat(assets, name)
	return err == nil
}

func init() {
	if exists("webster1913.txt") {
		dict.Register("webster1913", -50, func() ([]dict.Entry, error) {
			f, err := assets.Open("webster1913.txt")
			if err != nil {
				return nil, err
			}
			defer f.Close()

			return Parse(f)
		})
	}
}

var meaningTagRe = regexp.MustCompile(`^\s*\(((?:\s*[A-Za-z][A-Za-z0-9]+\.?\s*)+)\)\s*`)

func Parse(r io.Reader) ([]dict.Entry, error) {
	wbd, err := ParseDict(r)
	if err != nil {
		return nil, err
	}

	var entries []dict.Entry
	for _, e := range wbd {
		var ew dict.Entry
		ew.Terms = append(ew.Terms, e.Headword)
		ew.Name = e.Headword
		ew.Info = e.Etymology
		ew.Source = "Webster's 1913 Unabridged Dictionary"
		var ewm dict.EntryMeaning
		if e.Info != "" {
			ewm.Info = append(ewm.Info, e.Info)
		}
		if len(e.Variant) != 0 {
			ewm.Info = append(ewm.Info, strings.Join(e.Variant, ","))
			ew.Terms = append(ew.Terms, e.Variant...)
		}
		for _, m := range e.Meanings {
			var ewmi dict.EntryMeaningItem
			ewmi.Text = m.Text
			if tmp := meaningTagRe.FindStringSubmatch(ewmi.Text); tmp != nil {
				ewmi.Text = strings.TrimPrefix(ewmi.Text, tmp[0])
				ewmi.Tags = append(ewmi.Tags, tmp[1])
			}
			if m.Example != "" {
				ewmi.Examples = append(ewmi.Tags, m.Example)
			}
			ewm.Meanings = append(ewm.Meanings, ewmi)
			ewm.WordVariants = append(ewm.WordVariants, e.Variant...)
		}
		ew.MeaningGroups = append(ew.MeaningGroups, ewm)
		entries = append(entries, ew)
	}
	return entries, nil
}
