// Package patchdef contains helpers for patching Android apps.
package patchdef

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"unicode/utf8"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

var patches sync.Map

func Register(name string, inst ...Instruction) {
	if name == "" {
		panic("missing patch name")
	}
	if _, exists := patches.LoadOrStore(name, &Patch{name, inst}); exists {
		panic(fmt.Sprintf("duplicate patch %q", name))
	}
}

func Patches() []*Patch {
	var ps []*Patch
	patches.Range(func(key, value any) bool {
		ps = append(ps, value.(*Patch))
		return true
	})
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].name < ps[j].name
	})
	return ps
}

type Patch struct {
	name string
	inst []Instruction
}

func (p Patch) String() string {
	return p.name
}

func (p Patch) Name() string {
	return p.name
}

func (p Patch) Apply(apk string, diffwriter io.Writer) error {
	for i, inst := range p.inst {
		if err := inst.Do(apk, diffwriter); err != nil {
			return fmt.Errorf("apply patch %q: inst %d: %w", p.name, i, err)
		}
	}
	return nil
}

type Instruction interface {
	Do(apk string, diffwriter io.Writer) error
}

type writeInst struct {
	To   string
	Data []byte
}

func WriteFile(to string, content []byte) Instruction {
	return &writeInst{to, content}
}

func WriteFileString(to, content string) Instruction {
	return &writeInst{to, []byte(content)}
}

func (w *writeInst) Do(apk string, diffwriter io.Writer) error {
	if utf8.Valid(w.Data) {
		diff := gotextdiff.ToUnified(
			"/dev/null", "b/"+w.To, "",
			myers.ComputeEdits(span.URIFromPath(w.To), "", string(w.Data)),
		)
		if _, err := fmt.Fprint(diffwriter, diff); err != nil {
			return fmt.Errorf("write diff: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(diffwriter, "--- /dev/null\n+++ b/%s\nBinary file\n", w.To); err != nil {
			return fmt.Errorf("write diff: %w", err)
		}
	}

	p := filepath.Join(apk, filepath.Clean(filepath.FromSlash(w.To)))
	if err := os.MkdirAll(filepath.Dir(p), 0777); err != nil {
		return err
	}
	if err := os.WriteFile(p, []byte(w.Data), 0666); err != nil {
		return err
	}
	return nil
}

type deleteInst struct {
	Name string
}

func DeleteFile(name string) Instruction {
	return &deleteInst{name}
}

func (d *deleteInst) Do(apk string, diffwriter io.Writer) error {
	if _, err := fmt.Fprintf(diffwriter, "--- a/%s\n+++ /dev/null\nBinary file\n", d.Name); err != nil {
		return fmt.Errorf("write diff: %w", err)
	}

	p := filepath.Join(apk, filepath.Clean(filepath.FromSlash(d.Name)))
	return os.Remove(p)
}

type patchInst struct {
	Sources  []string
	Patchers []StringPatcher
}

func PatchFile(src string, pt ...StringPatcher) *patchInst {
	return &patchInst{[]string{src}, pt}
}

func PatchFiles(src []string, pt ...StringPatcher) *patchInst {
	return &patchInst{src, pt}
}

func (p *patchInst) Do(apk string, diffwriter io.Writer) error {
	for _, source := range p.Sources {
		srcp := filepath.Join(apk, filepath.Clean(filepath.FromSlash(source)))

		buf, err := os.ReadFile(srcp)
		if err != nil {
			return fmt.Errorf("patch %q: read %q: %w", source, srcp, err)
		}

		// normalize line endings (apktool will emit crlf on windows)
		// note: we don't need to do this in the replacements since go/scanner normalizes raw literals
		buf = bytes.ReplaceAll(buf, []byte{'\r', '\n'}, []byte{'\n'})

		obuf := string(buf)
		sbuf := string(buf)

		for i, x := range p.Patchers {
			out, err := x.PatchString(sbuf)
			if err != nil {
				return fmt.Errorf("patch %q: patcher %d: %w", source, i, err)
			}
			sbuf = out
		}

		diff := gotextdiff.ToUnified(
			"a/"+source, "b/"+source, obuf,
			myers.ComputeEdits(span.URIFromPath(source), obuf, sbuf),
		)
		if _, err := fmt.Fprint(diffwriter, diff); err != nil {
			return fmt.Errorf("patch %q: could not write diff: %w", source, err)
		}

		if err := os.WriteFile(srcp, []byte(sbuf), 0666); err != nil {
			return fmt.Errorf("patch %q: could not write output: %w", source, err)
		}
	}
	return nil
}

type StringPatcher interface {
	PatchString(string) (string, error)
}

type StringPatcherFunc func(string) (string, error)

func (fn StringPatcherFunc) PatchString(s string) (string, error) {
	return fn(s)
}

func ReplaceString(find, replace string) StringPatcher {
	return StringPatcherFunc(func(s string) (string, error) {
		if !strings.Contains(s, find) {
			return s, fmt.Errorf("could not find %q", find)
		}
		return strings.ReplaceAll(s, find, replace), nil
	})
}

func ReplaceStringAppend(find, replace string) StringPatcher {
	return ReplaceString(find, find+replace)
}

func ReplaceStringPrepend(find, replace string) StringPatcher {
	return ReplaceString(find, replace+find)
}

func ReplaceStringRe(find *regexp.Regexp, replace string) StringPatcher {
	return StringPatcherFunc(func(x string) (string, error) {
		if !find.MatchString(x) {
			return x, fmt.Errorf("could not find %q", find.String())
		}
		return find.ReplaceAllString(x, replace), nil
	})
}

func ReplaceStringReLiteral(find *regexp.Regexp, replace string) StringPatcher {
	return StringPatcherFunc(func(x string) (string, error) {
		if !find.MatchString(x) {
			return x, fmt.Errorf("could not find %q", find.String())
		}
		return find.ReplaceAllLiteralString(x, replace), nil
	})
}

func AppendString(s string) StringPatcher {
	return StringPatcherFunc(func(x string) (string, error) {
		return x + s, nil
	})
}

func ReplaceWith(s string) StringPatcher {
	return StringPatcherFunc(func(x string) (string, error) {
		return s, nil
	})
}

func MustContain(s string) StringPatcher {
	return StringPatcherFunc(func(x string) (string, error) {
		if !strings.Contains(x, s) {
			return x, fmt.Errorf("could not find %q", s)
		}
		return x + "\n", nil // hack to be able to use in InMethod
	})
}

func inMethod(method string, pt StringPatcher) StringPatcher {
	return StringPatcherFunc(func(smali string) (string, error) {
		lines := strings.Split(smali, "\n")

		var cmethod string
		var chunk string
		var chunks []string
		for _, l := range lines {
			lf := strings.Fields(l)
			if len(lf) >= 1 && lf[0] == ".method" {
				cmethod = lf[len(lf)-1]
			} else if len(lf) >= 2 && lf[0] == ".end" && lf[1] == "method" {
				cmethod = ""
				if chunk != "" {
					chunks = append(chunks, chunk)
					chunk = ""
				}
			} else if cmethod == method {
				chunk += l + "\n"
			}
		}
		if len(chunks) == 0 {
			return smali, fmt.Errorf("could not find method %q", method)
		}

		var crepl bool
		for _, chunk := range chunks {
			nchunk, err := pt.PatchString(chunk)
			if err != nil {
				return smali, fmt.Errorf("could not run patcher in method %q: %w", method, err)
			}
			ns := strings.Replace(smali, chunk, nchunk, 1)
			if chunk != nchunk {
				if ns == smali {
					panic("chunk but not smali changed")
				}
				crepl = true
			}
			smali = ns
		}
		if !crepl {
			return smali, fmt.Errorf("identical output") // NOTE: this may not be an error in some cases, change this?
		}

		return smali, nil
	})
}

func InMethod(method string, pt ...StringPatcher) StringPatcher {
	return StringPatcherFunc(func(smali string) (string, error) {
		var err error
		for _, x := range pt {
			if smali, err = inMethod(method, x).PatchString(smali); err != nil {
				return smali, err
			}
		}
		return smali, nil
	})
}

func InConstant(constant string, pt StringPatcher) StringPatcher {
	return StringPatcherFunc(func(smali string) (string, error) {
		lines := strings.Split(smali, "\n")

		var found bool
		for n, l := range lines {
			lf := strings.Fields(l)
			if len(lf) >= 1 && lf[0] == ".field" {
				for i := 1; i < len(lf); i++ {
					if lf[i] == "=" && lf[i-1] == constant {
						found = true
						nstr, err := pt.PatchString(l)
						if err != nil {
							return smali, fmt.Errorf("replace constant: %w", err)
						}
						lines[n] = nstr
						break
					}
				}
			}
		}

		if !found {
			return smali, errors.New("replace constant: constant not found")
		}

		smali = strings.Join(lines, "\n")
		return smali, nil
	})
}

func FixIndent(s string) string {
	if s == "" {
		return ""
	}

	// remove the initial newline
	if s[0] != '\n' {
		panic("doesn't start on a new line")
	}
	s = s[1:]

	// find the base indentation
	n := len(s) - len(strings.TrimRight(s, "\t"))
	s = s[:len(s)-n]

	// remove the base indentation, remove extra indentation, and convert remaining tabs to 4 spaces
	var b strings.Builder
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		v := sc.Bytes()
		if len(v) != 0 {
			// check initial indentation
			for i := 0; i < n; i++ {
				if i >= len(v) || v[i] != '\t' {
					panic("not indented")
				}
			}
			// convert additional indentation
			var x int
			for i := n; i < len(v); i++ {
				if v[i] != '\t' {
					break
				}
				x++
			}
			// if line contains non-indentation chars
			if n+x != len(v) {
				for i := 0; i < x; i++ {
					b.WriteString("    ")
				}
				b.Write(v[n+x:])
			}
		}
		b.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	return b.String()
}

func ExecuteTemplate(tmpl string, data any) string {
	var b bytes.Buffer
	if err := template.Must(template.New("").Funcs(template.FuncMap{
		"AddInt": func(i, j int) int { return i + j },
	}).Parse(tmpl)).Execute(&b, data); err != nil {
		panic(err)
	}
	return b.String()
}

type resIDInst struct {
	Path string
	Type string
	Name string
}

func (r *resIDInst) Do(apk string, diffwriter io.Writer) error {
	var nid string
	if err := PatchFile("res/values/public.xml", StringPatcherFunc(func(s string) (string, error) {
		var obj struct {
			XMLName xml.Name `xml:"resources"`
			Public  []struct {
				Type string `xml:"type,attr"`
				Name string `xml:"name,attr"`
				ID   string `xml:"id,attr"`
			} `xml:"public"`
		}
		if err := xml.Unmarshal([]byte(s), &obj); err != nil {
			return "", fmt.Errorf("parse existing resources: %w", err)
		}

		var last uint64
		for _, x := range obj.Public {
			if x.Type == r.Type {
				if x.Name == r.Name {
					return s, nil
				}
				v, err := strconv.ParseUint(x.ID, 0, 32)
				if err != nil {
					return "", fmt.Errorf("parse existing resources: %w", err)
				}
				if v >= last {
					last = v
				}
			}
		}
		if last == 0 {
			return "", fmt.Errorf("no existing resources found with type %q", r.Type)
		}
		nid = "0x" + strconv.FormatUint(last+1, 16)

		return ReplaceStringPrepend(
			"\n</resources>",
			"\n    <public type=\""+r.Type+"\" name=\""+r.Name+"\" id=\""+nid+"\" />",
		).PatchString(s)
	})).Do(apk, diffwriter); err != nil {
		return err
	}
	if nid != "" {
		if err := PatchFile(r.Path+"/R$"+r.Type+".smali", StringPatcherFunc(func(s string) (string, error) {
			return s + "\n\n.field public static final " + r.Name + ":I = " + nid + "\n", nil
		})).Do(apk, diffwriter); err != nil {
			return err
		}
	}
	return nil
}

func DefineR(path, typ, name string) Instruction {
	return &resIDInst{
		Path: path,
		Type: typ,
		Name: name,
	}
}
