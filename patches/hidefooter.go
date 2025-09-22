// # Hide reader footer bar
//
// Adds a preference to hide the bottom chapter/footer bar in the reader.
// When enabled, tapping to show chrome will not show the footer bar.
package patches

import (
	. "github.com/pgaskin/lithiumpatch/patches/patchdef"
)

func init() {
	Register("hidefooter",
		// Add toggle in Settings → Advanced
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Hide footer bar (reader)" android:key="hide_reader_footer" android:defaultValue="false" />`,
			),
		),

		// In ReaderActivity#setChromeVisible(boolean), override only the nav bar behavior
		// so the toolbar still follows chrome visibility, but the footer bar stays hidden
		// if the preference is enabled.
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("setChromeVisible(Z)V",
				// increase locals to have spare temp registers v6,v7,v8
				ReplaceString(".locals 6", ".locals 10"),
				// Force nav bar translationY path to hidden on show when enabled
				ReplaceStringPrepend(
					"    if-eqz p1, :cond_3\n",
					"    # lithiumpatch: hide footer bar on show when enabled (translationY)\n"+
						"    const/4 v8, 0x0\n"+
						"    iget-object v7, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;\n"+
						"    const-string v6, \"hide_reader_footer\"\n"+
						"    invoke-interface {v7, v6, v8}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z\n"+
						"    move-result v8\n"+
						"    if-eqz v8, :lith_patch_hf_tyalive\n"+
						"    goto :cond_3\n"+
						"    :lith_patch_hf_tyalive\n",
				),
				// Force nav bar listener choice to invisibleAfter on show when enabled
				ReplaceStringPrepend(
					"    if-eqz p1, :cond_4\n",
					"    # lithiumpatch: hide footer bar on show when enabled (listener)\n"+
						"    const/4 v8, 0x0\n"+
						"    iget-object v7, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;\n"+
						"    const-string v6, \"hide_reader_footer\"\n"+
						"    invoke-interface {v7, v6, v8}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z\n"+
						"    move-result v8\n"+
						"    if-eqz v8, :lith_patch_hf_lalive\n"+
						"    goto :cond_4\n"+
						"    :lith_patch_hf_lalive\n",
				),
			),
		),
	)
}
