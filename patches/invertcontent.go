// # Content color inversion
//
// Optionally invert the color (preserving the hue) for FXL books.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("invertcontent",
		WriteFileString("res/drawable/ic_image_24dp.xml", FixIndent(`
			<vector xmlns:android="http://schemas.android.com/apk/res/android" android:width="24dp" android:height="24dp" android:viewportWidth="24" android:viewportHeight="24">
				<path android:fillColor="#ff000000" android:pathData="M21,19V5c0,-1.1 -0.9,-2 -2,-2H5c-1.1,0 -2,0.9 -2,2v14c0,1.1 0.9,2 2,2h14c1.1,0 2,-0.9 2,-2zM8.5,13.5l2.5,3.01L14.5,12l4.5,6H5l3.5,-4.5z"/>
			</vector>
		`)),
		DefineR("smali/com/faultexception/reader", "drawable", "ic_image_24dp"),

		WriteFileString("res/drawable/ic_article_24dp.xml", FixIndent(`
			<vector xmlns:android="http://schemas.android.com/apk/res/android" android:width="24dp" android:height="24dp" android:viewportWidth="24" android:viewportHeight="24">
				<path android:fillColor="#ff000000" android:pathData="M19,3L5,3c-1.1,0 -2,0.9 -2,2v14c0,1.1 0.9,2 2,2h14c1.1,0 2,-0.9 2,-2L21,5c0,-1.1 -0.9,-2 -2,-2zM14,17L7,17v-2h7v2zM17,13L7,13v-2h10v2zM17,9L7,9L7,7h10v2z"/>
			</vector>
		`)),
		DefineR("smali/com/faultexception/reader", "drawable", "ic_article_24dp"),

		PatchFiles([]string{
			"res/layout/fragment_display_settings.xml",
			"res/layout-v17/fragment_display_settings.xml",
		}, ReplaceStringAppend(
			FixIndent(`
					<LinearLayout android:gravity="center_vertical" android:orientation="horizontal" android:id="@id/text_align" android:paddingLeft="24.0dip" android:paddingRight="8.0dip" android:layout_width="fill_parent" android:layout_height="wrap_content">
						<LinearLayout android:gravity="center_vertical" android:orientation="vertical" android:layout_width="0.0dip" android:layout_height="wrap_content" android:layout_weight="1.0">
							<TextView android:layout_width="wrap_content" android:layout_height="wrap_content" android:text="@string/display_settings_text_align" style="@style/DisplaySettingsHeader" />
							<TextView android:id="@id/text_align_value" android:layout_width="wrap_content" android:layout_height="wrap_content" style="@style/DisplaySettingsValue" />
						</LinearLayout>
						<ImageButton android:id="@id/text_align_start" android:background="@drawable/action_ripple" android:padding="16.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:src="@drawable/ic_align_start" android:contentDescription="@string/display_settings_text_align_start" app:tint="@color/display_settings_control_color_selector" />
						<ImageButton android:id="@id/text_align_justify" android:background="@drawable/action_ripple" android:padding="16.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:src="@drawable/ic_align_justify" android:contentDescription="@string/display_settings_text_align_justify" app:tint="@color/display_settings_control_color_selector" />
					</LinearLayout>
			`),
			FixIndent(`
					<LinearLayout android:gravity="center_vertical" android:orientation="horizontal" android:id="@id/content_invert" android:paddingLeft="24.0dip" android:paddingRight="8.0dip" android:layout_width="fill_parent" android:layout_height="wrap_content">
						<LinearLayout android:gravity="center_vertical" android:orientation="vertical" android:layout_width="0.0dip" android:layout_height="wrap_content" android:layout_weight="1.0">
							<TextView android:layout_width="wrap_content" android:layout_height="wrap_content" android:text="Invert" style="@style/DisplaySettingsHeader" />
							<TextView android:id="@id/content_invert_value" android:layout_width="wrap_content" android:layout_height="wrap_content" style="@style/DisplaySettingsValue" />
						</LinearLayout>
						<ImageButton android:id="@id/content_invert_image" android:background="@drawable/action_ripple" android:padding="16.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:src="@drawable/ic_image_24dp" android:contentDescription="Images" app:tint="@color/display_settings_control_color_selector" />
						<ImageButton android:id="@id/content_invert_page" android:background="@drawable/action_ripple" android:padding="16.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:src="@drawable/ic_article_24dp" android:contentDescription="Page" app:tint="@color/display_settings_control_color_selector" />
					</LinearLayout>
			`),
		)),
		DefineR("smali/com/faultexception/reader", "id", "content_invert"),
		DefineR("smali/com/faultexception/reader", "id", "content_invert_value"),
		DefineR("smali/com/faultexception/reader", "id", "content_invert_image"),
		DefineR("smali/com/faultexception/reader", "id", "content_invert_page"),
		PatchFile("res/values/ids.xml", ReplaceStringPrepend("\n"+`</resources>`, "\n"+`    <item type="id" name="content_invert" />`)),
		PatchFile("res/values/ids.xml", ReplaceStringPrepend("\n"+`</resources>`, "\n"+`    <item type="id" name="content_invert_value" />`)),
		PatchFile("res/values/ids.xml", ReplaceStringPrepend("\n"+`</resources>`, "\n"+`    <item type="id" name="content_invert_image" />`)),
		PatchFile("res/values/ids.xml", ReplaceStringPrepend("\n"+`</resources>`, "\n"+`    <item type="id" name="content_invert_page" />`)),

		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("updateFeaturesForBookView()V",
				ReplaceStringAppend(
					// margin is always applied
					FixIndent(`
						:goto_0
						iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mBookView:Lcom/faultexception/reader/content/BookView;

						iget-object v2, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;

						invoke-virtual {p0}, Lcom/faultexception/reader/ReaderActivity;->getResources()Landroid/content/res/Resources;

						move-result-object v3

						const v4, 0x7f0a0006

						invoke-virtual {v3, v4}, Landroid/content/res/Resources;->getInteger(I)I

						move-result v3

						const-string v4, "margin"

						invoke-interface {v2, v4, v3}, Landroid/content/SharedPreferences;->getInt(Ljava/lang/String;I)I

						move-result v2

						invoke-virtual {v0, v2}, Lcom/faultexception/reader/content/BookView;->setMargin(I)V
					`),
					FixIndent(`
						iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity;->mBookView:Lcom/faultexception/reader/content/BookView;
						iget-object v2, p0, Lcom/faultexception/reader/ReaderActivity;->mPrefs:Landroid/content/SharedPreferences;
						const-string v3, "None"
						const-string v4, "content_invert"
						invoke-interface {v2, v4, v3}, Landroid/content/SharedPreferences;->getString(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
						move-result-object v2
						invoke-virtual {v0, v2}, Lcom/faultexception/reader/content/BookView;->setContentInvert(Ljava/lang/String;)V
					`),
				),
			),
		),

		PatchFile("smali/com/faultexception/reader/DisplaySettingsFragment.smali",
			ReplaceStringAppend(
				FixIndent(`
				.field private mTextAlignJustifyButton:Landroid/widget/ImageButton;

				.field private mTextAlignStartButton:Landroid/widget/ImageButton;

				.field private mTextAlignValueView:Landroid/widget/TextView;

				.field private mTextAlignView:Landroid/view/View;
				`),
				FixIndent(`
				.field private mContentInvertImageButton:Landroid/widget/ImageButton;
				.field private mContentInvertPageButton:Landroid/widget/ImageButton;
				.field private mContentInvertValueView:Landroid/widget/TextView;
				.field private mContentInvertView:Landroid/view/View;
				`),
			),
			ReplaceStringPrepend(
				FixIndent(`
				.method private updateTextAlign()V
				`),
				FixIndent(`
				.method private updateContentInvert()V
					.locals 3

					const-string v2, "None"

					iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mPrefs:Landroid/content/SharedPreferences;
					const-string v1, "content_invert"
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getString(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v2

					iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertValueView:Landroid/widget/TextView;
					invoke-virtual {v0, v2}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V

					iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertImageButton:Landroid/widget/ImageButton;
					const-string v1, "Images"
					invoke-virtual {v1, v2}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z
					move-result v1
					invoke-virtual {v0, v1}, Landroid/widget/ImageButton;->setActivated(Z)V

					iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertPageButton:Landroid/widget/ImageButton;
					const-string v1, "Page"
					invoke-virtual {v1, v2}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z
					move-result v1
					invoke-virtual {v0, v1}, Landroid/widget/ImageButton;->setActivated(Z)V

					return-void
				.end method
				`),
			),
			InMethod("update()V",
				ReplaceStringPrepend(
					FixIndent(`
						iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mTextAlignView:Landroid/view/View;

						const/16 v4, 0x10

						invoke-direct {p0, v0, v4}, Lcom/faultexception/reader/DisplaySettingsFragment;->setVisibilityForFeature(Landroid/view/View;I)V
					`),
					FixIndent(`
						iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertView:Landroid/view/View;
						const/16 v4, 0 # VISIBLE
						invoke-virtual {v0, v4}, Landroid/view/View;->setVisibility(I)V
					`),
				),
				ReplaceStringAppend(
					"\n    invoke-direct {p0}, Lcom/faultexception/reader/DisplaySettingsFragment;->updateTextAlign()V",
					"\n    invoke-direct {p0}, Lcom/faultexception/reader/DisplaySettingsFragment;->updateContentInvert()V",
				),
			),
			InMethod("onCreateView(Landroid/view/LayoutInflater;Landroid/view/ViewGroup;Landroid/os/Bundle;)Landroid/view/View;",
				MustContain(
					FixIndent(`
						invoke-virtual {v0}, Landroid/widget/ImageButton;->getDrawable()Landroid/graphics/drawable/Drawable;

						move-result-object v2

						invoke-virtual {v2}, Landroid/graphics/drawable/Drawable;->mutate()Landroid/graphics/drawable/Drawable;

						move-result-object v2

						invoke-static {v2}, Landroidx/core/graphics/drawable/DrawableCompat;->wrap(Landroid/graphics/drawable/Drawable;)Landroid/graphics/drawable/Drawable;

						move-result-object v2
					`),
				),
				ReplaceStringAppend(
					FixIndent(`
						iput-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mTextAlignView:Landroid/view/View;
					`),
					FixIndent(`
						sget v0, Lcom/faultexception/reader/R$id;->content_invert:I
						invoke-virtual {p2, v0}, Lcom/faultexception/reader/widget/ExpansionScrollView;->findViewById(I)Landroid/view/View;
						move-result-object v0
						iput-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertView:Landroid/view/View;

						sget v0, Lcom/faultexception/reader/R$id;->content_invert_value:I
						invoke-virtual {p2, v0}, Lcom/faultexception/reader/widget/ExpansionScrollView;->findViewById(I)Landroid/view/View;
						move-result-object v0
						check-cast v0, Landroid/widget/TextView;
						iput-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertValueView:Landroid/widget/TextView;

						sget v0, Lcom/faultexception/reader/R$id;->content_invert_image:I
						invoke-virtual {p2, v0}, Lcom/faultexception/reader/widget/ExpansionScrollView;->findViewById(I)Landroid/view/View;
						move-result-object v0
						check-cast v0, Landroid/widget/ImageButton;
						iput-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertImageButton:Landroid/widget/ImageButton;
						invoke-virtual {v0, p0}, Landroid/widget/ImageButton;->setOnClickListener(Landroid/view/View$OnClickListener;)V

						iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertImageButton:Landroid/widget/ImageButton;
						invoke-virtual {v0}, Landroid/widget/ImageButton;->getDrawable()Landroid/graphics/drawable/Drawable;
						move-result-object v2
						invoke-virtual {v2}, Landroid/graphics/drawable/Drawable;->mutate()Landroid/graphics/drawable/Drawable;
						move-result-object v2
						invoke-static {v2}, Landroidx/core/graphics/drawable/DrawableCompat;->wrap(Landroid/graphics/drawable/Drawable;)Landroid/graphics/drawable/Drawable;
						move-result-object v2
						invoke-virtual {v0, v2}, Landroid/widget/ImageButton;->setImageDrawable(Landroid/graphics/drawable/Drawable;)V

						sget v0, Lcom/faultexception/reader/R$id;->content_invert_page:I
						invoke-virtual {p2, v0}, Lcom/faultexception/reader/widget/ExpansionScrollView;->findViewById(I)Landroid/view/View;
						move-result-object v0
						check-cast v0, Landroid/widget/ImageButton;
						iput-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertPageButton:Landroid/widget/ImageButton;
						invoke-virtual {v0, p0}, Landroid/widget/ImageButton;->setOnClickListener(Landroid/view/View$OnClickListener;)V

						iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertPageButton:Landroid/widget/ImageButton;
						invoke-virtual {v0}, Landroid/widget/ImageButton;->getDrawable()Landroid/graphics/drawable/Drawable;
						move-result-object v2
						invoke-virtual {v2}, Landroid/graphics/drawable/Drawable;->mutate()Landroid/graphics/drawable/Drawable;
						move-result-object v2
						invoke-static {v2}, Landroidx/core/graphics/drawable/DrawableCompat;->wrap(Landroid/graphics/drawable/Drawable;)Landroid/graphics/drawable/Drawable;
						move-result-object v2
						invoke-virtual {v0, v2}, Landroid/widget/ImageButton;->setImageDrawable(Landroid/graphics/drawable/Drawable;)V
					`),
				),
			),
			InMethod("onClick(Landroid/view/View;)V",
				ReplaceStringAppend(
					FixIndent(`
						.locals 8
					`),
					FixIndent(`
						invoke-direct {p0, p1}, Lcom/faultexception/reader/DisplaySettingsFragment;->onClickContentInvert(Landroid/view/View;)V
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent(`
				.method public onClick(Landroid/view/View;)V
				`),
				FixIndent(`
				.method private onClickContentInvert(Landroid/view/View;)V
					.locals 7

					iget-object v0, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertImageButton:Landroid/widget/ImageButton;
					iget-object v1, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mContentInvertPageButton:Landroid/widget/ImageButton;

					if-eq p1, v0, :get_current
					if-eq p1, v1, :get_current
					return-void

					:get_current
					const-string v2, "content_invert"
					iget-object v6, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mPrefs:Landroid/content/SharedPreferences;

					const-string v3, "None"
					invoke-interface {v6, v2, v3}, Landroid/content/SharedPreferences;->getString(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v5

					if-eq p1, v0, :check_image
					if-eq p1, v1, :check_page
					return-void

					:check_image
					const-string v4, "Images"
					invoke-virtual {v4, v5}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z
					move-result v5
					goto :toggle

					:check_page
					const-string v4, "Page"
					invoke-virtual {v4, v5}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z
					move-result v5
					goto :toggle

					:toggle
					if-nez v5, :save
					move-object v3, v4
					goto :save

					:save
					invoke-interface {v6}, Landroid/content/SharedPreferences;->edit()Landroid/content/SharedPreferences$Editor;
					move-result-object v6
					invoke-interface {v6, v2, v3}, Landroid/content/SharedPreferences$Editor;->putString(Ljava/lang/String;Ljava/lang/String;)Landroid/content/SharedPreferences$Editor;
					move-result-object v6
					invoke-interface {v6}, Landroid/content/SharedPreferences$Editor;->apply()V

					invoke-direct {p0}, Lcom/faultexception/reader/DisplaySettingsFragment;->updateContentInvert()V
					iget-object v6, p0, Lcom/faultexception/reader/DisplaySettingsFragment;->mOnSettingChangedListener:Lcom/faultexception/reader/DisplaySettingsFragment$OnSettingChangedListener;
					invoke-interface {v6, v3}, Lcom/faultexception/reader/DisplaySettingsFragment$OnSettingChangedListener;->onContentInvertChanged(Ljava/lang/String;)V

					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/DisplaySettingsFragment$OnSettingChangedListener.smali",
			ReplaceStringAppend(
				FixIndent(`
				.method public abstract onTextAlignChanged(I)V
				.end method
				`),
				FixIndent(`
				.method public abstract onContentInvertChanged(Ljava/lang/String;)V
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/ReaderActivity$7.smali",
			ReplaceStringAppend(
				FixIndent(`
				.method public onTextAlignChanged(I)V
					.locals 1

					.line 1873
					iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity$7;->this$0:Lcom/faultexception/reader/ReaderActivity;

					invoke-static {v0}, Lcom/faultexception/reader/ReaderActivity;->access$1100(Lcom/faultexception/reader/ReaderActivity;)Lcom/faultexception/reader/content/BookView;

					move-result-object v0

					if-eqz v0, :cond_0

					.line 1874
					iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity$7;->this$0:Lcom/faultexception/reader/ReaderActivity;

					invoke-static {v0}, Lcom/faultexception/reader/ReaderActivity;->access$1100(Lcom/faultexception/reader/ReaderActivity;)Lcom/faultexception/reader/content/BookView;

					move-result-object v0

					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/BookView;->setTextAlign(I)V

					:cond_0
					return-void
				.end method
				`),
				FixIndent(`
				.method public onContentInvertChanged(Ljava/lang/String;)V
					.locals 1

					iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity$7;->this$0:Lcom/faultexception/reader/ReaderActivity;
					invoke-static {v0}, Lcom/faultexception/reader/ReaderActivity;->access$1100(Lcom/faultexception/reader/ReaderActivity;)Lcom/faultexception/reader/content/BookView;
					move-result-object v0
					if-eqz v0, :cond_0

					iget-object v0, p0, Lcom/faultexception/reader/ReaderActivity$7;->this$0:Lcom/faultexception/reader/ReaderActivity;
					invoke-static {v0}, Lcom/faultexception/reader/ReaderActivity;->access$1100(Lcom/faultexception/reader/ReaderActivity;)Lcom/faultexception/reader/content/BookView;
					move-result-object v0
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/BookView;->setContentInvert(Ljava/lang/String;)V

					:cond_0
					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/content/BookView.smali",
			ReplaceStringAppend(
				FixIndent(`
				.method public setTextAlign(I)V
					.locals 0

					return-void
				.end method
				`),
				FixIndent(`
				.method public setContentInvert(Ljava/lang/String;)V
					.locals 0
					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/content/EPubBookView.smali",
			ReplaceStringAppend(
				"\n"+`.field private mTextAlign:I`,
				"\n"+`.field private mContentInvert:Ljava/lang/String;`,
			),
			ReplaceStringAppend(
				FixIndent(`
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
				FixIndent(`
				.method public setContentInvert(Ljava/lang/String;)V
					.locals 1

					iput-object p1, p0, Lcom/faultexception/reader/content/EPubBookView;->mContentInvert:Ljava/lang/String;

					iget-object v0, p0, Lcom/faultexception/reader/content/EPubBookView;->mContentView:Lcom/faultexception/reader/content/ContentView;
					if-eqz v0, :cond_0

					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/ContentView;->setContentInvert(Ljava/lang/String;)V

					:cond_0
					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/content/ContentView.smali",
			ReplaceStringAppend(
				FixIndent(`
				.method public setTextAlign(I)V
					.locals 0

					return-void
				.end method
				`),
				FixIndent(`
				.method public setContentInvert(Ljava/lang/String;)V
					.locals 0
					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/content/HtmlContentView.smali",
			ReplaceStringAppend(
				FixIndent(`
				.method public setTextAlign(I)V
					.locals 1

					.line 126
					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentView;->mContentWebView:Lcom/faultexception/reader/content/HtmlContentWebView;

					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->setTextAlign(I)V

					return-void
				.end method
				`),
				FixIndent(`
				.method public setContentInvert(Ljava/lang/String;)V
					.locals 1
					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentView;->mContentWebView:Lcom/faultexception/reader/content/HtmlContentWebView;
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->setContentInvert(Ljava/lang/String;)V
					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceStringAppend(
				"\n"+`.field private mTextAlign:I`,
				"\n"+`.field private mContentInvert:Ljava/lang/String;`,
			),
			ReplaceStringPrepend(
				FixIndent(`
					const-string v3, "</script><style type=\'text/css\' id=\'__LithiumThemeStyle\'></style>"

					invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
				`),
				FixIndent(`
					const-string v3, "document.addEventListener('DOMContentLoaded', function() { LithiumJs.setContentInvert('"
					invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					iget-object v3, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContentInvert:Ljava/lang/String;
					invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					const-string v3, "');});"
					invoke-virtual {v5, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
				`),
			),
			ReplaceStringAppend(
				FixIndent(`
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
				FixIndent(`
				.method public setContentInvert(Ljava/lang/String;)V
					.locals 2
					iput-object p1, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mContentInvert:Ljava/lang/String;

					iget-object v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mUrl:Ljava/lang/String;
					if-eqz v0, :cond_0

					iget-boolean v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mDisplaySettingsInjected:Z
					if-eqz v0, :cond_0

					new-instance v0, Ljava/lang/StringBuilder;
					invoke-direct {v0}, Ljava/lang/StringBuilder;-><init>()V
					const-string v1, "LithiumJs.setContentInvert('"
					invoke-virtual {v0, v1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v0, p1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					const-string p1, "')"
					invoke-virtual {v0, p1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v0}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;
					move-result-object p1
					invoke-virtual {p0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->executeJavascript(Ljava/lang/String;)V

					:cond_0
					return-void
				.end method
				`),
			),
		),

		PatchFile("assets/js/epub.js",
			ReplaceStringPrepend(
				"\n    var styleElement = void 0;",
				"\n    var invertStyleElement = void 0;",
			),
			ReplaceStringPrepend(
				FixIndent(`
						styleElement = document.createElement('style');
						styleElement.setAttribute('type', 'text/css');
						document.head.appendChild(styleElement);

				`),
				FixIndent(`
						invertStyleElement = document.createElement('style');
						invertStyleElement.setAttribute('type', 'text/css');
						document.head.appendChild(invertStyleElement);
				`),
			),
			ReplaceStringAppend(
				FixIndent(`
					function setTextAlign(align) {
						textAlign = align;
						updateStyleElement();
						reflowIfNecessary();
					}
				`),
				FixIndent(`
					function setContentInvert(target) {
						var css = ' { filter: invert(1) hue-rotate(180deg) brightness(0.9) }';
						switch (target) {
						case 'Page':
							invertStyleElement.innerText = 'html' + css;
							break;
						case 'Images':
							invertStyleElement.innerText = 'img, svg' + css;
							break;
						default:
							invertStyleElement.innerText = '';
							break;
						}
					}
				`),
			),
			ReplaceStringPrepend(
				"\n        setTextAlign: setTextAlign",
				"\n        setContentInvert: setContentInvert,",
			),
		),
	)
}
