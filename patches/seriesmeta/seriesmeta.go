// Package seriesmeta adds support for parsing and displaying calibre-style
// series metadata.
package seriesmeta

import (
	. "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"
)

// TODO: replace the "must follow pattern" ones with a regex to automate updating them, the others are already version-independent
// TODO: add epub3-style metadata support somehow
// TODO: refactor this... it might be easier just to run through the OPF twice and have an entirely separate parsing function for series metadata

func init() {
	Register("seriesmeta",
		// DISPLAY
		PatchFile("res/values/ids.xml",
			ReplaceStringPrepend(
				"\n"+`</resources>`,
				"\n"+`    <item type="id" name="series" />`,
			),
		),
		DefineR("smali/com/faultexception/reader", "id", "series"),
		PatchFile("res/layout/books_grid_item.xml",
			ReplaceStringAppend(
				"\n"+`            <TextView android:textSize="12.0sp" android:textColor="#ffffffff" android:ellipsize="end" android:id="@id/creator" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
				"\n"+`            <TextView android:textSize="10.0sp" android:textColor="#ffffffff" android:ellipsize="end" android:id="@id/series" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif"/>`,
			),
		),
		PatchFile("res/layout/books_list_item.xml",
			ReplaceString(
				`<FrameLayout android:id="@id/cover_container" android:layout_width="48.0dip" android:layout_height="fill_parent">`,
				`<FrameLayout android:id="@id/cover_container" android:layout_width="60.0dip" android:layout_height="fill_parent">`,
			),
			ReplaceStringAppend(
				"\n"+`            <TextView android:textSize="14.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/creator" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
				"\n"+`            <TextView android:textSize="14.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/series" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
			),
		),
		PatchFile("res/layout-v17/books_list_item.xml",
			ReplaceString(
				`<FrameLayout android:id="@id/cover_container" android:layout_width="48.0dip" android:layout_height="fill_parent">`,
				`<FrameLayout android:id="@id/cover_container" android:layout_width="60.0dip" android:layout_height="fill_parent">`,
			),
			ReplaceStringAppend(
				"\n"+`            <TextView android:textSize="14.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/creator" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
				"\n"+`            <TextView android:textSize="14.0sp" android:textColor="?android:textColorSecondary" android:ellipsize="end" android:id="@id/series" android:layout_width="fill_parent" android:layout_height="wrap_content" android:maxLines="1" android:fontFamily="sans-serif" />`,
			),
		),
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Show series metadata" android:key="series_metadata" android:defaultValue="false" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/BooksFragment.smali",
			ReplaceString(
				`"title LIKE ? OR creator LIKE ?"`,
				`"title LIKE ? OR (creator || series) LIKE ?"`, // a hack to not have to reorder the whole parameter array
			),
			ReplaceString(
				`"creator ASC"`,
				`"creator ASC, series ASC, LENGTH(series_index) ASC, series_index ASC"`,
			),
		),
		PatchFile("smali/com/faultexception/reader/BooksAdapter$ViewHolder.smali",
			ReplaceStringAppend(
				"\n"+`.field public creatorView:Landroid/widget/TextView;`,
				"\n"+`.field public seriesView:Landroid/widget/TextView;`,
			),
			InMethod("<init>(Lcom/faultexception/reader/BooksAdapter;Landroid/view/View;)V",
				StringPatcherFunc(func(smali string) (string, error) {
					// must follow pattern of previous (p0=ViewHolder p1=id p2=this)
					return ReplaceStringAppend(
						FixIndent("\n"+`
							invoke-virtual {p2, p1}, Landroid/view/View;->findViewById(I)Landroid/view/View;

							move-result-object p1

							check-cast p1, Landroid/widget/TextView;

							iput-object p1, p0, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->creatorView:Landroid/widget/TextView;
						`),
						FixIndent("\n"+`
							sget p1, Lcom/faultexception/reader/R$id;->series:I
							invoke-virtual {p2, p1}, Landroid/view/View;->findViewById(I)Landroid/view/View;
							move-result-object p1
							check-cast p1, Landroid/widget/TextView;
							iput-object p1, p0, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->seriesView:Landroid/widget/TextView;
						`),
					).PatchString(smali)
				}),
			),
		),
		PatchFile("smali/com/faultexception/reader/BooksAdapter.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private highlightSearchQuery(Ljava/lang/String;)Landroid/text/Spannable;
				`),
				FixIndent("\n"+`
				.method private getCurrentSeriesString()Ljava/lang/String;
					.locals 7

					# v0=cursor, v1=stringbuilder
					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mCursor:Landroid/database/Cursor;
					new-instance v1, Ljava/lang/StringBuilder;
					invoke-direct {v1}, Ljava/lang/StringBuilder;-><init>()V

					# v2=series
					iget-object v2, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
					iget v2, v2, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->series:I
					invoke-interface {v0, v2}, Landroid/database/Cursor;->getString(I)Ljava/lang/String;
					move-result-object v2

					if-eqz v2, :retstr
					invoke-virtual {v1, v2}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					# v3=series_index, v4=separator, v5=find, v6=replace
					iget-object v3, p0, Lcom/faultexception/reader/BooksAdapter;->mIndexes:Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;
					iget v3, v3, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->seriesIndex:I
					invoke-interface {v0, v3}, Landroid/database/Cursor;->getString(I)Ljava/lang/String;
					move-result-object v3

					if-eqz v3, :retstr
					const-string v5, ".0"
					const-string v6, ""
					invoke-virtual {v3, v5, v6}, Ljava/lang/String;->replace(Ljava/lang/CharSequence;Ljava/lang/CharSequence;)Ljava/lang/String;
					move-result-object v3

					const-string v4, " #"
					invoke-virtual {v1, v4}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;
					invoke-virtual {v1, v3}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					:retstr
					invoke-virtual {v1}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;
					move-result-object v1

					return-object v1
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V
				`),
				FixIndent("\n"+`
				.method private maybeHideSeries(Landroid/widget/TextView;)V
					.locals 3

					iget-object v0, p0, Lcom/faultexception/reader/BooksAdapter;->mActivity:Landroidx/appcompat/app/AppCompatActivity;
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "series_metadata"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v1

					const/16 v2, 0x0 # android.View.VISIBLE
					if-nez v1, :visible
					const/16 v2, 0x8 # android.View.GONE
					:visible
					invoke-virtual {p1, v2}, Landroid/widget/TextView;->setVisibility(I)V

					return-void
				.end method
				`),
			),
			InMethod("onBindViewHolder(Lcom/faultexception/reader/BooksAdapter$ViewHolder;I)V",
				// v0 must be used to store the title or creator, as we will be overriding it after it is done with
				MustContain(`invoke-interface {p2, v0}, Landroid/database/Cursor;->getString(I)Ljava/lang/String;`),
				// must follow pattern of previous (p0=BooksAdapter p1=ViewHolder p2=TextView v0=string)
				ReplaceStringAppend(
					FixIndent("\n"+`
						iget-object p2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->creatorView:Landroid/widget/TextView;

						invoke-direct {p0, v0}, Lcom/faultexception/reader/BooksAdapter;->highlightSearchQuery(Ljava/lang/String;)Landroid/text/Spannable;

						move-result-object v0

						invoke-virtual {p2, v0}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V
					`),
					FixIndent("\n"+`
						iget-object p2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->seriesView:Landroid/widget/TextView;
						invoke-direct {p0}, Lcom/faultexception/reader/BooksAdapter;->getCurrentSeriesString()Ljava/lang/String;
						move-result-object v0
						invoke-direct {p0, v0}, Lcom/faultexception/reader/BooksAdapter;->highlightSearchQuery(Ljava/lang/String;)Landroid/text/Spannable;
						move-result-object v0
						invoke-virtual {p2, v0}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V
						invoke-direct {p0, p2}, Lcom/faultexception/reader/BooksAdapter;->maybeHideSeries(Landroid/widget/TextView;)V
					`),
				),
				// must follow pattern of previous (p0=BooksAdapter p1=ViewHolder p2=TextView v0=string)
				ReplaceStringAppend(
					FixIndent("\n"+`
						iget-object p2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->creatorView:Landroid/widget/TextView;

						invoke-virtual {p2, v0}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V
					`),
					FixIndent("\n"+`
						iget-object p2, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->seriesView:Landroid/widget/TextView;
						invoke-direct {p0}, Lcom/faultexception/reader/BooksAdapter;->getCurrentSeriesString()Ljava/lang/String;
						move-result-object v0
						invoke-virtual {p2, v0}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V
						invoke-direct {p0, p2}, Lcom/faultexception/reader/BooksAdapter;->maybeHideSeries(Landroid/widget/TextView;)V
					`),
				),
				// must follow pattern of previous (p1=ViewHolder v2=color v3=TextView)
				ReplaceStringAppend(
					FixIndent("\n"+`
						iget-object v3, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->creatorView:Landroid/widget/TextView;

						invoke-virtual {v3, v2}, Landroid/widget/TextView;->setTextColor(I)V
					`),
					FixIndent("\n"+`
						iget-object v3, p1, Lcom/faultexception/reader/BooksAdapter$ViewHolder;->seriesView:Landroid/widget/TextView;
						invoke-virtual {v3, v2}, Landroid/widget/TextView;->setTextColor(I)V
					`),
				),
			),
			InMethod(`swapCursor(Landroid/database/Cursor;)V`,
				// must follow pattern of previous (p0=BooksAdapter v0=CursorIndexContainer v1=column,index)
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
						const-string v1, "series_index"
						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I
						move-result v1
						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->seriesIndex:I
					`),
				),
				// must follow pattern of previous (p0=BooksAdapter v0=CursorIndexContainer v1=column,index)
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
						const-string v1, "series"
						invoke-interface {p1, v1}, Landroid/database/Cursor;->getColumnIndexOrThrow(Ljava/lang/String;)I
						move-result v1
						iput v1, v0, Lcom/faultexception/reader/BooksAdapter$CursorIndexContainer;->series:I
					`),
				),
			),
		),
		// DATABASE
		PatchFile("smali/com/faultexception/reader/db/BooksTable.smali",
			ReplaceStringAppend(
				"\n"+`.field public static final COLUMN_CREATOR:Ljava/lang/String; = "creator"`,
				"\n"+`.field public static final COLUMN_SERIES_INDEX:Ljava/lang/String; = "series_index"`,
			),
			ReplaceStringAppend(
				"\n"+`.field public static final COLUMN_CREATOR:Ljava/lang/String; = "creator"`,
				"\n"+`.field public static final COLUMN_SERIES:Ljava/lang/String; = "series"`,
			),
		),
		PatchFile("smali/com/faultexception/reader/BooksAdapter$CursorIndexContainer.smali",
			ReplaceStringAppend(
				"\n"+`.field creator:I`,
				"\n"+`.field seriesIndex:I`,
			),
			ReplaceStringAppend(
				"\n"+`.field creator:I`,
				"\n"+`.field series:I`,
			),
		),
		PatchFile("smali/com/faultexception/reader/db/DatabaseOpenHelper.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public constructor <init>(Landroid/content/Context;)V
				`),
				FixIndent("\n"+`
				.method private static tryAddSeriesStuff(Landroid/database/sqlite/SQLiteDatabase;)V
					.locals 1
					:ts
					const-string v0, "ALTER TABLE books ADD COLUMN series text default null;"
					invoke-virtual {p0, v0}, Landroid/database/sqlite/SQLiteDatabase;->execSQL(Ljava/lang/String;)V
					:te
					.catch Ljava/lang/Exception; {:ts .. :te} :ts1
					:ts1
					const-string v0, "ALTER TABLE books ADD COLUMN series_index text default null;"
					invoke-virtual {p0, v0}, Landroid/database/sqlite/SQLiteDatabase;->execSQL(Ljava/lang/String;)V
					:te1
					.catch Ljava/lang/Exception; {:ts1 .. :te1} :ret
					:ret
					return-void
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public constructor <init>(Landroid/content/Context;)V
				`),
				FixIndent("\n"+`
				.method public onOpen(Landroid/database/sqlite/SQLiteDatabase;)V
					.locals 0
					invoke-static {p1}, Lcom/faultexception/reader/db/DatabaseOpenHelper;->tryAddSeriesStuff(Landroid/database/sqlite/SQLiteDatabase;)V
					return-void
				.end method
				`),
			),
		),
		// METADATA
		PatchFile("smali/com/faultexception/reader/book/Book.smali",
			ReplaceStringAppend(
				"\n"+`.method public abstract getCreator()Ljava/lang/String;`+"\n"+`.end method`,
				"\n"+`.method public abstract getSeriesIndex()Ljava/lang/String;`+"\n"+`.end method`,
			),
			ReplaceStringAppend(
				"\n"+`.method public abstract getCreator()Ljava/lang/String;`+"\n"+`.end method`,
				"\n"+`.method public abstract getSeries()Ljava/lang/String;`+"\n"+`.end method`,
			),
		),
		PatchFile("smali/com/faultexception/reader/book/EPubBook.smali",
			ReplaceStringAppend(
				"\n"+`.field private mCreator:Ljava/lang/String;`,
				"\n"+`.field private mSeriesIndex:Ljava/lang/String;`,
			),
			ReplaceStringAppend(
				"\n"+`.field private mCreator:Ljava/lang/String;`,
				"\n"+`.field private mSeries:Ljava/lang/String;`,
			),
			InMethod("readOpfFile(Ljava/lang/String;)V",
				MustContain(`iget-object v2, v1, Lcom/faultexception/reader/book/EPubBook;->mRendition:Lcom/faultexception/reader/book/Rendition;`), // ensure EPubBook is in v1
				MustContain(`invoke-interface {v5}, Lorg/xmlpull/v1/XmlPullParser;->nextText()Ljava/lang/String;`),                                  // ensure XmlPullParser is in v5
				MustContain(`const-string v8, "package/metadata/meta"`),                                                                             // links to a goto then a switch
				// remember to make sure pswitch_4 comes from the switch table index of the package/metadata/meta
				ReplaceStringAppend(
					"\n"+`    :pswitch_4`,
					"\n"+`    invoke-direct {v1, v5}, Lcom/faultexception/reader/book/EPubBook;->tryParseSeriesIndex(Lorg/xmlpull/v1/XmlPullParser;)V`,
				),
				ReplaceStringAppend(
					"\n"+`    :pswitch_4`,
					"\n"+`    invoke-direct {v1, v5}, Lcom/faultexception/reader/book/EPubBook;->tryParseSeries(Lorg/xmlpull/v1/XmlPullParser;)V`,
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private readOpfFile(Ljava/lang/String;)V
				`),
				FixIndent("\n"+`
				.method private tryParseMeta(Lorg/xmlpull/v1/XmlPullParser;Ljava/lang/String;)Ljava/lang/String;
					.locals 2
					# p0=book p1=parser p2=name-value-target v0=namespace/null-return v1=attr/value/cond
					const/4 v0, 0x0
					const-string v1, "name"
					invoke-interface {p1, v0, v1}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v1
					if-eqz v1, :ret
					invoke-virtual {v1, p2}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z
					move-result v1
					if-eqz v1, :ret
					const-string v1, "content"
					invoke-interface {p1, v0, v1}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v1
					return-object v1
					:ret
					return-object v0
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private readOpfFile(Ljava/lang/String;)V
				`),
				FixIndent("\n"+`
				.method private tryParseSeriesIndex(Lorg/xmlpull/v1/XmlPullParser;)V
					.locals 1
					# p0=book p1=parser v0=name/content
					const-string v0, "calibre:series_index"
					invoke-direct {p0, p1, v0}, Lcom/faultexception/reader/book/EPubBook;->tryParseMeta(Lorg/xmlpull/v1/XmlPullParser;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v0
					if-eqz v0, :ret
					iput-object v0, p0, Lcom/faultexception/reader/book/EPubBook;->mSeriesIndex:Ljava/lang/String;
					:ret
					return-void
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private readOpfFile(Ljava/lang/String;)V
				`),
				FixIndent("\n"+`
				.method private tryParseSeries(Lorg/xmlpull/v1/XmlPullParser;)V
					.locals 1
					# p0=book p1=parser v0=name/content
					const-string v0, "calibre:series"
					invoke-direct {p0, p1, v0}, Lcom/faultexception/reader/book/EPubBook;->tryParseMeta(Lorg/xmlpull/v1/XmlPullParser;Ljava/lang/String;)Ljava/lang/String;
					move-result-object v0
					if-eqz v0, :ret
					iput-object v0, p0, Lcom/faultexception/reader/book/EPubBook;->mSeries:Ljava/lang/String;
					:ret
					return-void
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`.method public getCreator()Ljava/lang/String;`),
				FixIndent("\n"+`
				.method public getSeries()Ljava/lang/String;
					.locals 1
					iget-object v0, p0, Lcom/faultexception/reader/book/EPubBook;->mSeries:Ljava/lang/String;
					return-object v0
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public getCreator()Ljava/lang/String;
				`),
				FixIndent("\n"+`
				.method public getSeriesIndex()Ljava/lang/String;
					.locals 1
					iget-object v0, p0, Lcom/faultexception/reader/book/EPubBook;->mSeriesIndex:Ljava/lang/String;
					return-object v0
				.end method
				`),
			),
		),
		PatchFile("smali/com/faultexception/reader/library/LibraryManager.smali",
			InMethod("scanBookInternal(Ljava/lang/String;ILjava/lang/String;J)Lcom/faultexception/reader/library/LibraryManager$ScanResult;",
				// must follow pattern of previous
				ReplaceStringAppend(
					FixIndent("\n"+`
						invoke-virtual {v0}, Lcom/faultexception/reader/book/Book;->getCreator()Ljava/lang/String;

						move-result-object p1

						const-string v5, "creator"

						invoke-virtual {v4, v5, p1}, Landroid/content/ContentValues;->put(Ljava/lang/String;Ljava/lang/String;)V
					`),
					FixIndent("\n"+`
						invoke-virtual {v0}, Lcom/faultexception/reader/book/Book;->getSeriesIndex()Ljava/lang/String;
						move-result-object p1
						const-string v5, "series_index"
						invoke-virtual {v4, v5, p1}, Landroid/content/ContentValues;->put(Ljava/lang/String;Ljava/lang/String;)V
					`),
				),
				// must follow pattern of previous
				ReplaceStringAppend(
					FixIndent("\n"+`
						invoke-virtual {v0}, Lcom/faultexception/reader/book/Book;->getCreator()Ljava/lang/String;

						move-result-object p1

						const-string v5, "creator"

						invoke-virtual {v4, v5, p1}, Landroid/content/ContentValues;->put(Ljava/lang/String;Ljava/lang/String;)V
					`),
					FixIndent("\n"+`
						invoke-virtual {v0}, Lcom/faultexception/reader/book/Book;->getSeries()Ljava/lang/String;
						move-result-object p1
						const-string v5, "series"
						invoke-virtual {v4, v5, p1}, Landroid/content/ContentValues;->put(Ljava/lang/String;Ljava/lang/String;)V
					`),
				),
			),
		),
	)
}
