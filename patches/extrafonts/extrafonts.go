// Package extrafonts adds additional fonts from the res directory.
package extrafonts

import (
	"io"

	"github.com/pgaskin/lithiumpatch/fonts"
	. "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"
)

func init() {
	Register("extrafonts", extrafonts{})
}

type extrafonts struct{}

func (extrafonts) Do(apk string, diffwriter io.Writer) error {
	xs := fonts.All()
	if len(xs) == 0 {
		return nil
	}
	var pt []Instruction
	for _, f := range xs {
		if f.Regular != nil {
			pt = append(pt, WriteFile("assets/fonts/"+f.Base+"-Regular.ttf", f.Regular))
		}
		if f.Bold != nil {
			pt = append(pt, WriteFile("assets/fonts/"+f.Base+"-Bold.ttf", f.Bold))
		}
		if f.Italic != nil {
			pt = append(pt, WriteFile("assets/fonts/"+f.Base+"-Italic.ttf", f.Italic))
		}
		if f.BoldItalic != nil {
			pt = append(pt, WriteFile("assets/fonts/"+f.Base+"-BoldItalic.ttf", f.BoldItalic))
		}
	}
	pt = append(pt, PatchFile("smali/com/faultexception/reader/fonts/Fonts.smali",
		InMethod("<clinit>()V",
			ReplaceStringAppend(
				FixIndent("\n"+`
					invoke-static {v0}, Ljava/util/Arrays;->asList([Ljava/lang/Object;)Ljava/util/List;

					move-result-object v0
				`),
				FixIndent("\n"+`
					new-instance v1, Ljava/util/ArrayList;
					invoke-direct {v1, v0}, Ljava/util/ArrayList;-><init>(Ljava/util/Collection;)V
					move-object v0, v1

					invoke-static {v0}, Lcom/faultexception/reader/fonts/Fonts;->initCustomFonts(Ljava/util/ArrayList;)V
				`),
			),
		),
		AppendString(
			FixIndent(ExecuteTemplate("\n"+`
			.method private static initCustomFonts(Ljava/util/ArrayList;)V
				.locals 7
				{{range .}}
				const-string v1, "{{.Name}}"
				{{if .Regular -}}
				const-string v2, "{{.Base}}-Regular.ttf"
				{{- else -}}
				const/4 v2, 0x0
				{{- end}}
				{{if .Bold -}}
				const-string v3, "{{.Base}}-Bold.ttf"
				{{- else -}}
				const/4 v3, 0x0
				{{- end}}
				{{if .Italic -}}
				const-string v4, "{{.Base}}-Italic.ttf"
				{{- else -}}
				const/4 v4, 0x0
				{{- end}}
				{{if .BoldItalic -}}
				const-string v5, "{{.Base}}-BoldItalic.ttf"
				{{- else -}}
				const/4 v5, 0x0
				{{- end}}
				const/4 v6, {{.Script.Flags | printf "%#x"}} # {{.Script}}

				new-instance v0, Lcom/faultexception/reader/fonts/Font;
				invoke-direct/range {v0 .. v6}, Lcom/faultexception/reader/fonts/Font;-><init>(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;I)V
				invoke-virtual {p0, v0}, Ljava/util/ArrayList;->add(Ljava/lang/Object;)Z
				{{end}}
				return-void
			.end method
			`, xs)),
		),
	))
	for _, x := range pt {
		if err := x.Do(apk, diffwriter); err != nil {
			return err
		}
	}
	return nil
}
