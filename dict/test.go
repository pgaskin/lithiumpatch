//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"runtime"

	"github.com/pgaskin/lithiumpatch/dict"
	_ "github.com/pgaskin/lithiumpatch/dict/edgedict"
	_ "github.com/pgaskin/lithiumpatch/dict/webster1913"

	_ "github.com/ncruces/go-sqlite3/embed"
)

var tmpl = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title></title>
</head>
<body>
<div contenteditable>test</div>
<script src="dict.js" data-dictionaries="{{range $i, $x := .}}{{if $i}} {{end}}{{$x}}{{end}}"></script>
<script src="dictionary.js"></script>
<script>
(async function() {
{{- range . }}
globalThis["{{.}}"] = await dict("{{.}}")
{{- end }}
console.log("loaded dictionaries")
})()
</script>
`))

func main() {
	var b bytes.Buffer
	if err := tmpl.Execute(&b, dict.Dicts()); err != nil {
		panic(err)
	}
	if err := dict.Parse(true); err != nil {
		panic(err)
	}
	if err := os.RemoveAll("build"); err != nil {
		panic(err)
	}
	if err := os.Mkdir("build", 0777); err != nil {
		panic(err)
	}
	if err := dict.Build("build"); err != nil {
		panic(err)
	}
	if _, err := os.Stat("go.mod"); err == nil {
		_ = os.Symlink("../dict/dict.js", "build/dict.js")
		_ = os.Symlink("../patches/dictionary/dictionary.js", "build/dictionary.js")
	} else if err := os.WriteFile("build/dict.js", dict.JS(), 0666); err != nil {
		panic(err)
	}
	if err := os.WriteFile("build/index.html", b.Bytes(), 0666); err != nil {
		panic(err)
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}