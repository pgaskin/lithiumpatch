// # Full-bleed reader background
//
// Optionally expand the full-screen reader background into the status bar
// display cutout area.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("fullbleed",
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <SwitchPreferenceCompat android:title="@string/pref_fullscreen_title" android:key="fullscreen" android:defaultValue="true" />`,
				"\n"+`    <SwitchPreferenceCompat android:title="Fullscreen reading full-bleed" android:key="fullscreen_bleed" android:defaultValue="true" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("onCreate(Landroid/os/Bundle;)V",
				// if mFullscreenEnabled and api 28
				MustContain("\n"+`    const/4 v1, 0x1`), // LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES
				MustContain("\n"+`    iput v1, v4, Landroid/view/WindowManager$LayoutParams;->layoutInDisplayCutoutMode:I`),
			),
			InMethod("setTheme(Lcom/faultexception/reader/themes/Theme;)V",
				ReplaceStringAppend(
					FixIndent("\n"+`
						.line 1329
						iget v1, p1, Lcom/faultexception/reader/themes/Theme;->backgroundColor:I

						or-int/2addr v1, v0

						goto :goto_0

						:cond_0
						const/4 v1, -0x1

						.line 1330
						:goto_0
					`),
					FixIndent("\n"+`
						invoke-direct {p0, v1}, Lcom/faultexception/reader/ReaderActivity;->maybeSetDisplayCutoutBackground(I)V
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public setTheme(Lcom/faultexception/reader/themes/Theme;)V
				`),
				FixIndent("\n"+`
				.method private maybeSetDisplayCutoutBackground(I)V
					.locals 3

					# only if fullscreen active
					iget-boolean v0, p0, Lcom/faultexception/reader/ReaderActivity;->mFullscreenEnabled:Z
					if-eqz v0, :end

					# only if fullscreen full-bleed enabled
					invoke-static {p0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0
					const-string v1, "fullscreen_bleed"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v0
					if-eqz v0, :end

					# reader content cutout frame
					sget v0, Lcom/faultexception/reader/R$id;->content_high_cutout_frame:I
					invoke-virtual {p0, v0}, Lcom/faultexception/reader/ReaderActivity;->findViewById(I)Landroid/view/View;
					move-result-object v0
					check-cast v0, Lcom/faultexception/reader/widget/DisplayCutoutFrameLayout;
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/widget/DisplayCutoutFrameLayout;->setInsetCutoutColor(I)V

					# reader content frame
					sget v0, Lcom/faultexception/reader/R$id;->content_high_frame:I
					invoke-virtual {p0, v0}, Lcom/faultexception/reader/ReaderActivity;->findViewById(I)Landroid/view/View;
					move-result-object v0
					check-cast v0, Lcom/faultexception/reader/widget/SystemBarsFrame;
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/widget/SystemBarsFrame;->setSystemBarsBackgroundColor(I)V

					# reader content loading
					iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mBookContainerView:Lcom/faultexception/reader/widget/DisplayCutoutFrameLayout;
					if-eqz v0, :end
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/widget/DisplayCutoutFrameLayout;->setInsetCutoutColor(I)V

					:end
					return-void
				.end method
				`),
			),
		),
		PatchFile("smali/com/faultexception/reader/widget/DisplayCutoutFrameLayout.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public setInsetCutout(I)V
				`),
				FixIndent("\n"+`
				.method public setInsetCutoutColor(I)V
					.locals 1

					iget-boolean v0, p0, Lcom/faultexception/reader/widget/DisplayCutoutFrameLayout;->mPaintCutout:Z
					if-eqz v0, :end

					new-instance v0, Landroid/graphics/drawable/ColorDrawable;
					invoke-direct {v0, p1}, Landroid/graphics/drawable/ColorDrawable;-><init>(I)V
					iput-object v0, p0, Lcom/faultexception/reader/widget/DisplayCutoutFrameLayout;->mColor:Landroid/graphics/drawable/ColorDrawable;

					:end
					return-void
				.end method
				`),
			),
		),
	)
}
