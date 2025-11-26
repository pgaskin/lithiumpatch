// # Grid/list view reading progress
//
// Add an option to show reading progress in the grid and list views.
package patches

import (
	"regexp"

	. "github.com/pgaskin/lithiumpatch/patches/patchdef"
)

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
				"\n"+`    <item type="id" name="reading_progress" />`,
			),
		),
		DefineR("smali/com/faultexception/reader", "id", "reading_progress"),
		// GRID VIEW - Add progress badge on cover
		PatchFile("res/layout/books_grid_item.xml",
			ReplaceStringRe(
				regexp.MustCompile(`(?s)<ImageView android:id="@id/cover".*?/>`),
				"$0"+"\n"+`            <TextView android:textSize="10.0sp" android:textColor="#ffffffff" android:gravity="center" android:id="@id/reading_progress" android:background="#cc000000" android:paddingLeft="4.0dip" android:paddingTop="2.0dip" android:paddingRight="4.0dip" android:paddingBottom="2.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:layout_gravity="top|right" android:layout_margin="4.0dip" android:fontFamily="sans-serif-medium" />`,
			),
		),
		// LIST VIEW - Add progress badge on cover
		PatchFile("res/layout/books_list_item.xml",
			ReplaceStringRe(
				regexp.MustCompile(`(?s)<ImageView android:layout_gravity="center" android:id="@id/noCover".*?/>`),
				"$0"+"\n"+`            <TextView android:textSize="10.0sp" android:textColor="#ffffffff" android:gravity="center" android:id="@id/reading_progress" android:background="#cc000000" android:paddingLeft="4.0dip" android:paddingTop="2.0dip" android:paddingRight="4.0dip" android:paddingBottom="2.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:layout_gravity="top|right" android:layout_margin="2.0dip" android:fontFamily="sans-serif-medium" />`,
			),
		),
		PatchFile("res/layout-v17/books_list_item.xml",
			ReplaceStringRe(
				regexp.MustCompile(`(?s)<ImageView android:layout_gravity="center" android:id="@id/noCover".*?/>`),
				"$0"+"\n"+`            <TextView android:textSize="10.0sp" android:textColor="#ffffffff" android:gravity="center" android:id="@id/reading_progress" android:background="#cc000000" android:paddingLeft="4.0dip" android:paddingTop="2.0dip" android:paddingRight="4.0dip" android:paddingBottom="2.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:layout_gravity="top|right" android:layout_margin="2.0dip" android:fontFamily="sans-serif-medium" />`,
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
						sget p1, Lcom/faultexception/reader/R$id;->reading_progress:I
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
			// Add page map cache field
			ReplaceStringAppend(
				"\n"+`.field private mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;`,
				"\n"+`.field private mPageMapCache:Ljava/util/Map;`,
			),
			// Initialize page map cache in constructor
			InMethod("<init>(Landroidx/appcompat/app/AppCompatActivity;Lcom/faultexception/reader/BooksAdapter$OnItemClickListener;Lcom/faultexception/reader/util/ActionModeMultiCallback;)V",
				ReplaceStringAppend(
					"\n"+`    iput-object p1, p0, Lcom/faultexception/reader/BooksAdapter;->mGlide:Lcom/bumptech/glide/RequestManager;`,
					"\n"+`    new-instance v0, Ljava/util/HashMap;`+"\n"+`    invoke-direct {v0}, Ljava/util/HashMap;-><init>()V`+"\n"+`    iput-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mPageMapCache:Ljava/util/Map;`,
				),
			),
			// Add method to calculate progress percentage
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V
				`),
				FixIndent("\n"+`
				.method private getPageMap(J)Lcom/faultexception/reader/book/EPubPageMap;
					.locals 4

					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mPageMapCache:Ljava/util/Map;
					if-nez v0, :has_cache
					new-instance v0, Ljava/util/HashMap;
					invoke-direct {v0}, Ljava/util/HashMap;-><init>()V
					iput-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mPageMapCache:Ljava/util/Map;

					:has_cache
					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mPageMapCache:Ljava/util/Map;
					invoke-static {p1, p2}, Ljava/lang/Long;->valueOf(J)Ljava/lang/Long;
					move-result-object v1
					invoke-interface {v0, v1}, Ljava/util/Map;->get(Ljava/lang/Object;)Ljava/lang/Object;
					move-result-object v0
					check-cast v0, Lcom/faultexception/reader/book/EPubPageMap;

					if-nez v0, :return_map

					iget-object v2, p0, Lcom/faultexception/reader/BooksAdapter;->mActivity:Landroidx/appcompat/app/AppCompatActivity;
					invoke-static {v2}, Lcom/faultexception/reader/db/DatabaseProvider;->getDatabase(Landroid/content/Context;)Landroid/database/sqlite/SQLiteDatabase;
					move-result-object v2
					invoke-static {p1, p2, v2}, Lcom/faultexception/reader/book/EPubPageMap;->readFromCache(JLandroid/database/sqlite/SQLiteDatabase;)Lcom/faultexception/reader/book/EPubPageMap;
					move-result-object v0

					if-eqz v0, :return_map

					iget-object v2, p0, Lcom/faultexception/reader/BooksAdapter;->mPageMapCache:Ljava/util/Map;
					invoke-interface {v2, v1, v0}, Ljava/util/Map;->put(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;

					:return_map
					return-object v0
				.end method

				.method private getProgressPercentage()I
					.locals 14

					:try_start_0
					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mCursor:Landroid/database/Cursor;
					iget-object v1, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
					iget v1, v1, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->progress:I
					invoke-interface {v0, v1}, Landroid/database/Cursor;->getString(I)Ljava/lang/String;
					move-result-object v0

					if-nez v0, :cond_0
					goto :return_zero

					:cond_0
					invoke-virtual {v0}, Ljava/lang/String;->isEmpty()Z
					move-result v1
					if-eqz v1, :cond_1
					goto :return_zero

					:cond_1
					iget-object v1, p0, Lcom/faultexception/reader/BooksAdapter;->mCursor:Landroid/database/Cursor;
					iget-object v2, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
					iget v2, v2, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->id:I
					invoke-interface {v1, v2}, Landroid/database/Cursor;->getLong(I)J
					move-result-wide v1

					new-instance v7, Lcom/google/gson/JsonParser;
					invoke-direct {v7}, Lcom/google/gson/JsonParser;-><init>()V
					invoke-virtual {v7, v0}, Lcom/google/gson/JsonParser;->parse(Ljava/lang/String;)Lcom/google/gson/JsonElement;
					move-result-object v0
					invoke-virtual {v0}, Lcom/google/gson/JsonElement;->getAsJsonObject()Lcom/google/gson/JsonObject;
					move-result-object v0

					const-string v3, "position"
					invoke-virtual {v0, v3}, Lcom/google/gson/JsonObject;->has(Ljava/lang/String;)Z
					move-result v4
					if-nez v4, :cond_2
					goto :return_zero

					:cond_2
					const-string v4, "url"
					invoke-virtual {v0, v4}, Lcom/google/gson/JsonObject;->has(Ljava/lang/String;)Z
					move-result v5
					if-nez v5, :cond_3
					goto :return_zero

					:cond_3
					invoke-virtual {v0, v3}, Lcom/google/gson/JsonObject;->get(Ljava/lang/String;)Lcom/google/gson/JsonElement;
					move-result-object v3
					invoke-virtual {v3}, Lcom/google/gson/JsonElement;->getAsFloat()F
					move-result v3

					invoke-virtual {v0, v4}, Lcom/google/gson/JsonObject;->get(Ljava/lang/String;)Lcom/google/gson/JsonElement;
					move-result-object v0
					invoke-virtual {v0}, Lcom/google/gson/JsonElement;->getAsString()Ljava/lang/String;
					move-result-object v0

					invoke-direct {p0, v1, v2}, Lcom/faultexception/reader/BooksAdapter;->getPageMap(J)Lcom/faultexception/reader/book/EPubPageMap;
					move-result-object v4

					const/high16 v5, 0x42c80000    # 100.0f

					if-eqz v4, :fallback_percent
					iget-object v6, v4, Lcom/faultexception/reader/book/EPubPageMap;->items:Ljava/util/Map;
					invoke-interface {v6, v0}, Ljava/util/Map;->get(Ljava/lang/Object;)Ljava/lang/Object;
					move-result-object v6
					check-cast v6, Lcom/faultexception/reader/book/EPubPageMap$Item;

					if-eqz v6, :fallback_percent
					iget v7, v6, Lcom/faultexception/reader/book/EPubPageMap$Item;->pageCount:I
					if-lez v7, :fallback_percent
					iget v4, v4, Lcom/faultexception/reader/book/EPubPageMap;->totalPageCount:I
					if-lez v4, :fallback_percent

					iget v8, v6, Lcom/faultexception/reader/book/EPubPageMap$Item;->pageStart:I
					add-int/lit8 v7, v7, -0x1
					int-to-float v7, v7
					mul-float v7, v7, v3
					invoke-static {v7}, Ljava/lang/Math;->round(F)I
					move-result v7
					add-int/2addr v8, v7
					add-int/lit8 v8, v8, 0x1
					int-to-float v7, v8
					mul-float v7, v7, v5
					int-to-float v4, v4
					div-float/2addr v7, v4
					invoke-static {v7}, Ljava/lang/Math;->round(F)I
					move-result v0
					return v0

					:fallback_percent
					mul-float v3, v3, v5
					invoke-static {v3}, Ljava/lang/Math;->round(F)I
					move-result v0
					if-gtz v0, :return_fb
					const/4 v5, 0x0
					cmpl-float v6, v3, v5
					if-lez v6, :return_fb
					const/4 v0, 0x1
					:return_fb
					return v0

					:try_end_0
					.catch Ljava/lang/Exception; {:try_start_0 .. :try_end_0} :catch_0

					:catch_0
					move-exception v0

					:return_zero
					const/4 v0, 0x0
					return v0
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
					.locals 5

					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mActivity:Landroidx/appcompat/app/AppCompatActivity;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "show_grid_progress"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v1

					iget-object v3, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->progressView:Landroid/widget/TextView;

					if-nez v1, :pref_enabled

					const/16 v2, 0x8 # android.View.GONE
					invoke-virtual {v3, v2}, Landroid/widget/TextView;->setVisibility(I)V
					return-void

					:pref_enabled
					# Get percentage
					invoke-direct {p0}, Lcom/faultexception/reader/BooksAdapter;->getProgressPercentage()I
					move-result v0

					# Check if percentage > 0
					if-lez v0, :hide_progress

					# Format percentage string
					new-instance v1, Ljava/lang/StringBuilder;
					invoke-direct {v1}, Ljava/lang/StringBuilder;-><init>()V
					invoke-virtual {v1, v0}, Ljava/lang/StringBuilder;->append(I)Ljava/lang/StringBuilder;
					const-string v0, "%"
					invoke-virtual {v1, v0}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v1}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;
					move-result-object v0

					# Show percentage
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
			// Call updateProgressView in onBindViewHolder for both layouts (just after cover/noCover handling)
			InMethod("onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V",
				ReplaceStringAppend(
					"\n"+`    iget-object v0, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->noCoverView:Landroid/view/View;`+"\n\n"+`    if-eqz p2, :cond_7`,
					"\n"+`    iget-object v0, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->noCoverView:Landroid/view/View;`+"\n\n"+`    if-eqz p2, :cond_7`+"\n"+`    invoke-direct {p0, p1}, Lcom/faultexception/reader/BooksAdapter;->updateProgressView(Lcom/faultexception/reader/BooksAdapter$ViewHolder;)V`,
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
						const-string v1, "current_position"
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
						const-string v1, "current_position"
						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I
						move-result v1
						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->progress:I

					`),
				),
			),
		),
	)
}
