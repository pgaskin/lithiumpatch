// # Series in main drawer
//
// Add a series section to the main drawer.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func seriesdrawer() {
	const filterId = "3" // one more than loadFolder
	Register("seriesdrawer",
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_advanced">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Show series in drawer" android:key="series_in_drawer" android:defaultValue="false" />`,
			),
		),

		// based on empty_category
		WriteFileString("res/layout/empty_series.xml", FixIndent(`
		<?xml version="1.0" encoding="utf-8"?>
		<LinearLayout android:gravity="center" android:orientation="vertical" android:padding="20.0dip" android:clipToPadding="false" android:layout_width="fill_parent" android:layout_height="wrap_content" android:layout_marginBottom="?actionBarSize" xmlns:android="http://schemas.android.com/apk/res/android" xmlns:app="http://schemas.android.com/apk/res-auto">
			<ImageView android:id="@id/empty_image" android:layout_width="wrap_content" android:layout_height="wrap_content" android:src="@drawable/ic_library_empty" android:contentDescription="@null" app:tint="?android:textColorPrimary" />
			<TextView android:textSize="20.0sp" android:textColor="?android:textColorPrimary" android:gravity="center" android:id="@id/empty_title" android:layout_width="300.0dip" android:layout_height="wrap_content" android:text="No books in series" android:fontFamily="sans-serif-condensed" />
			<Button android:id="@id/empty_goto_all" android:paddingLeft="24.0dip" android:paddingRight="24.0dip" android:layout_width="wrap_content" android:layout_height="wrap_content" android:layout_marginTop="24.0dip" android:text="@string/no_books_goto_all" />
		</LinearLayout>
		`)),
		DefineR("smali/com/faultexception/reader", "layout", "empty_series"),

		PatchFile("smali/com/faultexception/reader/BooksFragment.smali",
			ReplaceStringAppend(
				"\n"+`.field public static final EXTRA_FOLDER:Ljava/lang/String; = "folder"`,
				"\n"+`.field public static final EXTRA_SERIES:Ljava/lang/String; = "series"`,
			),

			InMethod(`getFilter()I`,
				MustContain(FixIndent("\n"+`
					:cond_1
					const-string v2, "folder"
				`)),
				ReplaceStringAppend(
					FixIndent("\n"+`
						:cond_1
					`),
					// else if (getArguments().containsKey(EXTRA_SERIES)) return 3
					FixIndent("\n"+`
						sget-object v2, Lcom/faultexception/reader/BooksFragment;->EXTRA_SERIES:Ljava/lang/String;
						invoke-virtual {v0, v2}, Landroid/os/Bundle;->containsKey(Ljava/lang/String;)Z
						move-result v2
						if-eqz v2, :cond_1a

						const/4 v2, `+filterId+`
						return v2

						:cond_1a
					`),
				),
			),

			MustContain(`.method private loadFolder(Ljava/lang/String;)V`),
			ReplaceStringPrepend(
				FixIndent("\n"+`.method private loadNone()V`),
				FixIndent("\n"+`
				.method private loadSeries(Ljava/lang/String;)V
					.locals 1

					const/4 v0, `+filterId+`
					iput v0, p0, Lcom/faultexception/reader/BooksFragment;->mFilter:I
					iput-object p1, p0, Lcom/faultexception/reader/BooksFragment;->mTitle:Ljava/lang/String;

					return-void
				.end method
				`),
			),

			InMethod(`onCreate(Landroid/os/Bundle;)V`,
				MustContain(FixIndent("\n"+`
					goto :goto_0

					:cond_0
					const-string v0, "folder"
				`)),
				MustContain(FixIndent("\n"+`
					invoke-virtual {p1, v0}, Landroid/os/Bundle;->containsKey(Ljava/lang/String;)Z

					move-result v1

					if-eqz v1, :cond_1
				`)),
				MustContain(FixIndent("\n"+`
					:cond_1
					invoke-direct {p0}, Lcom/faultexception/reader/BooksFragment;->loadNone()V
				`)),
				ReplaceStringAppend(
					FixIndent("\n"+`
						:cond_1
					`),
					// else if (getArguments().containsKey(EXTRA_SERIES)) loadFolder(getArguments().getString(EXTRA_SERIES))
					FixIndent("\n"+`
						sget-object v0, Lcom/faultexception/reader/BooksFragment;->EXTRA_SERIES:Ljava/lang/String;
						invoke-virtual {p1, v0}, Landroid/os/Bundle;->containsKey(Ljava/lang/String;)Z
						move-result v1
						if-eqz v1, :cond_1a

						invoke-virtual {p1, v0}, Landroid/os/Bundle;->getString(Ljava/lang/String;)Ljava/lang/String;
						move-result-object p1
						invoke-direct {p0, p1}, Lcom/faultexception/reader/BooksFragment;->loadSeries(Ljava/lang/String;)V
						goto :goto_0

						:cond_1a
					`),
				),
			),

			MustContain(`.method public static newFolderInstance(Ljava/lang/String;)Landroidx/fragment/app/Fragment;`),
			ReplaceStringPrepend(
				FixIndent("\n"+`.method public static newInstance()Landroidx/fragment/app/Fragment;`),
				FixIndent("\n"+`
				.method public static newSeriesInstance(Ljava/lang/String;)Landroidx/fragment/app/Fragment;
					.locals 3

					new-instance v0, Lcom/faultexception/reader/BooksFragment;
					invoke-direct {v0}, Lcom/faultexception/reader/BooksFragment;-><init>()V

					new-instance v1, Landroid/os/Bundle;
					invoke-direct {v1}, Landroid/os/Bundle;-><init>()V

					sget-object v2, Lcom/faultexception/reader/BooksFragment;->EXTRA_SERIES:Ljava/lang/String;
					invoke-virtual {v1, v2, p0}, Landroid/os/Bundle;->putString(Ljava/lang/String;Ljava/lang/String;)V
					invoke-virtual {v0, v1}, Lcom/faultexception/reader/BooksFragment;->setArguments(Landroid/os/Bundle;)V

					return-object v0
				.end method
				`),
			),

			InMethod(`equalsFragment(Landroidx/fragment/app/Fragment;)Z`,
				MustContain(FixIndent("\n"+`
					invoke-virtual {p1}, Lcom/faultexception/reader/BooksFragment;->getArguments()Landroid/os/Bundle;

					move-result-object v0
				`)),
				MustContain(FixIndent("\n"+`
					invoke-virtual {p0}, Lcom/faultexception/reader/BooksFragment;->getArguments()Landroid/os/Bundle;

					move-result-object v1
				`)),
				ReplaceStringAppend(
					FixIndent("\n"+`
						:cond_0
						invoke-virtual {p1}, Lcom/faultexception/reader/BooksFragment;->getFilter()I

						move-result p1

						invoke-virtual {p0}, Lcom/faultexception/reader/BooksFragment;->getFilter()I

						move-result v4

						if-ne p1, v4, :cond_1
					`),
					// && arguments.getString(EXTRA_SERIES).equals(arguments2.getString(EXTRA_SERIES))
					FixIndent("\n"+`
						sget-object p1, Lcom/faultexception/reader/BooksFragment;->EXTRA_SERIES:Ljava/lang/String;
						invoke-virtual {v0, p1}, Landroid/os/Bundle;->getString(Ljava/lang/String;)Ljava/lang/String;
						move-result-object v4
						invoke-virtual {v1, p1}, Landroid/os/Bundle;->getString(Ljava/lang/String;)Ljava/lang/String;
						move-result-object p1
						invoke-static {v4, p1}, Landroid/text/TextUtils;->equals(Ljava/lang/CharSequence;Ljava/lang/CharSequence;)Z
						move-result p1
						if-eqz p1, :cond_1
					`),
				),
			),

			InMethod("runQuery(Z)V",
				ReplaceString(
					".locals 23\n",
					".locals 24\n",
				),
				ReplaceString(
					// public Cursor query (String table, String[] columns, String selection, String[] selectionArgs, String groupBy, String having, String orderBy, String limit)
					// v15.query(v16, v17, v18, v19, v20, v21, v22)
					FixIndent("\n"+`
						invoke-virtual/range {v15 .. v22}, Landroid/database/sqlite/SQLiteDatabase;->query(Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Landroid/database/Cursor;
					`),
					// this is slightly less efficient, but much saner than trying to preserve the existing control flow and registers while patching in our stuff
					FixIndent("\n"+`
						move-object/from16 v23, p0
						invoke-static/range {v15 .. v23}, Lcom/faultexception/reader/BooksFragment;->runQuery_seriesHelper(Landroid/database/sqlite/SQLiteDatabase;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Lcom/faultexception/reader/BooksFragment;)Landroid/database/Cursor;
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public runQuery(Z)V
				`),
				FixIndent("\n"+`
				.method private static runQuery_seriesHelper(Landroid/database/sqlite/SQLiteDatabase;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Lcom/faultexception/reader/BooksFragment;)Landroid/database/Cursor;
					.locals 3

					const v0, `+filterId+`
					iget v1, p8, Lcom/faultexception/reader/BooksFragment;->mFilter:I
					if-ne v0, v1, :done

					const-string v0, "series=?"
					invoke-static {p3, v0}, Landroid/database/DatabaseUtils;->concatenateWhere(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
					move-result-object p3

					const v1, 1
					new-array v0, v1, [Ljava/lang/String;
					iget-object v2, p8, Lcom/faultexception/reader/BooksFragment;->mTitle:Ljava/lang/String;
					const v1, 0
					aput-object v2, v0, v1
					invoke-static {p4, v0}, Landroid/database/DatabaseUtils;->appendSelectionArgs([Ljava/lang/String;[Ljava/lang/String;)[Ljava/lang/String;
					move-result-object p4

					:done
					invoke-virtual/range {p0 .. p7}, Landroid/database/sqlite/SQLiteDatabase;->query(Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;[Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Landroid/database/Cursor;
					move-result-object p0
					return-object p0
				.end method
				`),
			),

			InMethod("inflateEmptyView(Landroid/view/LayoutInflater;Landroid/view/ViewGroup;)V",
				MustContain(FixIndent("\n"+`
					iget v0, p0, Lcom/faultexception/reader/BooksFragment;->mFilter:I
				`)),
				ReplaceStringAppend(
					FixIndent("\n"+`
						if-eq v0, v1, :cond_1

						const/4 v1, 0x2

						if-eq v0, v1, :cond_0
					`),
					FixIndent("\n"+`
						const v1, `+filterId+`

						if-eq v0, v1, :cond_0a
					`),
				),
				MustContain(FixIndent("\n"+`
					goto :goto_0

					:cond_1
				`)),
				ReplaceStringPrepend(
					FixIndent("\n"+`
						:cond_0
					`),
					FixIndent("\n"+`
						:cond_0a

						sget v0, Lcom/faultexception/reader/R$layout;->empty_series:I
						invoke-virtual {p1, v0, p2}, Landroid/view/LayoutInflater;->inflate(ILandroid/view/ViewGroup;)Landroid/view/View;
						move-result-object p1

						sget v2, Lcom/faultexception/reader/R$id;->empty_goto_all:I
						invoke-virtual {p1, v2}, Landroid/view/View;->findViewById(I)Landroid/view/View;
						move-result-object p1

						check-cast p1, Landroid/widget/Button;
						iput-object p1, p0, Lcom/faultexception/reader/BooksFragment;->mEmptyGotoAllButton:Landroid/widget/Button;

						invoke-virtual {p1, p0}, Landroid/widget/Button;->setOnClickListener(Landroid/view/View$OnClickListener;)V

						goto :goto_0
					`),
				),
			),
		),

		// like ic_drawer_label, but with material collections_bookmark
		WriteFileString("res/drawable-anydpi-v21/ic_drawer_collection.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<vector xmlns:android="http://schemas.android.com/apk/res/android" android:height="24dp" android:width="24dp" android:viewportWidth="24" android:viewportHeight="24">
				<path android:fillColor="#ff000000" android:pathData="M4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6z" />
				<path android:fillColor="#ff000000" android:pathData="M20 2H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 10l-2.5-1.5L15 12V4h5v8z" />
			</vector>
			`),
		),
		DefineR("smali/com/faultexception/reader", "drawable", "ic_drawer_collection"),

		// based on FoldersAdapter, CategoriesAdapter
		WriteFileString("smali/com/faultexception/reader/MainDrawerFragment$SeriesAdapter.smali", FixIndent(`
		.class Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;
		.super Landroid/widget/CursorAdapter;

		.annotation system Ldalvik/annotation/EnclosingClass;
			value = Lcom/faultexception/reader/MainDrawerFragment;
		.end annotation

		.annotation system Ldalvik/annotation/InnerClass;
			accessFlags = 0x2
			name = "SeriesAdapter"
		.end annotation

		.field final synthetic this$0:Lcom/faultexception/reader/MainDrawerFragment;

		.method public constructor <init>(Lcom/faultexception/reader/MainDrawerFragment;Landroid/content/Context;)V
			.locals 1

			iput-object p1, p0, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;->this$0:Lcom/faultexception/reader/MainDrawerFragment;
			const/4 p1, 0x0
			const/4 v0, 0x0
			invoke-direct {p0, p2, p1, v0}, Landroid/widget/CursorAdapter;-><init>(Landroid/content/Context;Landroid/database/Cursor;I)V

			return-void
		.end method

		.method public bindView(Landroid/view/View;Landroid/content/Context;Landroid/database/Cursor;)V
			.locals 5

			const v0, 1
			invoke-interface {p3, v0}, Landroid/database/Cursor;->getString(I)Ljava/lang/String;
			move-result-object v0

			iget-object v1, p0, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;->this$0:Lcom/faultexception/reader/MainDrawerFragment;
			invoke-static {v1}, Lcom/faultexception/reader/MainDrawerFragment;->access$300(Lcom/faultexception/reader/MainDrawerFragment;)Landroidx/fragment/app/Fragment;
			move-result-object v1
			instance-of v2, v1, Lcom/faultexception/reader/BooksFragment;
			if-eqz v2, :nomatch

			check-cast v1, Lcom/faultexception/reader/BooksFragment;
			invoke-virtual {v1}, Lcom/faultexception/reader/BooksFragment;->getFilter()I
			move-result v2
			const v3, `+filterId+`
			if-ne v3, v2, :nomatch

			invoke-virtual {v1}, Lcom/faultexception/reader/BooksFragment;->getArguments()Landroid/os/Bundle;
			move-result-object v1
			sget-object v2, Lcom/faultexception/reader/BooksFragment;->EXTRA_SERIES:Ljava/lang/String;
			invoke-virtual {v1, v2}, Landroid/os/Bundle;->getString(Ljava/lang/String;)Ljava/lang/String;
			move-result-object v2
			invoke-static {v0, v2}, Landroid/text/TextUtils;->equals(Ljava/lang/CharSequence;Ljava/lang/CharSequence;)Z
			move-result v2
			if-eqz v2, :nomatch

			const v1, 1
			goto :match
			:nomatch
			const v1, 0
			:match

			sget v2, Lcom/faultexception/reader/R$drawable;->ic_drawer_collection:I
			invoke-static {p2, p1, v2, v0, v1}, Lcom/faultexception/reader/util/DrawerAdapterHelper;->bindView(Landroid/content/Context;Landroid/view/View;ILjava/lang/String;Z)V

			return-void
		.end method

		.method public newView(Landroid/content/Context;Landroid/database/Cursor;Landroid/view/ViewGroup;)Landroid/view/View;
			.locals 0
			invoke-static {p1, p3}, Lcom/faultexception/reader/util/DrawerAdapterHelper;->inflateView(Landroid/content/Context;Landroid/view/ViewGroup;)Landroid/view/View;
			move-result-object p1
			return-object p1
		.end method

		.method public getSeriesAtIndex(I)Ljava/lang/String;
			.locals 1

			invoke-virtual {p0}, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;->getCursor()Landroid/database/Cursor;
			move-result-object v0

			invoke-interface {v0, p1}, Landroid/database/Cursor;->moveToPosition(I)Z
			move-result p1

			if-eqz p1, :null

			const p1, 1
			invoke-interface {v0, p1}, Landroid/database/Cursor;->getString(I)Ljava/lang/String;
			move-result-object p1
			return-object p1

			:null
			const p1, 0x0
			return-object p1
		.end method
		`)),

		// based on folders_list_header
		WriteFileString("res/layout/series_list_header.xml", FixIndent(`
		<?xml version="1.0" encoding="utf-8"?>
		<LinearLayout xmlns:android="http://schemas.android.com/apk/res/android" android:orientation="vertical" android:layout_width="match_parent" android:layout_height="match_parent">
			<View android:background="@color/divider" android:layout_width="match_parent" android:layout_height="1dp" android:layout_marginTop="8dp"/>
			<TextView android:textSize="14sp" android:textColor="?android:attr/textColorSecondary" android:gravity="center_vertical" android:paddingLeft="16dp" android:paddingRight="16dp" android:layout_width="match_parent" android:layout_height="48dp" android:text="Series" android:fontFamily="sans-serif-medium"/>
		</LinearLayout>
		`)),
		DefineR("smali/com/faultexception/reader", "layout", "series_list_header"),

		PatchFile("smali/com/faultexception/reader/MainDrawerFragment.smali",
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.field private mFolderSection:Lcom/faultexception/reader/util/adapters/MultiAdapter;

				.field private mFoldersAdapter:Lcom/faultexception/reader/MainDrawerFragment$FoldersAdapter;
				`),
				FixIndent("\n"+`
				.field private mSeriesSection:Lcom/faultexception/reader/util/adapters/MultiAdapter;

				.field private mSeriesAdapter:Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;
				`),
			),

			InMethod("onCreateView(Landroid/view/LayoutInflater;Landroid/view/ViewGroup;Landroid/os/Bundle;)Landroid/view/View;",
				// before folders section
				ReplaceStringAppend(
					FixIndent("\n"+`
						iput-object p3, p0, Lcom/faultexception/reader/MainDrawerFragment;->mFolderSection:Lcom/faultexception/reader/util/adapters/MultiAdapter;
					`),
					FixIndent("\n"+`
						invoke-direct {p0}, Lcom/faultexception/reader/MainDrawerFragment;->onCreateViewSeries()V
					`),
				),
			),
			// based on part of onCreateView()
			ReplaceStringPrepend(
				FixIndent("\n"+`.method public onCreateView(Landroid/view/LayoutInflater;Landroid/view/ViewGroup;Landroid/os/Bundle;)Landroid/view/View;`),
				FixIndent("\n"+`
				.method private onCreateViewSeries()V
					.locals 4

					invoke-virtual {p0}, Lcom/faultexception/reader/MainDrawerFragment;->requireContext()Landroid/content/Context;
					move-result-object v0

					new-instance v1, Lcom/faultexception/reader/util/adapters/MultiAdapter;
					invoke-direct {v1}, Lcom/faultexception/reader/util/adapters/MultiAdapter;-><init>()V
					iput-object v1, p0, Lcom/faultexception/reader/MainDrawerFragment;->mSeriesSection:Lcom/faultexception/reader/util/adapters/MultiAdapter;

					new-instance v2, Lcom/faultexception/reader/util/adapters/SingleViewAdapter;
					sget v3, Lcom/faultexception/reader/R$layout;->series_list_header:I
					invoke-direct {v2, v0, v3}, Lcom/faultexception/reader/util/adapters/SingleViewAdapter;-><init>(Landroid/content/Context;I)V
					invoke-virtual {v1, v2}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->addAdapter(Landroid/widget/BaseAdapter;)V

					new-instance v2, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;
					invoke-direct {v2, p0, v0}, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;-><init>(Lcom/faultexception/reader/MainDrawerFragment;Landroid/content/Context;)V
					invoke-virtual {v1, v2}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->addAdapter(Landroid/widget/BaseAdapter;)V
					iput-object v2, p0, Lcom/faultexception/reader/MainDrawerFragment;->mSeriesAdapter:Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;

					iget-object v2, p0, Lcom/faultexception/reader/MainDrawerFragment;->mAdapter:Lcom/faultexception/reader/util/adapters/MultiAdapter;
					invoke-virtual {v2, v1}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->addAdapter(Landroid/widget/BaseAdapter;)V

					return-void
				.end method
				`),
			),

			InMethod("refresh()V",
				ReplaceStringPrepend(
					FixIndent("\n"+`
						invoke-virtual {v0}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->notifyDataSetChanged()V
					`),
					FixIndent("\n"+`
						invoke-direct {p0}, Lcom/faultexception/reader/MainDrawerFragment;->refreshSeries()V
					`),
				),
			),
			// based on part of refresh()
			ReplaceStringPrepend(
				FixIndent("\n"+`.method public refresh()V`),
				FixIndent("\n"+`
				.method private refreshSeries()V
					.locals 3

					invoke-virtual {p0}, Lcom/faultexception/reader/MainDrawerFragment;->isAdded()Z
					move-result v0
					if-eqz v0, :done

					invoke-virtual {p0}, Lcom/faultexception/reader/MainDrawerFragment;->getActivity()Landroidx/fragment/app/FragmentActivity;
					move-result-object v0
					invoke-static {v0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0
					const-string v1, "series_in_drawer"
					const v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v2

					iget-object v0, p0, Lcom/faultexception/reader/MainDrawerFragment;->mAdapter:Lcom/faultexception/reader/util/adapters/MultiAdapter;
					iget-object v1, p0, Lcom/faultexception/reader/MainDrawerFragment;->mSeriesSection:Lcom/faultexception/reader/util/adapters/MultiAdapter;
					invoke-virtual {v0, v1, v2}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->setEnabled(Landroid/widget/BaseAdapter;Z)V
					if-eqz v2, :done

					iget-object v0, p0, Lcom/faultexception/reader/MainDrawerFragment;->mDatabase:Landroid/database/sqlite/SQLiteDatabase;
					const-string v1, "SELECT _id, series FROM books WHERE hidden=0 AND series IS NOT NULL GROUP BY series"
					const v2, 0x0
					invoke-virtual {v0, v1, v2}, Landroid/database/sqlite/SQLiteDatabase;->rawQuery(Ljava/lang/String;[Ljava/lang/String;)Landroid/database/Cursor;
					move-result-object v0
					iget-object v1, p0, Lcom/faultexception/reader/MainDrawerFragment;->mSeriesAdapter:Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;
					invoke-virtual {v1, v0}, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;->changeCursor(Landroid/database/Cursor;)V

					:done
					return-void
				.end method
				`),
			),

			InMethod("onItemClick(Landroid/widget/AdapterView;Landroid/view/View;IJ)V",
				ReplaceStringAppend(
					FixIndent("\n"+`
						iget-object p1, p0, Lcom/faultexception/reader/MainDrawerFragment;->mAdapter:Lcom/faultexception/reader/util/adapters/MultiAdapter;

						invoke-virtual {p1, p3}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->getProjection(I)Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;

						move-result-object p1
					`),
					FixIndent("\n"+`
						invoke-direct {p0, p1}, Lcom/faultexception/reader/MainDrawerFragment;->checkItemClickSeries(Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;)V
					`),
				),
			),
			// based on the part of onItemClick which handles folders
			ReplaceStringPrepend(
				FixIndent("\n"+`.method public onItemClick(Landroid/widget/AdapterView;Landroid/view/View;IJ)V`),
				FixIndent("\n"+`
				.method private checkItemClickSeries(Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;)V
					.locals 3

					iget-object v0, p0, Lcom/faultexception/reader/MainDrawerFragment;->mSeriesSection:Lcom/faultexception/reader/util/adapters/MultiAdapter;
					iget-object v1, p1, Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;->adapter:Landroid/widget/BaseAdapter;
					if-ne v0, v1, :not

					iget v2, p1, Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;->index:I
					invoke-virtual {v0, v2}, Lcom/faultexception/reader/util/adapters/MultiAdapter;->getProjection(I)Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;
					move-result-object p1

					iget-object v0, p0, Lcom/faultexception/reader/MainDrawerFragment;->mSeriesAdapter:Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;
					iget-object v1, p1, Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;->adapter:Landroid/widget/BaseAdapter;
					if-ne v0, v1, :not

					iget v2, p1, Lcom/faultexception/reader/util/adapters/MultiAdapter$Projection;->index:I
					invoke-virtual {v0, v2}, Lcom/faultexception/reader/MainDrawerFragment$SeriesAdapter;->getSeriesAtIndex(I)Ljava/lang/String;
					move-result-object p1

					invoke-static {p1}, Lcom/faultexception/reader/BooksFragment;->newSeriesInstance(Ljava/lang/String;)Landroidx/fragment/app/Fragment;
					move-result-object p1

					iget-object v2, p0, Lcom/faultexception/reader/MainDrawerFragment;->mListener:Lcom/faultexception/reader/MainDrawerFragment$Listener;
					invoke-interface {v2, p1}, Lcom/faultexception/reader/MainDrawerFragment$Listener;->switchToFragment(Landroidx/fragment/app/Fragment;)V

					:not
					return-void
				.end method
				`),
			),
		),

		PatchFile("smali/com/faultexception/reader/MainActivity.smali",
			InMethod("onCreate(Landroid/os/Bundle;)V",
				MustContain(FixIndent("\n"+`
					const-string v3, "filter"

					invoke-interface {v2, v3, v0}, Landroid/content/SharedPreferences;->getInt(Ljava/lang/String;I)I

					move-result v2

					if-eqz v2, :cond_4

					if-eq v2, v1, :cond_3

					const/4 v3, 0x2

					if-eq v2, v3, :cond_2

					goto :goto_1

					.line 182
					:cond_2
					iget-object v2, p0, Lcom/faultexception/reader/MainActivity;->mPrefs:Landroid/content/SharedPreferences;

					const/4 v3, 0x0

					const-string v4, "folder"

					.line 183
					invoke-interface {v2, v4, v3}, Landroid/content/SharedPreferences;->getString(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;

					move-result-object v2

					invoke-static {v2}, Lcom/faultexception/reader/BooksFragment;->newFolderInstance(Ljava/lang/String;)Landroidx/fragment/app/Fragment;

					move-result-object v2

					.line 182
					invoke-virtual {p0, v2, v0}, Lcom/faultexception/reader/MainActivity;->switchToFragment(Landroidx/fragment/app/Fragment;Z)V

					goto :goto_1
				`)),
				ReplaceStringAppend(
					FixIndent("\n"+`
						const/4 v3, 0x2

						if-eq v2, v3, :cond_2
					`),
					FixIndent("\n"+`
						const v3, `+filterId+`

						if-eq v2, v3, :cond_2a
					`),
				),
				ReplaceStringPrepend(
					FixIndent("\n"+`
						:cond_2
						iget-object v2, p0, Lcom/faultexception/reader/MainActivity;->mPrefs:Landroid/content/SharedPreferences;
					`),
					FixIndent("\n"+`
						:cond_2a
						iget-object v2, p0, Lcom/faultexception/reader/MainActivity;->mPrefs:Landroid/content/SharedPreferences;
						const v3, 0x0
						const-string v4, "series"
						invoke-interface {v2, v4, v3}, Landroid/content/SharedPreferences;->getString(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
						move-result-object v2
						invoke-static {v2}, Lcom/faultexception/reader/BooksFragment;->newSeriesInstance(Ljava/lang/String;)Landroidx/fragment/app/Fragment;
						move-result-object v2
						invoke-virtual {p0, v2, v0}, Lcom/faultexception/reader/MainActivity;->switchToFragment(Landroidx/fragment/app/Fragment;Z)V
						goto :goto_1
					`),
				),
			),
			InMethod("updateFragmentInfo()V",
				MustContain(FixIndent("\n"+`
					const-string v3, "category_id"

					invoke-interface {v2, v3, v0, v1}, Landroid/content/SharedPreferences$Editor;->putLong(Ljava/lang/String;J)Landroid/content/SharedPreferences$Editor;

					goto :goto_0
				`)),
				ReplaceStringPrepend(
					FixIndent("\n"+`
						const/4 v3, 0x2

						if-ne v0, v3, :cond_1

						const-string v0, "folder"

						.line 274
						invoke-virtual {v1, v0}, Landroid/os/Bundle;->getString(Ljava/lang/String;)Ljava/lang/String;

						move-result-object v1

						invoke-interface {v2, v0, v1}, Landroid/content/SharedPreferences$Editor;->putString(Ljava/lang/String;Ljava/lang/String;)Landroid/content/SharedPreferences$Editor;

						.line 277
						:cond_1
					`),
					FixIndent("\n"+`
						const v3, `+filterId+`
						if-ne v0, v3, :cond_1a

						sget-object v0, Lcom/faultexception/reader/BooksFragment;->EXTRA_SERIES:Ljava/lang/String;
						invoke-virtual {v1, v0}, Landroid/os/Bundle;->getString(Ljava/lang/String;)Ljava/lang/String;
						move-result-object v1
						invoke-interface {v2, v0, v1}, Landroid/content/SharedPreferences$Editor;->putString(Ljava/lang/String;Ljava/lang/String;)Landroid/content/SharedPreferences$Editor;

						goto :goto_0

						:cond_1a
					`),
				),
			),
		),
	)
}
