package fonts

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
)

//go:embed *.ttf */*.ttf
var embedded embed.FS

func init() {
	if _, err := LoadFrom(embedded); err != nil {
		panic(fmt.Errorf("parse embedded fonts: %w", err))
	}
}

var fonts sync.Map

type Font struct {
	Name       string
	Base       string
	Script     Script
	Regular    []byte
	Bold       []byte
	Italic     []byte
	BoldItalic []byte
}

func (f Font) String() string {
	var s []string
	if f.Regular != nil {
		s = append(s, "Regular")
	}
	if f.Bold != nil {
		s = append(s, "Bold")
	}
	if f.Italic != nil {
		s = append(s, "Italic")
	}
	if f.BoldItalic != nil {
		s = append(s, "BoldItalic")
	}
	return fmt.Sprintf("%s (%s) [%s] {%s}", f.Name, f.Base, f.Script, strings.Join(s, ","))
}

var naRe = regexp.MustCompile(`[^A-Za-z]+`)

func LoadFrom(fsys fs.FS) (int, error) {
	xm := map[string]Font{}
	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		buf, err := fs.ReadFile(fsys, path) // doesn't allocate a new byte slice for embed.FS
		if err != nil {
			return fmt.Errorf("process %q: %w", path, err)
		}
		ttf, err := truetype.Parse(buf)
		if err != nil {
			return fmt.Errorf("process %q: parse ttf: %w", path, err)
		}
		ff := ttf.Name(truetype.NameIDFontFamily)
		if ff == "" {
			return fmt.Errorf("process %q: parse ttf: no font family name", path)
		}
		f := xm[ff]
		switch fsf := ttf.Name(truetype.NameIDFontSubfamily); fsf {
		case "Regular", "Roman":
			f.Name = ff
			f.Base = naRe.ReplaceAllLiteralString(ff, "")
			switch filepath.Dir(path) {
			case "latin":
				f.Script = FontScriptLatin
			case "cyrillic":
				f.Script = FontScriptCyrillic
			case "greek":
				f.Script = FontScriptGreek
			case "thai":
				f.Script = FontScriptThai
			default:
				f.Script = FontScriptAll.Filter(func(c rune) bool {
					return ttf.Index(c) != 0
				})
			}
			if f.Regular != nil {
				return fmt.Errorf("process %q: already have font face %q %s", path, fsf, ff)
			}
			f.Regular = buf
		case "Bold":
			if f.Bold != nil {
				return fmt.Errorf("process %q: already have font face %q %s", path, fsf, ff)
			}
			f.Bold = buf
		case "Italic":
			if f.Italic != nil {
				return fmt.Errorf("process %q: already have font face %q %s", path, fsf, ff)
			}
			f.Italic = buf
		case "Bold Italic":
			if f.BoldItalic != nil {
				return fmt.Errorf("process %q: already have font face %q %s", path, fsf, ff)
			}
			f.BoldItalic = buf
		case "":
			return fmt.Errorf("process %q: parse ttf: no subfamily name", path)
		default:
			return fmt.Errorf("process %q: unsupported subfamily %q", path, fsf)
		}
		xm[ff] = f
		return nil
	}); err != nil {
		return 0, err
	}
	for xf, x := range xm {
		if x.Regular == nil {
			return 0, fmt.Errorf("missing regular subfamily for %q", xf)
		}
		if x.Script == 0 {
			return 0, fmt.Errorf("no supported scripts detected in %q", x.Name)
		}
	}
	for _, x := range xm {
		Add(x)
	}
	return len(xm), nil
}

func Add(f Font) {
	if f.Name == "" {
		panic("no name for font")
	}
	if f.Script == 0 {
		panic("no supported scripts for font")
	}
	if f.Base == "" {
		panic("no base name for font")
	}
	fonts.Store(f.Base, f)
}

func Range(fn func(Font) bool) {
	fonts.Range(func(k, v any) bool {
		return fn(v.(Font))
	})
}

func All() []Font {
	var xs []Font
	Range(func(x Font) bool {
		xs = append(xs, x)
		return true
	})
	sort.Slice(xs, func(i, j int) bool {
		return xs[i].Name < xs[j].Name
	})
	return xs
}

type Script uint

const (
	FontScriptAll      Script = FontScriptLatin | FontScriptCyrillic | FontScriptGreek | FontScriptThai
	FontScriptLatin    Script = 1
	FontScriptCyrillic Script = 2
	FontScriptGreek    Script = 4
	FontScriptThai     Script = 8
)

func (s Script) String() string {
	var b []string
	if s&FontScriptLatin != 0 {
		b = append(b, "Latin")
	}
	if s&FontScriptCyrillic != 0 {
		b = append(b, "Cyrillic")
	}
	if s&FontScriptGreek != 0 {
		b = append(b, "Greek")
	}
	if s&FontScriptThai != 0 {
		b = append(b, "Thai")
	}
	return strings.Join(b, "|")
}

func (s Script) Filter(fn func(rune) bool) Script {
	if s&FontScriptLatin != 0 {
		for _, c := range "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" {
			if !fn(c) {
				s ^= FontScriptLatin
				break
			}
		}
	}
	if s&FontScriptCyrillic != 0 {
		for _, c := range "АБВГҐДЂЃЕЁЄЖЗЗ́ЅИІЇЙЈКЛЉМНЊОПРСС́ТЋЌУЎФХЦЧЏШЩЪЫЬЭЮЯ" {
			if !fn(c) {
				s ^= FontScriptCyrillic
				break
			}
		}
	}
	if s&FontScriptGreek != 0 {
		for _, c := range "ΑαΒβΓγΔδΕεΖζΗηΘθΙιΚκΛλΜμΝνΞξΟοΠπΡρΣσΤτΥυΦφΧχΨψΩω" {
			if !fn(c) {
				s ^= FontScriptGreek
				break
			}
		}
	}
	if s&FontScriptThai != 0 {
		for _, c := range "กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรลวศษสหฬอฮะาเแโใไฤๅฦๆ" {
			if !fn(c) {
				s ^= FontScriptThai
				break
			}
		}
	}
	return s
}

func (s Script) Flags() uint {
	return uint(s)
}
