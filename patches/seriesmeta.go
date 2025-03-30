// # Series metadata
//
// Add support for parsing and displaying EPUB3 or Calibre series metadata.
package patches

import (
	. "github.com/pgaskin/lithiumpatch/patches/patchdef"
)

// TODO: replace the "must follow pattern" ones with a regex to automate updating them, the others are already version-independent

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
				`"title LIKE ? OR (coalesce(creator, '') || coalesce(series, '')) LIKE ?"`, // a hack to not have to reorder the whole parameter array
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
				ReplaceStringPrepend(
					FixIndent("\n"+`
						iget-object v7, v1, Lcom/faultexception/reader/book/EPubBook;->mZip:Lcom/faultexception/reader/util/ZipFileCompat;

						invoke-virtual {v7, v4}, Lcom/faultexception/reader/util/ZipFileCompat;->getInputStream(Ljava/util/zip/ZipEntry;)Ljava/io/InputStream;

						move-result-object v4
						:try_end_0
					`),
					FixIndent("\n"+`
						invoke-direct {v1, v4}, Lcom/faultexception/reader/book/EPubBook;->parseSeries(Ljava/util/zip/ZipEntry;)V
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private readOpfFile(Ljava/lang/String;)V
				`),
				FixIndent("\n"+`
				.method private parseSeries(Ljava/util/zip/ZipEntry;)V
					.locals 1
					iget-object v0, p0, Lcom/faultexception/reader/book/EPubBook;->mZip:Lcom/faultexception/reader/util/ZipFileCompat;
					invoke-virtual {v0, p1}, Lcom/faultexception/reader/util/ZipFileCompat;->getInputStream(Ljava/util/zip/ZipEntry;)Ljava/io/InputStream;
					move-result-object v0
					invoke-direct {p0, v0}, Lcom/faultexception/reader/book/EPubBook;->parseSeries(Ljava/io/InputStream;)Ljava/lang/String;
					return-void
				.end method
				`),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method private readOpfFile(Ljava/lang/String;)V
				`),
				/*
					// $ANDROID_HOME/build-tools/30.0.3/dx --dex --output SeriesParser.{dex,class}
					// java -jar baksmali-2.5.2.jar d SeriesParser.dex

					import org.xmlpull.v1.XmlPullParser;
					import org.xmlpull.v1.XmlPullParserException;
					import org.xmlpull.v1.XmlPullParserFactory;

					import java.io.IOException;
					import java.io.InputStream;
					import java.nio.file.Files;
					import java.nio.file.Paths;
					import java.util.LinkedHashMap;
					import java.util.LinkedHashSet;

					public class SeriesParser {
						public static void main(String[] args) {
							try {
								final SeriesParser p = new SeriesParser();
								System.out.println(p.parseSeries(Files.newInputStream(Paths.get("package.opf"))));
								System.out.println(p.mSeries + " #" + p.mSeriesIndex);
							} catch (Exception ex) {
								throw new RuntimeException(ex);
							}
						}

						private String mSeries;
						private String mSeriesIndex;

						private String parseSeries(InputStream is) throws XmlPullParserException, IOException {
							final XmlPullParser xpp = XmlPullParserFactory.newInstance().newPullParser(); // Landroid/util/Xml;->newPullParser()Lorg/xmlpull/v1/XmlPullParser;
							xpp.setFeature("http://xmlpull.org/v1/doc/features.html#process-namespaces", true);
							xpp.setInput(is, null);
							final LinkedHashSet<String> hSeriesSkip = new LinkedHashSet<>();
							final LinkedHashMap<String, String> hSeries = new LinkedHashMap<>();
							final LinkedHashMap<String, String> hSeriesIndex = new LinkedHashMap<>();
							hSeries.put(null, null); // calibre series metadata first
							final StringBuilder txt = new StringBuilder();
							for (int depth = 0, depthMatch = 0, evt = xpp.getEventType(); evt != XmlPullParser.END_DOCUMENT; ) {
								switch (evt) {
									case XmlPullParser.END_TAG:
										if (depth-- < depthMatch) {
											depthMatch--;
										}
										break;
									case XmlPullParser.START_TAG:
										if (depth++ == depthMatch) {
											if ("http://www.idpf.org/2007/opf".equals(xpp.getNamespace())) {
												switch (depth) {
													case 1:
														if ("package".equals(xpp.getName()))
															depthMatch++;
														break;
													case 2:
														if ("metadata".equals(xpp.getName()))
															depthMatch++;
														break;
													case 3:
														if ("meta".equals(xpp.getName()))
															depthMatch++;
														break;
												}
											}
										}
										// if we're at a package>metadata>meta
										if (depthMatch == 3) {
											// get the attributes we want
											final String pName = xpp.getAttributeValue(null, "name");
											final String pContent = xpp.getAttributeValue(null, "content");
											final String pProperty = xpp.getAttributeValue(null, "property");
											final String pId = xpp.getAttributeValue(null, "id");
											final String pRefines = xpp.getAttributeValue(null, "refines");
											// get the text within the element
											txt.setLength(0);
											for (evt = xpp.next(); !(depth == 3 && evt == XmlPullParser.END_TAG); evt = xpp.next()) {
												switch (evt) {
													case XmlPullParser.START_TAG:
														depth++;
														break;
													case XmlPullParser.END_TAG:
														depth--;
														break;
													case XmlPullParser.TEXT:
														final String tmp = xpp.getText();
														if (tmp != null) {
															txt.append(tmp);
														}
														break;
												}
											}
											// get an identifier (null for calibre metadata) and key/value meta pair
											String vSrc, vKey, vValue;
											if (pName != null) {
												vSrc = null;
												vKey = pName;
												vValue = pContent;
											} else {
												if (pRefines != null && pRefines.startsWith("#")) {
													vSrc = pRefines.substring(1);
												} else if (pId != null) {
													vSrc = pId;
												} else {
													vSrc = "";
												}
												vKey = pProperty;
												vValue = txt.toString().trim();
												if (vValue.isEmpty()) {
													vValue = null;
												}
											}
											// if we have a key/value pair, process it
											if (vKey != null && vValue != null)
												if ("calibre:series".equals(vKey) || "belongs-to-collection".equals(vKey))
													hSeries.put(vSrc, vValue);
												else if ("calibre:series_index".equals(vKey) || "group-position".equals(vKey))
													hSeriesIndex.put(vSrc, vValue);
												else if ("collection-type".equals(vKey) && !"series".equals(vValue))
													hSeriesSkip.add(vSrc);
											continue; // we already consumed the next token (END_TAG) in the txt loop
										}
										break;
								}
								evt = xpp.next();
							}
							// get the first series
							for (final String src : hSeries.keySet()) {
								final String series = hSeries.get(src);
								if (series != null) {
									final String seriesIndex = hSeriesIndex.get(src);
									if (seriesIndex != null) {
										if (!hSeriesSkip.contains(src)) {
											mSeries = series; // Lcom/faultexception/reader/book/EPubBook;->mSeries:Ljava/lang/String;
											mSeriesIndex = seriesIndex; // Lcom/faultexception/reader/book/EPubBook;->mSeriesIndex:Ljava/lang/String;
											return src != null ? "#" + src : "calibre";
										}
									}
								}
							}
							return null;
						}
					}
				*/
				FixIndent("\n"+`
				.method private parseSeries(Ljava/io/InputStream;)Ljava/lang/String;
					.registers 28
					.param p1, "is"    # Ljava/io/InputStream;
					.annotation system Ldalvik/annotation/Throws;
						value = {
							Lorg/xmlpull/v1/XmlPullParserException;,
							Ljava/io/IOException;
						}
					.end annotation

					.prologue
					.line 30
					invoke-static {}, Landroid/util/Xml;->newPullParser()Lorg/xmlpull/v1/XmlPullParser;

					move-result-object v23

					.line 31
					.local v23, "xpp":Lorg/xmlpull/v1/XmlPullParser;
					const-string v24, "http://xmlpull.org/v1/doc/features.html#process-namespaces"

					const/16 v25, 0x1

					invoke-interface/range {v23 .. v25}, Lorg/xmlpull/v1/XmlPullParser;->setFeature(Ljava/lang/String;Z)V

					.line 32
					const/16 v24, 0x0

					move-object/from16 v0, v23

					move-object/from16 v1, p1

					move-object/from16 v2, v24

					invoke-interface {v0, v1, v2}, Lorg/xmlpull/v1/XmlPullParser;->setInput(Ljava/io/InputStream;Ljava/lang/String;)V

					.line 33
					new-instance v9, Ljava/util/LinkedHashSet;

					invoke-direct {v9}, Ljava/util/LinkedHashSet;-><init>()V

					.line 34
					.local v9, "hSeriesSkip":Ljava/util/LinkedHashSet;, "Ljava/util/LinkedHashSet<Ljava/lang/String;>;"
					new-instance v7, Ljava/util/LinkedHashMap;

					invoke-direct {v7}, Ljava/util/LinkedHashMap;-><init>()V

					.line 35
					.local v7, "hSeries":Ljava/util/LinkedHashMap;, "Ljava/util/LinkedHashMap<Ljava/lang/String;Ljava/lang/String;>;"
					new-instance v8, Ljava/util/LinkedHashMap;

					invoke-direct {v8}, Ljava/util/LinkedHashMap;-><init>()V

					.line 36
					.local v8, "hSeriesIndex":Ljava/util/LinkedHashMap;, "Ljava/util/LinkedHashMap<Ljava/lang/String;Ljava/lang/String;>;"
					const/16 v24, 0x0

					const/16 v25, 0x0

					move-object/from16 v0, v24

					move-object/from16 v1, v25

					invoke-virtual {v7, v0, v1}, Ljava/util/LinkedHashMap;->put(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;

					.line 37
					new-instance v19, Ljava/lang/StringBuilder;

					invoke-direct/range {v19 .. v19}, Ljava/lang/StringBuilder;-><init>()V

					.line 38
					.local v19, "txt":Ljava/lang/StringBuilder;
					const/4 v3, 0x0

					.local v3, "depth":I
					const/4 v5, 0x0

					.local v5, "depthMatch":I
					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->getEventType()I

					move-result v6

					.local v6, "evt":I
					move v4, v3

					.end local v3    # "depth":I
					.local v4, "depth":I
					:goto_40
					const/16 v24, 0x1

					move/from16 v0, v24

					if-eq v6, v0, :cond_199

					.line 39
					packed-switch v6, :pswitch_data_1f6

					move v3, v4

					.line 122
					.end local v4    # "depth":I
					.restart local v3    # "depth":I
					:cond_4a
					:goto_4a
					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->next()I

					move-result v6

					move v4, v3

					.end local v3    # "depth":I
					.restart local v4    # "depth":I
					goto :goto_40

					.line 41
					:pswitch_50
					add-int/lit8 v3, v4, -0x1

					.end local v4    # "depth":I
					.restart local v3    # "depth":I
					if-ge v4, v5, :cond_4a

					.line 42
					add-int/lit8 v5, v5, -0x1

					goto :goto_4a

					.line 46
					.end local v3    # "depth":I
					.restart local v4    # "depth":I
					:pswitch_57
					add-int/lit8 v3, v4, 0x1

					.end local v4    # "depth":I
					.restart local v3    # "depth":I
					if-ne v4, v5, :cond_6a

					.line 47
					const-string v24, "http://www.idpf.org/2007/opf"

					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->getNamespace()Ljava/lang/String;

					move-result-object v25

					invoke-virtual/range {v24 .. v25}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_6a

					.line 48
					packed-switch v3, :pswitch_data_1fe

					.line 65
					:cond_6a
					:goto_6a
					const/16 v24, 0x3

					move/from16 v0, v24

					if-ne v5, v0, :cond_4a

					.line 67
					const/16 v24, 0x0

					const-string v25, "name"

					invoke-interface/range {v23 .. v25}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;

					move-result-object v12

					.line 68
					.local v12, "pName":Ljava/lang/String;
					const/16 v24, 0x0

					const-string v25, "content"

					invoke-interface/range {v23 .. v25}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;

					move-result-object v10

					.line 69
					.local v10, "pContent":Ljava/lang/String;
					const/16 v24, 0x0

					const-string v25, "property"

					invoke-interface/range {v23 .. v25}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;

					move-result-object v13

					.line 70
					.local v13, "pProperty":Ljava/lang/String;
					const/16 v24, 0x0

					const-string v25, "id"

					invoke-interface/range {v23 .. v25}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;

					move-result-object v11

					.line 71
					.local v11, "pId":Ljava/lang/String;
					const/16 v24, 0x0

					const-string v25, "refines"

					invoke-interface/range {v23 .. v25}, Lorg/xmlpull/v1/XmlPullParser;->getAttributeValue(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;

					move-result-object v14

					.line 73
					.local v14, "pRefines":Ljava/lang/String;
					const/16 v24, 0x0

					move-object/from16 v0, v19

					move/from16 v1, v24

					invoke-virtual {v0, v1}, Ljava/lang/StringBuilder;->setLength(I)V

					.line 74
					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->next()I

					move-result v6

					:goto_a5
					const/16 v24, 0x3

					move/from16 v0, v24

					if-ne v3, v0, :cond_b1

					const/16 v24, 0x3

					move/from16 v0, v24

					if-eq v6, v0, :cond_fa

					.line 75
					:cond_b1
					packed-switch v6, :pswitch_data_208

					.line 74
					:cond_b4
					:goto_b4
					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->next()I

					move-result v6

					goto :goto_a5

					.line 50
					.end local v10    # "pContent":Ljava/lang/String;
					.end local v11    # "pId":Ljava/lang/String;
					.end local v12    # "pName":Ljava/lang/String;
					.end local v13    # "pProperty":Ljava/lang/String;
					.end local v14    # "pRefines":Ljava/lang/String;
					:pswitch_b9
					const-string v24, "package"

					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->getName()Ljava/lang/String;

					move-result-object v25

					invoke-virtual/range {v24 .. v25}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_6a

					.line 51
					add-int/lit8 v5, v5, 0x1

					goto :goto_6a

					.line 54
					:pswitch_c8
					const-string v24, "metadata"

					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->getName()Ljava/lang/String;

					move-result-object v25

					invoke-virtual/range {v24 .. v25}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_6a

					.line 55
					add-int/lit8 v5, v5, 0x1

					goto :goto_6a

					.line 58
					:pswitch_d7
					const-string v24, "meta"

					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->getName()Ljava/lang/String;

					move-result-object v25

					invoke-virtual/range {v24 .. v25}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_6a

					.line 59
					add-int/lit8 v5, v5, 0x1

					goto :goto_6a

					.line 77
					.restart local v10    # "pContent":Ljava/lang/String;
					.restart local v11    # "pId":Ljava/lang/String;
					.restart local v12    # "pName":Ljava/lang/String;
					.restart local v13    # "pProperty":Ljava/lang/String;
					.restart local v14    # "pRefines":Ljava/lang/String;
					:pswitch_e6
					add-int/lit8 v3, v3, 0x1

					.line 78
					goto :goto_b4

					.line 80
					:pswitch_e9
					add-int/lit8 v3, v3, -0x1

					.line 81
					goto :goto_b4

					.line 83
					:pswitch_ec
					invoke-interface/range {v23 .. v23}, Lorg/xmlpull/v1/XmlPullParser;->getText()Ljava/lang/String;

					move-result-object v18

					.line 84
					.local v18, "tmp":Ljava/lang/String;
					if-eqz v18, :cond_b4

					.line 85
					move-object/from16 v0, v19

					move-object/from16 v1, v18

					invoke-virtual {v0, v1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					goto :goto_b4

					.line 92
					.end local v18    # "tmp":Ljava/lang/String;
					:cond_fa
					if-eqz v12, :cond_128

					.line 93
					const/16 v21, 0x0

					.line 94
					.local v21, "vSrc":Ljava/lang/String;
					move-object/from16 v20, v12

					.line 95
					.local v20, "vKey":Ljava/lang/String;
					move-object/from16 v22, v10

					.line 111
					.local v22, "vValue":Ljava/lang/String;
					:cond_102
					:goto_102
					if-eqz v20, :cond_1f3

					if-eqz v22, :cond_1f3

					.line 112
					const-string v24, "calibre:series"

					move-object/from16 v0, v24

					move-object/from16 v1, v20

					invoke-virtual {v0, v1}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-nez v24, :cond_11e

					const-string v24, "belongs-to-collection"

					move-object/from16 v0, v24

					move-object/from16 v1, v20

					invoke-virtual {v0, v1}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_157

					.line 113
					:cond_11e
					move-object/from16 v0, v21

					move-object/from16 v1, v22

					invoke-virtual {v7, v0, v1}, Ljava/util/LinkedHashMap;->put(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;

					move v4, v3

					.end local v3    # "depth":I
					.restart local v4    # "depth":I
					goto/16 :goto_40

					.line 97
					.end local v4    # "depth":I
					.end local v20    # "vKey":Ljava/lang/String;
					.end local v21    # "vSrc":Ljava/lang/String;
					.end local v22    # "vValue":Ljava/lang/String;
					.restart local v3    # "depth":I
					:cond_128
					if-eqz v14, :cond_14f

					const-string v24, "#"

					move-object/from16 v0, v24

					invoke-virtual {v14, v0}, Ljava/lang/String;->startsWith(Ljava/lang/String;)Z

					move-result v24

					if-eqz v24, :cond_14f

					.line 98
					const/16 v24, 0x1

					move/from16 v0, v24

					invoke-virtual {v14, v0}, Ljava/lang/String;->substring(I)Ljava/lang/String;

					move-result-object v21

					.line 104
					.restart local v21    # "vSrc":Ljava/lang/String;
					:goto_13c
					move-object/from16 v20, v13

					.line 105
					.restart local v20    # "vKey":Ljava/lang/String;
					invoke-virtual/range {v19 .. v19}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;

					move-result-object v24

					invoke-virtual/range {v24 .. v24}, Ljava/lang/String;->trim()Ljava/lang/String;

					move-result-object v22

					.line 106
					.restart local v22    # "vValue":Ljava/lang/String;
					invoke-virtual/range {v22 .. v22}, Ljava/lang/String;->isEmpty()Z

					move-result v24

					if-eqz v24, :cond_102

					.line 107
					const/16 v22, 0x0

					goto :goto_102

					.line 99
					.end local v20    # "vKey":Ljava/lang/String;
					.end local v21    # "vSrc":Ljava/lang/String;
					.end local v22    # "vValue":Ljava/lang/String;
					:cond_14f
					if-eqz v11, :cond_154

					.line 100
					move-object/from16 v21, v11

					.restart local v21    # "vSrc":Ljava/lang/String;
					goto :goto_13c

					.line 102
					.end local v21    # "vSrc":Ljava/lang/String;
					:cond_154
					const-string v21, ""

					.restart local v21    # "vSrc":Ljava/lang/String;
					goto :goto_13c

					.line 114
					.restart local v20    # "vKey":Ljava/lang/String;
					.restart local v22    # "vValue":Ljava/lang/String;
					:cond_157
					const-string v24, "calibre:series_index"

					move-object/from16 v0, v24

					move-object/from16 v1, v20

					invoke-virtual {v0, v1}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-nez v24, :cond_16f

					const-string v24, "group-position"

					move-object/from16 v0, v24

					move-object/from16 v1, v20

					invoke-virtual {v0, v1}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_179

					.line 115
					:cond_16f
					move-object/from16 v0, v21

					move-object/from16 v1, v22

					invoke-virtual {v8, v0, v1}, Ljava/util/LinkedHashMap;->put(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;

					move v4, v3

					.end local v3    # "depth":I
					.restart local v4    # "depth":I
					goto/16 :goto_40

					.line 116
					.end local v4    # "depth":I
					.restart local v3    # "depth":I
					:cond_179
					const-string v24, "collection-type"

					move-object/from16 v0, v24

					move-object/from16 v1, v20

					invoke-virtual {v0, v1}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-eqz v24, :cond_1f3

					const-string v24, "series"

					move-object/from16 v0, v24

					move-object/from16 v1, v22

					invoke-virtual {v0, v1}, Ljava/lang/String;->equals(Ljava/lang/Object;)Z

					move-result v24

					if-nez v24, :cond_1f3

					.line 117
					move-object/from16 v0, v21

					invoke-virtual {v9, v0}, Ljava/util/LinkedHashSet;->add(Ljava/lang/Object;)Z

					move v4, v3

					.end local v3    # "depth":I
					.restart local v4    # "depth":I
					goto/16 :goto_40

					.line 125
					.end local v10    # "pContent":Ljava/lang/String;
					.end local v11    # "pId":Ljava/lang/String;
					.end local v12    # "pName":Ljava/lang/String;
					.end local v13    # "pProperty":Ljava/lang/String;
					.end local v14    # "pRefines":Ljava/lang/String;
					.end local v20    # "vKey":Ljava/lang/String;
					.end local v21    # "vSrc":Ljava/lang/String;
					.end local v22    # "vValue":Ljava/lang/String;
					:cond_199
					invoke-virtual {v7}, Ljava/util/LinkedHashMap;->keySet()Ljava/util/Set;

					move-result-object v24

					invoke-interface/range {v24 .. v24}, Ljava/util/Set;->iterator()Ljava/util/Iterator;

					move-result-object v24

					:cond_1a1
					invoke-interface/range {v24 .. v24}, Ljava/util/Iterator;->hasNext()Z

					move-result v25

					if-eqz v25, :cond_1f0

					invoke-interface/range {v24 .. v24}, Ljava/util/Iterator;->next()Ljava/lang/Object;

					move-result-object v17

					check-cast v17, Ljava/lang/String;

					.line 126
					.local v17, "src":Ljava/lang/String;
					move-object/from16 v0, v17

					invoke-virtual {v7, v0}, Ljava/util/LinkedHashMap;->get(Ljava/lang/Object;)Ljava/lang/Object;

					move-result-object v15

					check-cast v15, Ljava/lang/String;

					.line 127
					.local v15, "series":Ljava/lang/String;
					if-eqz v15, :cond_1a1

					.line 128
					move-object/from16 v0, v17

					invoke-virtual {v8, v0}, Ljava/util/LinkedHashMap;->get(Ljava/lang/Object;)Ljava/lang/Object;

					move-result-object v16

					check-cast v16, Ljava/lang/String;

					.line 129
					.local v16, "seriesIndex":Ljava/lang/String;
					if-eqz v16, :cond_1a1

					.line 130
					move-object/from16 v0, v17

					invoke-virtual {v9, v0}, Ljava/util/LinkedHashSet;->contains(Ljava/lang/Object;)Z

					move-result v25

					if-nez v25, :cond_1a1

					.line 131
					move-object/from16 v0, p0

					iput-object v15, v0, Lcom/faultexception/reader/book/EPubBook;->mSeries:Ljava/lang/String;

					.line 132
					move-object/from16 v0, v16

					move-object/from16 v1, p0

					iput-object v0, v1, Lcom/faultexception/reader/book/EPubBook;->mSeriesIndex:Ljava/lang/String;

					.line 133
					if-eqz v17, :cond_1ed

					new-instance v24, Ljava/lang/StringBuilder;

					invoke-direct/range {v24 .. v24}, Ljava/lang/StringBuilder;-><init>()V

					const-string v25, "#"

					invoke-virtual/range {v24 .. v25}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					move-result-object v24

					move-object/from16 v0, v24

					move-object/from16 v1, v17

					invoke-virtual {v0, v1}, Ljava/lang/StringBuilder;->append(Ljava/lang/String;)Ljava/lang/StringBuilder;

					move-result-object v24

					invoke-virtual/range {v24 .. v24}, Ljava/lang/StringBuilder;->toString()Ljava/lang/String;

					move-result-object v24

					.line 138
					.end local v15    # "series":Ljava/lang/String;
					.end local v16    # "seriesIndex":Ljava/lang/String;
					.end local v17    # "src":Ljava/lang/String;
					:goto_1ec
					return-object v24

					.line 133
					.restart local v15    # "series":Ljava/lang/String;
					.restart local v16    # "seriesIndex":Ljava/lang/String;
					.restart local v17    # "src":Ljava/lang/String;
					:cond_1ed
					const-string v24, "calibre"

					goto :goto_1ec

					.line 138
					.end local v15    # "series":Ljava/lang/String;
					.end local v16    # "seriesIndex":Ljava/lang/String;
					.end local v17    # "src":Ljava/lang/String;
					:cond_1f0
					const/16 v24, 0x0

					goto :goto_1ec

					.end local v4    # "depth":I
					.restart local v3    # "depth":I
					.restart local v10    # "pContent":Ljava/lang/String;
					.restart local v11    # "pId":Ljava/lang/String;
					.restart local v12    # "pName":Ljava/lang/String;
					.restart local v13    # "pProperty":Ljava/lang/String;
					.restart local v14    # "pRefines":Ljava/lang/String;
					.restart local v20    # "vKey":Ljava/lang/String;
					.restart local v21    # "vSrc":Ljava/lang/String;
					.restart local v22    # "vValue":Ljava/lang/String;
					:cond_1f3
					move v4, v3

					.end local v3    # "depth":I
					.restart local v4    # "depth":I
					goto/16 :goto_40

					.line 39
					:pswitch_data_1f6
					.packed-switch 0x2
						:pswitch_57
						:pswitch_50
					.end packed-switch

					.line 48
					:pswitch_data_1fe
					.packed-switch 0x1
						:pswitch_b9
						:pswitch_c8
						:pswitch_d7
					.end packed-switch

					.line 75
					:pswitch_data_208
					.packed-switch 0x2
						:pswitch_e6
						:pswitch_e9
						:pswitch_ec
					.end packed-switch
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
	seriesdrawer() // must be after seriesmeta
}
