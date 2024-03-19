package dispsettingsexpand

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

func init() {
	Register("dispsettingsexpand",
		PatchFile("res/layout/fragment_display_settings.xml",
			ReplaceString(
				`com.faultexception.reader.widget.ExpansionScrollView`,
				`ScrollView`,
			),
			ReplaceString(
				`<Space android:id="@id/end_padding" android:visibility="gone"`,
				`<Space android:id="@id/end_padding" android:visibility="invisible"`,
			),
			ReplaceString(
				`<ImageButton android:id="@id/expand_more"`,
				`<ImageButton android:id="@id/expand_more" android:visibility="gone"`,
			),
			ReplaceString(
				`<LinearLayout android:orientation="vertical" android:id="@id/more_section" android:paddingBottom="8.0dip" android:visibility="gone"`,
				`<LinearLayout android:orientation="vertical" android:id="@id/more_section" android:paddingBottom="8.0dip" android:visibility="visible"`,
			),
		),
		PatchFile("res/layout-v17/fragment_display_settings.xml",
			ReplaceString(
				`com.faultexception.reader.widget.ExpansionScrollView`,
				`ScrollView`,
			),
			ReplaceString(
				`<Space android:id="@id/end_padding" android:visibility="gone"`,
				`<Space android:id="@id/end_padding" android:visibility="invisible"`,
			),
			ReplaceString(
				`<ImageButton android:id="@id/expand_more"`,
				`<ImageButton android:id="@id/expand_more" android:visibility="gone"`,
			),
			ReplaceString(
				`<LinearLayout android:orientation="vertical" android:id="@id/more_section" android:paddingBottom="8.0dip" android:visibility="gone"`,
				`<LinearLayout android:orientation="vertical" android:id="@id/more_section" android:paddingBottom="8.0dip" android:visibility="visible"`,
			),
		),
		PatchFile(`smali/com/faultexception/reader/DisplaySettingsFragment.smali`,
			ReplaceString(
				`Lcom/faultexception/reader/widget/ExpansionScrollView;->findViewById(I)Landroid/view/View`,
				`Landroid/widget/ScrollView;->findViewById(I)Landroid/view/View`,
			),
			InMethod("onCreateView(Landroid/view/LayoutInflater;Landroid/view/ViewGroup;Landroid/os/Bundle;)Landroid/view/View;",
				ReplaceString(
					`check-cast p2, Lcom/faultexception/reader/widget/ExpansionScrollView;`,
					`check-cast p2, Landroid/widget/ScrollView;`,
				),
				ReplaceString(`new-instance v0, Lcom/faultexception/reader/DisplaySettingsFragment$1;`+"\n\n", ``),
				ReplaceString(`invoke-direct {v0, p0, p2}, Lcom/faultexception/reader/DisplaySettingsFragment$1;-><init>(Lcom/faultexception/reader/DisplaySettingsFragment;Lcom/faultexception/reader/widget/ExpansionScrollView;)V`+"\n\n", ``),
				ReplaceString(`invoke-virtual {p1, v0}, Landroid/widget/ImageButton;->setOnClickListener(Landroid/view/View$OnClickListener;)V`+"\n\n", ``),
				ReplaceString(`invoke-virtual {p2, p1, v0}, Lcom/faultexception/reader/widget/ExpansionScrollView;->setExpandButtonAndContainer(Landroid/view/View;Landroid/view/View;)V`+"\n\n", ``),
			),
		),
		// also see: ExpansionScrollView, DisplaySettingsFragment#updateExpandButton, DisplaySettingsFragment#onCreate
	)
}
