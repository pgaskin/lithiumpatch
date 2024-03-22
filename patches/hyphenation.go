// # Hyphenation
//
// Add an option to enable/disable hyphenation. Previously, hyphenation would be
// disabled unless the book contained styles to enable it (which is not common).
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("hyphenation",
		PatchFile("assets/js/epub.js",
			ReplaceStringPrepend(
				"\n"+`    var textAlign = void 0;`,
				"\n"+`    var hyphenation = void 0;`,
			),
			ReplaceStringPrepend(
				"\n"+`        setTextAlign: setTextAlign`,
				"\n"+`        setHyphenation: setHyphenation,`,
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
					function setTextAlign(align) {
						textAlign = align;
						updateStyleElement();
						reflowIfNecessary();
					}
				`),
				FixIndent("\n"+`
					function setHyphenation(hyp) {
						hyphenation = hyp;
						updateStyleElement();
						reflowIfNecessary();
					}
				`),
			),
			ReplaceStringPrepend(
				"\n"+`        styleElement.innerText = specificitySelector + ' * { ' + style + ' }';`,
				"\n"+`        style += hyphenation ? '-webkit-hyphens: auto; -webkit-hyphenate-limit-chars: 6 3 3; -webkit-hyphenate-limit-last: always; hyphens: auto; hyphenate-limit-chars: 6 3 3; hyphenate-limit-last: always; hyphenate-limit-zone: 8%; hyphenate-limit-lines: 2;' : '-webkit-hyphens: none; hyphens: none;';`,
			),
		),
		PatchFiles(
			[]string{
				"smali/com/faultexception/reader/content/BookView.smali",
				"smali/com/faultexception/reader/content/ContentView.smali",
			},
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public setTextAlign(I)V
					.locals 0

					return-void
				.end method
				`),
				FixIndent("\n"+`
				.method public setHyphenation(Z)V
					.locals 0

					return-void
				.end method
				`),
			),
		),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentView.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public setTextAlign(I)V
					.locals 1

					.line 126
					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentView;->mContentWebView:Lcom/faultexception/reader/content/HtmlContentWebView;

					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->setTextAlign(I)V

					return-void
				.end method
				`),
				FixIndent("\n"+`
				.method public setHyphenation(Z)V
					.locals 1

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentView;->mContentWebView:Lcom/faultexception/reader/content/HtmlContentWebView;

					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->setHyphenation(Z)V

					return-void
				.end method
				`),
			),
		),
		PatchFile("smali/com/faultexception/reader/content/EPubBookView.smali",
			ReplaceStringPrepend(
				"\n"+`.field private mTextAlign:I`,
				"\n"+`.field private mHyphenation:Z`,
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public setTextAlign(I)V
					.locals 1

					.line 332
					iput p1, p0, Lcom/faultexception/reader/content/EPubBookView;->mTextAlign:I

					.line 333
					iget-object v0, p0, Lcom/faultexception/reader/content/EPubBookView;->mContentView:Lcom/faultexception/reader/content/ContentView;

					if-eqz v0, :cond_0

					.line 334
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/ContentView;->setTextAlign(I)V

					:cond_0
					return-void
				.end method
				`),
				FixIndent("\n"+`
				.method public setHyphenation(Z)V
					.locals 1
					iput-boolean p1, p0, Lcom/faultexception/reader/content/EPubBookView;->mHyphenation:Z
					iget-object v0, p0, Lcom/faultexception/reader/content/EPubBookView;->mContentView:Lcom/faultexception/reader/content/ContentView;
					if-eqz v0, :cond_0
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/ContentView;->setHyphenation(Z)V
					:cond_0
					return-void
				.end method
				`),
			),
		),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceStringPrepend(
				"\n"+`.field private mTextAlign:I`,
				"\n"+`.field private mHyphenation:Z`,
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public setTextAlign(I)V
					.locals 2

					.line 812
					iput p1, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mTextAlign:I

					.line 813
					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mUrl:Ljava/lang/String;

					if-eqz v0, :cond_0

					iget-boolean v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mDisplaySettingsInjected:Z

					if-eqz v0, :cond_0

					.line 814
					new-instance v0, Ljava/lang/StringBuilder;

					invoke-direct {v0}, Ljava/lang/StringBuilder;-><init>()V

					const-string v1, "LithiumJs.setTextAlign("

					invoke-virtual {v0, v1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					invoke-virtual {v0, p1}, Ljava/lang/StringBuilder;->append(I)Ljava/lang/StringBuilder;

					const-string p1, ")"

					invoke-virtual {v0, p1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					invoke-virtual {v0}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;

					move-result-object p1

					invoke-virtual {p0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->executeJavascript(Ljava/lang/String;)V

					:cond_0
					return-void
				.end method
				`),
				FixIndent("\n"+`
				.method public setHyphenation(Z)V
					.locals 2
					iput-boolean p1, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mHyphenation:Z
					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mUrl:Ljava/lang/String;
					if-eqz v0, :cond_0
					iget-boolean v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mDisplaySettingsInjected:Z
					if-eqz v0, :cond_0
					new-instance v0, Ljava/lang/StringBuilder;
					invoke-direct {v0}, Ljava/lang/StringBuilder;-><init>()V
					const-string v1, "LithiumJs.setHyphenation("
					invoke-virtual {v0, v1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v0, p1}, Ljava/lang/StringBuilder;->append(Z)Ljava/lang/StringBuilder;
					const-string p1, ")"
					invoke-virtual {v0, p1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v0}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;
					move-result-object p1
					invoke-virtual {p0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->executeJavascript(Ljava/lang/String;)V
					:cond_0
					return-void
				.end method
				`),
			),
			InMethod("prepareContentStream(Ljava/io/InputStream;)Ljava/io/InputStream;",
				ReplaceStringPrepend(
					FixIndent("\n"+`
						const-string v3, ");   LithiumJs.setTextAlign("

						invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

						iget v3, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mTextAlign:I

						invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(I)Ljava/lang/StringBuilder;
					`),
					FixIndent("\n"+`
						const-string v3, ");   LithiumJs.setHyphenation("
						invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
						iget-boolean v3, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mHyphenation:Z
						invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Z)Ljava/lang/StringBuilder;
					`),
				),
			),
		),
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("updateFeaturesForBookView()V",
				ReplaceStringPrepend(
					FixIndent("\n"+`
						.line 619
						iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mBookView:Lcom/faultexception/reader/content/BookView;

						const/4 v1, 0x1

						invoke-virtual {v0, v1}, Lcom/faultexception/reader/content/BookView;->supportsFeature(I)Z

						move-result v0

						const/4 v1, 0x0

						if-eqz v0, :cond_0

						.line 620
						iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;

						const/16 v2, 0x64

						const-string v3, "textSize"

						invoke-interface {v0, v3, v2}, Landroid/content/SharedPreferences;->getInt(Ljava/lang/String;I)I

						move-result v0

						.line 621
						iget-object v2, p0, Lcom/faultexception/reader/ReaderActivity;->mBookView:Lcom/faultexception/reader/content/BookView;

						invoke-virtual {v2, v0}, Lcom/faultexception/reader/content/BookView;->setTextSize(I)V
					`),
					FixIndent("\n"+`
						iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;
						const/16 v2, 0x1
						const-string v3, "hyphenation"
						invoke-interface {v0, v3, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
						move-result v0
						iget-object v2, p0, Lcom/faultexception/reader/ReaderActivity;->mBookView:Lcom/faultexception/reader/content/BookView;
						invoke-virtual {v2, v0}, Lcom/faultexception/reader/content/BookView;->setHyphenation(Z)V
					`),
				),
			),
		),
		// note: we don't need the onTextAlignChanged stuff in ReaderActivity$7 since we can only change this from settings, not the display popup
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Use hyphenation" android:key="hyphenation" android:defaultValue="true" />`,
			),
		),
	)
}
