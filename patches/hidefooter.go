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
				"\n"+`        <SwitchPreferenceCompat android:title="Hide footer slider (reader)" android:key="hide_reader_footer" android:defaultValue="false" />`,
			),
		),

		// Hide only the page slider when enabled. Do this once on activity start.
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			// Add helper method before onCreate
			ReplaceStringPrepend(
				"\n"+`.method protected onCreate(Landroid/os/Bundle;)V`,
				FixIndent("\n"+`
                .method private applyHideFooterSlider()V
                    .locals 3
                    const/4 v2, 0x0
                    iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;
                    const-string v1, "hide_reader_footer"
                    invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
                    move-result v2
                    if-eqz v2, :lith_patch_hfs_show
                    iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mPageSeekView:Landroid/widget/SeekBar;
                    const/16 v1, 0x8
                    invoke-virtual {v0, v1}, Landroid/view/View;->setVisibility(I)V
                    return-void
                    :lith_patch_hfs_show
                    iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mPageSeekView:Landroid/widget/SeekBar;
                    const/4 v1, 0x0
                    invoke-virtual {v0, v1}, Landroid/view/View;->setVisibility(I)V
                    return-void
                .end method
                `),
			),
			// Call helper after SeekBar listener is set (robust minimal anchor)
			InMethod("onCreate(Landroid/os/Bundle;)V",
				ReplaceStringAppend(
					FixIndent("\n"+`
						.line 373
						invoke-virtual {v2, v0}, Landroid/widget/SeekBar;->setOnSeekBarChangeListener(Landroid/widget/SeekBar$OnSeekBarChangeListener;)V
					`),
					FixIndent("\n"+`
						invoke-direct {v0}, Lcom/faultexception/reader/ReaderActivity;->applyHideFooterSlider()V
					`),
				),
			),

			// Also apply on resume so changes from Settings take effect immediately
			InMethod("onResume()V",
				ReplaceStringAppend(
					"\n"+`    invoke-super {p0}, Lcom/faultexception/reader/BaseActivity;->onResume()V`,
					FixIndent("\n"+`
                        invoke-direct {p0}, Lcom/faultexception/reader/ReaderActivity;->applyHideFooterSlider()V
                    `),
				),
			),
		),
	)
}
