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
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			InMethod("getResponseForUrl(Ljava/lang/String;)Landroid/webkit/WebResourceResponse;",
				ReplaceStringAppend(
					FixIndent("\n"+`
						invoke-static {p1}, Landroid/net/Uri;->parse(Ljava/lang/String;)Landroid/net/Uri;

						move-result-object v0
					`),
					FixIndent("\n"+`
						invoke-direct {p0, v0}, Lcom/faultexception/reader/content/HtmlContentWebView;->getResponseForDictUrl(Landroid/net/Uri;)Landroid/webkit/WebResourceResponse;
						move-result-object v1
						if-eqz v1, :not_dict
						return-object v1
						:not_dict
					`),
				),
				MustContain(FixIndent("\n"+`
					invoke-virtual {v0}, Landroid/net/Uri;->getScheme()Ljava/lang/String;

					move-result-object v1
				`)),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private getResponseForUrl(Ljava/lang/String;)Landroid/webkit/WebResourceResponse;
				`),
				FixIndent("\n"+`
				.method private getResponseForDictUrl(Landroid/net/Uri;)Landroid/webkit/WebResourceResponse;
					.locals 5

					const-string v1, "dict.androidplatform.net"
					invoke-virtual {p1}, Landroid/net/Uri;->getHost()Ljava/lang/String;
					move-result-object v0
					invoke-virtual {v0, v1}, Ljava/lang/String;->equalsIgnoreCase(Ljava/lang/String;)Z
					move-result v0
					if-eqz v0, :not_dict

					const-string v1, "/"
					invoke-virtual {p1}, Landroid/net/Uri;->getPath()Ljava/lang/String;
					move-result-object v4
					invoke-virtual {v4, v1}, Ljava/lang/String;->startsWith(Ljava/lang/String;)Z
					move-result v0
					if-eqz v0, :path_cleaned
					const v1, 0x1
					invoke-virtual {v4, v1}, Ljava/lang/String;->substring(I)Ljava/lang/String;
					move-result-object v4
					:path_cleaned

					const-string v1, "dict/"
					invoke-virtual {v1, v4}, Ljava/lang/String;->concat(Ljava/lang/String;)Ljava/lang/String;
					move-result-object v4

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContext:Landroid/content/Context;
					invoke-virtual {v0}, Landroid/content/Context;->getAssets()Landroid/content/res/AssetManager;
					move-result-object v0

					:t1s
					invoke-virtual {v0, v4}, Landroid/content/res/AssetManager;->open(Ljava/lang/String;)Ljava/io/InputStream;
					move-result-object v0
					:t1e

					.catch Ljava/io/FileNotFoundException; {:t1s .. :t1e} :t1c1
					.catch Ljava/io/IOException; {:t1s .. :t1e} :t1c2
					goto :t1f

					:t1c1
					const p1, 0x0
					new-array p1, p1, [B
					new-instance v0, Ljava/io/ByteArrayInputStream;
					invoke-direct {v0, p1}, Ljava/io/ByteArrayInputStream;-><init>([B)V
					new-instance p1, Landroid/webkit/WebResourceResponse;
					const-string v2, "text/plain"
					const-string v1, "UTF-8"
					invoke-direct {p1, v2, v1, v0}, Landroid/webkit/WebResourceResponse;-><init>(Ljava/lang/String;Ljava/lang/String;Ljava/io/InputStream;)V
					const v0, 404
					const-string v1, "Not Found"
					invoke-virtual {p1, v0, v1}, Landroid/webkit/WebResourceResponse;->setStatusCodeAndReasonPhrase(ILjava/lang/String;)V
					goto :cors

					:t1c2
					const p1, 0x0
					new-array p1, p1, [B
					new-instance v0, Ljava/io/ByteArrayInputStream;
					invoke-direct {v0, p1}, Ljava/io/ByteArrayInputStream;-><init>([B)V
					new-instance p1, Landroid/webkit/WebResourceResponse;
					const-string v2, "text/plain"
					const-string v1, "UTF-8"
					invoke-direct {p1, v2, v1, v0}, Landroid/webkit/WebResourceResponse;-><init>(Ljava/lang/String;Ljava/lang/String;Ljava/io/InputStream;)V
					const v0, 500
					const-string v1, "Internal Server Error"
					invoke-virtual {p1, v0, v1}, Landroid/webkit/WebResourceResponse;->setStatusCodeAndReasonPhrase(ILjava/lang/String;)V
					goto :cors

					:t1f
					new-instance p1, Landroid/webkit/WebResourceResponse;
					const-string v1, ".js"
					invoke-virtual {v4, v1}, Ljava/lang/String;->endsWith(Ljava/lang/String;)Z
					move-result v1
					if-nez v1, :mt1
					const-string v2, "application/octet-stream"
					const v1, 0x0
					goto :mt
					:mt1
					const-string v2, "application/javascript"
					const-string v1, "UTF-8"
					:mt
					invoke-direct {p1, v2, v1, v0}, Landroid/webkit/WebResourceResponse;-><init>(Ljava/lang/String;Ljava/lang/String;Ljava/io/InputStream;)V
					goto :cors

					:cors
					new-instance v0, Ljava/util/HashMap;
					const-string v1, "Access-Control-Allow-Origin"
					const-string v2, "*"
					invoke-direct {v0}, Ljava/util/HashMap;-><init>()V
					invoke-virtual {v0, v1, v2}, Ljava/util/HashMap;->put(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;
					invoke-virtual {p1, v0}, Landroid/webkit/WebResourceResponse;->setResponseHeaders(Ljava/util/Map;)V
					return-object p1

					:not_dict
					const p1, 0x0
					return-object p1
				.end method
				`),
			),
		),
		WriteFile("assets/js/dictionary.js", dictionaryJS),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceStringAppend(
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/themes.js\'></script>`,
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/dictionary.js\' data-dict=\'https://dict.androidplatform.net/dict.js\' data-dicts=\'{{dicts}}\'></script>`,
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
