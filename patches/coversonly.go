// # Cover-only grid
//
// Add an option to only show covers on the grid.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("coversonly",
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Hide footer (grid view)" android:key="only_covers" android:defaultValue="false" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/BooksAdapter.smali",
			InMethod("onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V",
				// note: grid is mLayoutMode=0 is the grid, :cond_5 is for the list view in dark mode
				MustContain(FixIndent("\n"+`
					iget v3, p0, Lcom/faultexception/reader/BooksAdapter;->mLayoutMode:I

					if-nez v3, :cond_5
				`)),
				MustContain(FixIndent("\n"+`
					:cond_5
					iget-boolean v2, p0, Lcom/faultexception/reader/BooksAdapter;->mDarkMode:Z
				`)),
				// note: only the grid branch currently has the trailing goto
				ReplaceString(
					FixIndent("\n"+`
						iget-object v2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->footerView:Landroid/view/View;

						invoke-virtual {v2, v0}, Landroid/view/View;->setBackgroundColor(I)V

						goto :goto_3
					`),
					FixIndent("\n"+`
						iget-object v2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->footerView:Landroid/view/View;

						invoke-virtual {v2, v0}, Landroid/view/View;->setBackgroundColor(I)V

						invoke-direct {p0, v2}, Lcom/faultexception/reader/BooksAdapter;->maybeHideFooter(Landroid/view/View;)V

						goto :goto_3
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V
				`),
				FixIndent("\n"+`
				.method private maybeHideFooter(Landroid/view/View;)V
					.locals 3

					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mActivity:Landroidx/appcompat/app/AppCompatActivity;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "only_covers"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v1

					const/16 v2, 0x0 # android.View.VISIBLE
					if-eqz v1, :visible
					const/16 v2, 0x8 # android.View.GONE
					:visible
					invoke-virtual {p1, v2}, Landroid/view/View;->setVisibility(I)V

					return-void
				.end method
				`),
			),
		),
	)
}
