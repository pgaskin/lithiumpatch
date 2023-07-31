// Package dictionary adds offline dictionary support.
package dictionary

import (
	_ "embed"
	"encoding/xml"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pgaskin/lithiumpatch/dict"
	. "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"
	_ "github.com/pgaskin/lithiumpatch/patches/internal/public_hcwv_ctx"
)

//go:embed dictionary.js
var dictionaryJS []byte

func init() {
	Register("dictionary",
		Build("assets/dict"),
		WriteFile("assets/js/dictionary.js", dictionaryJS),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceStringAppend(
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/themes.js\'></script>`,
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/dictionary.js\'></script>`,
			),
			ReplaceStringAppend(
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/themes.js\'></script>`,
				`<script type=\'text/javascript\' src=\'file:///android_asset/dict/dict.js\' data-dictionaries=\'{{dicts}}\'></script>`,
			),
			StringPatcherFunc(func(s string) (string, error) {
				var x strings.Builder
				for i, d := range dict.Dicts() {
					if i != 0 {
						x.WriteByte(' ')
					}
					x.WriteString(html.EscapeString(d))
				}
				return strings.ReplaceAll(s, "{{dicts}}", x.String()), nil
			}),
		),
		PatchFile("res/values/arrays.xml",
			ReplaceStringPrepend(
				FixIndent("\n"+`</resources>`),
				FixIndent("\n"+`
					<string-array name="dict_names">
						{{dicts}}
					</string-array>
				`),
			),
			StringPatcherFunc(func(s string) (string, error) {
				var x strings.Builder
				for i, d := range dict.Dicts() {
					if i != 0 {
						x.WriteString("\n        ")
					}
					x.WriteString("<item>")
					xml.EscapeText(&x, []byte(d))
					x.WriteString("</item>")
				}
				return strings.ReplaceAll(s, "{{dicts}}", x.String()), nil
			}),
		),
		DefineR("smali/com/faultexception/reader", "array", "dict_names"),
		PatchFile("res/xml/preferences.xml",
			ReplaceStringPrepend(
				FixIndent("\n"+`
					<PreferenceCategory android:title="@string/pref_category_advanced">
				`),
				FixIndent("\n"+`
					<PreferenceCategory android:title="Dictionary">
						<MultiSelectListPreference android:title="Disable dictionaries" android:key="dict_disabled" android:entries="@array/dict_names" android:entryValues="@array/dict_names" />
						<SwitchPreferenceCompat android:title="Show examples" android:key="dict_show_examples" android:defaultValue="true" />
						<SwitchPreferenceCompat android:title="Show word info" android:key="dict_show_info" android:defaultValue="true" />
					</PreferenceCategory>
				`),
			),
		),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView$JsInterface.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBookReady()V
				`),
				FixIndent("\n"+`
				.method public getDictDisabled()Ljava/lang/String;
					.locals 3

					.annotation runtime Landroid/webkit/JavascriptInterface;
					.end annotation

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView$JsInterface;->this$0:Lcom/faultexception/reader/content/HtmlContentWebView;
					iget-object v0, v0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContext:Landroid/content/Context;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "dict_disabled"
					const v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getStringSet(Ljava/lang/String;Ljava/util/Set;)Ljava/util/Set;
					move-result-object v1

					const-string v0, ""
					if-eqz v1, :done
					const-string v0, " "
					invoke-static {v0, v1}, Ljava/lang/String;->join(Ljava/lang/CharSequence;Ljava/lang/Iterable;)Ljava/lang/String;
					move-result-object v0

					:done
					return-object v0
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBookReady()V
				`),
				FixIndent("\n"+`
				.method public getDictShowExamples()Z
					.locals 3

					.annotation runtime Landroid/webkit/JavascriptInterface;
					.end annotation

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView$JsInterface;->this$0:Lcom/faultexception/reader/content/HtmlContentWebView;
					iget-object v0, v0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContext:Landroid/content/Context;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "dict_show_examples"
					const v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v0

					return v0
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBookReady()V
				`),
				FixIndent("\n"+`
				.method public getDictShowInfo()Z
					.locals 3

					.annotation runtime Landroid/webkit/JavascriptInterface;
					.end annotation

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView$JsInterface;->this$0:Lcom/faultexception/reader/content/HtmlContentWebView;
					iget-object v0, v0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContext:Landroid/content/Context;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "dict_show_info"
					const v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v0

					return v0
				.end method
				`),
			),
		),
	)
}

type Build string

func (b Build) Do(apk string, diffwriter io.Writer) error {
	p := filepath.Join(apk, filepath.Clean(filepath.FromSlash(string(b))))
	if err := os.Mkdir(p, 0777); err != nil {
		return err
	}
	if err := dict.Build(p); err != nil {
		return err
	}
	return WriteFile(filepath.Join(string(b), "dict.js"), dict.JS()).Do(apk, diffwriter)
}
