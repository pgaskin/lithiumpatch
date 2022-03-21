// Package dictionary adds dictionary support using github.com/pgaskin/dictserver.
package dictionary

import (
	_ "embed"

	. "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"
	_ "github.com/pgaskin/lithiumpatch/patches/internal/public_hcwv_ctx"
)

//go:embed dictionary.js
var dictionaryJS []byte

const dictserver = "https://dict.api.pgaskin.net"

func init() {
	Register("dictionary",
		WriteFile("assets/js/dictionary.js", dictionaryJS),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceStringAppend(
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/themes.js\'></script>`,
				`<script type=\'text/javascript\' src=\'file:///android_asset/js/dictionary.js\'></script>`,
			),
		),
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <EditTextPreference android:title="Dictionary server URL" android:key="dictserver" android:defaultValue="`+dictserver+`" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView$JsInterface.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBookReady()V
				`),
				FixIndent("\n"+`
				.method public getDictionaryURL()Ljava/lang/String;
					.locals 3

					.annotation runtime Landroid/webkit/JavascriptInterface;
					.end annotation

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView$JsInterface;->this$0:Lcom/faultexception/reader/content/HtmlContentWebView;
					iget-object v0, v0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContext:Landroid/content/Context;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "dictserver"
					const-string v2, "`+dictserver+`"
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getString(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v0

					return-object v0
				.end method
				`),
			),
		),
	)
}
