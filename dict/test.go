//go:build ignore

package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
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
<div style="margin: 80vh 0" contenteditable>test test test word word word sample sample sample</div>
<script>
globalThis.LithiumThemes = {
    set: function() {},
}
document.addEventListener("DOMContentLoaded", () => {
    globalThis.LithiumThemes.set({
        backgroundColor: 16113331,
        bgIsDark: false,
        builtin: false,
        darkChrome: true,
        linkColor: 867715,
        name: "test",
        textColor: 1118481,
    })
})
</script>
<script src="dictionary.js" data-dicts="{{range $i, $x := .}}{{if $i}} {{end}}{{$x}}{{end}}"></script>
<script type="module">
import dict from "./dict.js"
{{- range . }}
globalThis["{{.}}"] = await dict("{{.}}")
{{- end }}
console.log("loaded dictionaries")
</script>
`))

//go:embed lib/*.java
var javaFS embed.FS

const javaMain = `
import java.io.File;
import java.nio.file.FileSystems;
import java.nio.file.Path;
import net.pgaskin.dictionary.Dictionary;

public class Main {
	public static void main(String[] args) {
        final Path self;
        try {
            self = FileSystems.getDefault().getPath(new File(Main.class.getProtectionDomain().getCodeSource().getLocation().toURI()).getParentFile().getPath());
        } catch (Exception ex) {
            throw new RuntimeException(ex);
        }
        for (int i = 1; i < args.length; i++) {
            System.out.print(Dictionary.load(Dictionary.FS.local(self.resolve(args[i]))).query(args[0]));
        }
	}
}
`

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
	if javac, err := exec.LookPath("javac"); err == nil {
		if jar, err := exec.LookPath("jar"); err == nil {
			if err := func() error {
				td, err := os.MkdirTemp("", "")
				if err != nil {
					return fmt.Errorf("make temp dir: %w", err)
				}
				defer os.RemoveAll(td)

				if err := os.MkdirAll(filepath.Join(td, "net", "pgaskin", "dictionary"), 0755); err != nil {
					return fmt.Errorf("make package dir: %w", err)
				}

				var java []string
				if err := fs.WalkDir(javaFS, ".", func(path string, d fs.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return err
					} else if buf, err := fs.ReadFile(javaFS, path); err != nil {
						return err
					} else if err := os.WriteFile(filepath.Join(td, "net", "pgaskin", "dictionary", d.Name()), buf, 0666); err != nil {
						return err
					} else {
						java = append(java, filepath.Join(td, "net", "pgaskin", "dictionary", d.Name()))
					}
					return nil
				}); err != nil {
					return fmt.Errorf("copy files: %w", err)
				}
				if err := os.WriteFile(filepath.Join(td, "Main.java"), []byte(javaMain), 0666); err != nil {
					return fmt.Errorf("write main: %w", err)
				} else {
					java = append(java, "Main.java")
				}

				var javacOut bytes.Buffer
				javacCmd := exec.Command(javac)
				javacCmd.Args[0] = "javac"
				javacCmd.Args = append(javacCmd.Args, java...)
				javacCmd.Dir = td
				javacCmd.Stdout = &javacOut
				javacCmd.Stderr = &javacOut
				if err := javacCmd.Run(); err != nil {
					return fmt.Errorf("javac %q: %w (output=%#v)", javacCmd.Args[1:], err, javacOut.String())
				}

				var jarOut bytes.Buffer
				jarCmd := exec.Command(jar, "cfe", "build/dict.jar", "Main", "-C", td, ".")
				jarCmd.Args[0] = "jar"
				jarCmd.Stdout = &jarOut
				jarCmd.Stderr = &jarOut
				if err := jarCmd.Run(); err != nil {
					return fmt.Errorf("jar %q: %w (output=%#v)", jarCmd.Args[1:], err, jarOut.String())
				}

				return nil
			}(); err != nil {
				panic(fmt.Errorf("java: %w", err))
			}
		}
	}
	if _, err := os.Stat("go.mod"); err == nil {
		_ = os.Symlink("../dict/lib/dict.js", "build/dict.js")
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
