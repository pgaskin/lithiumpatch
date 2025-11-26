// # Grid/list view reading progress
//
// Add an option to show reading progress in the grid and list views.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("gridprogress",
		// Add preference toggle
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Show reading progress (grid/list view)" android:key="show_grid_progress" android:defaultValue="false" />`,
			),
		),
		// Add progress view ID
		PatchFile("res/values/ids.xml",
			ReplaceStringPrepend(
				"\n"+`</resources>`,
				"\n"+`    <item type="id" name="progress" />`,
			),
		),
		DefineR("smali/com/faultexception/reader", "id", "progress"),
		// GRID VIEW - Add progress badge on cover
		PatchFile("res/layout/books_grid_item.xml",
			ReplaceStringAppend(
				"\n"+`            <ImageView android:id="@id/cover" android:layout_width="fill_parent" android:layout_height="fill_parent" android:scaleType="centerCrop" />`,
				"\n"+`            <TextView android:textSize="10.0sp" android:textColor="#ffffffff" android:gravity="center" android:id="@id/progress" android:background="#cc000000" android:paddingLeft="4.0dip" android:paddingTop="2.0dip" android:paddingRight="4.0dip" android:paddingBottom="2.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:layout_gravity="top|right" android:layout_margin="4.0dip" android:fontFamily="sans-serif-medium" />`,
			),
		),
		// LIST VIEW - Add progress as new line after author
		PatchFile("res/layout/books_list_item.xml",
			ReplaceStringAppend(
				"\n"+`            <TextView android:textSize="14.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/creator" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
				"\n"+`            <TextView android:textSize="12.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/progress" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
			),
		),
		PatchFile("res/layout-v17/books_list_item.xml",
			ReplaceStringAppend(
				"\n"+`            <TextView android:textSize="14.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/creator" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
				"\n"+`            <TextView android:textSize="12.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/progress" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
			),
		),
		// Add progress field to ViewHolder
		PatchFile("smali/com/faultexception/reader/BooksAdapter$ViewHolder.smali",
			ReplaceStringAppend(
				"\n"+`.field public creatorView:Landroid/widget/TextView;`,
				"\n"+`.field public progressView:Landroid/widget/TextView;`,
			),
			InMethod("<init>(Lcom/faultexception/reader/BooksAdapter;Landroid/view/View;)V",
				// must follow pattern of previous (p0=ViewHolder p1=id p2=this)
				ReplaceStringAppend(
					FixIndent("\n"+`
						invoke-virtual {p2, p1}, Landroid/view/View;->findViewById(I)Landroid/view/View;

						move-result-object p1

						check-cast p1, Landroid/widget/TextView;

						iput-object p1, p0, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->creatorView:Landroid/widget/TextView;
					`),
					FixIndent("\n"+`
						sget p1, Lcom/faultexception/reader/R$id;->progress:I
						invoke-virtual {p2, p1}, Landroid/view/View;->findViewById(I)Landroid/view/View;
						move-result-object p1
						check-cast p1, Landroid/widget/TextView;
						iput-object p1, p0, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->progressView:Landroid/widget/TextView;
					`),
				),
			),
		),
		// Add progress column index to CursorIndexContainer
		PatchFile("smali/com/faultexception/reader/BooksAdapter$CursorIndexContainer.smali",
			ReplaceStringAppend(
				"\n"+`.field creator:I`,
				"\n"+`.field progress:I`,
			),
		),
		// Modify BooksAdapter to handle progress display
		PatchFile("smali/com/faultexception/reader/BooksAdapter.smali",
			// Add method to get progress string
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V
				`),
				FixIndent("\n"+`
				.method private getProgressString(F)Ljava/lang/String;
					.locals 3

					# v0=percentage int
					const/high16 v1, 0x42c80000     # 100.0f
					mul-float/2addr p1, v1
					invoke-static {p1}, Ljava/lang/Math;->round(F)I
					move-result v0

					# return percentage as string with % sign
					new-instance v1, Ljava/lang/StringBuilder;
					invoke-direct {v1}, Ljava/lang/StringBuilder;-><init>()V
					invoke-virtual {v1, v0}, Ljava/lang/StringBuilder;->append(I)Ljava/lang/StringBuilder;
					const-string v2, "%"
					invoke-virtual {v1, v2}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v1}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;
					move-result-object v1

					return-object v1
				.end method
				`),
			),
			// Add method to update progress view
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V
				`),
				FixIndent("\n"+`
				.method private updateProgressView(Lcom/faultexception/reader/BooksAdapter$ViewHolder;)V
					.locals 4

					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mActivity:Landroidx/appcompat/app/AppCompatActivity;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "show_grid_progress"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v1

					iget-object v3, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->progressView:Landroid/widget/TextView;

					if-nez v1, :show_progress
					const/16 v2, 0x8 # android.View.GONE
					invoke-virtual {v3, v2}, Landroid/widget/TextView;->setVisibility(I)V
					return-void

					:show_progress
					# get progress from cursor
					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mCursor:Landroid/database/Cursor;
					iget-object v1, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
					iget v1, v1, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->progress:I
					invoke-interface {v0, v1}, Landroid/database/Cursor;->getFloat(I)F
					move-result v0

					# check if progress > 0
					const/4 v1, 0x0
					cmpl-float v1, v0, v1
					if-lez v1, :hide_progress

					# show progress
					invoke-direct {p0, v0}, Lcom/faultexception/reader/BooksAdapter;->getProgressString(F)Ljava/lang/String;
					move-result-object v0
					invoke-virtual {v3, v0}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V
					const/16 v2, 0x0 # android.View.VISIBLE
					invoke-virtual {v3, v2}, Landroid/widget/TextView;->setVisibility(I)V
					return-void

					:hide_progress
					const/16 v2, 0x8 # android.View.GONE
					invoke-virtual {v3, v2}, Landroid/widget/TextView;->setVisibility(I)V
					return-void
				.end method
				`),
			),
			// Call updateProgressView in onBindViewHolder for grid view (layoutMode 0x0)
			InMethod("onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V",
				// must follow pattern of previous - add after footer handling
				ReplaceStringAppend(
					"\n"+`    .line 193`+"\n"+`    iget-object v2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->footerView:Landroid/view/View;`,
					"\n"+`    invoke-direct {p0, p1}, Lcom/faultexception/reader/BooksAdapter;->updateProgressView(Lcom/faultexception/reader/BooksAdapter$ViewHolder;)V`,
				),
			),
			// Call updateProgressView in onBindViewHolder for list view (layoutMode 0x1)
			InMethod("onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V",
				// must follow pattern of previous - add after footer handling
				ReplaceStringAppend(
					"\n"+`    .line 195`+"\n"+`    iget-object v2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->footerView:Landroid/view/View;`,
					"\n"+`    invoke-direct {p0, p1}, Lcom/faultexception/reader/BooksAdapter;->updateProgressView(Lcom/faultexception/reader/BooksAdapter$ViewHolder;)V`,
				),
			),
			// Add progress column index in swapCursor - must follow pattern of previous
			InMethod(`swapCursor(Landroid/database/Cursor;)V`,
				ReplaceStringAppend(
					FixIndent("\n"+`
						iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;

						const-string v1, "creator"

						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I

						move-result v1

						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->creator:I
					`),
					FixIndent("\n"+`
						iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
						const-string v1, "progress"
						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I
						move-result v1
						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->progress:I
					`),
				),
			),
			// Add progress column index in swapCursor (second occurrence) - must follow pattern of previous
			InMethod(`swapCursor(Landroid/database/Cursor;)V`,
				ReplaceStringAppend(
					FixIndent("\n"+`
						iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;

						const-string v1, "creator"

						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I

						move-result v1

						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->creator:I
					`),
					FixIndent("\n"+`
						iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
						const-string v1, "progress"
						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I
						move-result v1
						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->progress:I
					`),
				),
			),
		),
	)
}
