// # Pseudo Material Design 3
//
// Make some components more M3-like, use Material You colors for library on
// Android 12+.
package patches

import (
	"strconv"

	. "github.com/pgaskin/lithiumpatch/patches/patchdef"
)

func init() {
	Register("pseudomat3",
		// material you colors in most places
		// https://github.com/material-components/material-components-android/blob/778a9f2a490394906e97881bb29ac491a64c23d7/docs/theming/Color.md?plain=1#L36
		// TODO: maybe add light colors too?
		WriteFileString("res/values-night-v31/colors.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<resources>
				<color name="app_primary">@android:color/system_accent1_200</color>
				<color name="app_primary_dark">@color/color_surface</color>
				<color name="app_secondary">@android:color/system_accent2_200</color>
				<color name="navigation_bar_color">@android:color/system_neutral1_900</color>
				<color name="color_surface">@android:color/system_neutral1_900</color>
				<color name="surface_status_bar_color">@color/color_surface</color>
				<color name="launch_toolbar_color">@color/color_surface</color>
				<color name="book_list_background_color">@android:color/system_neutral1_900</color>
				<color name="book_item_background_color">@android:color/system_neutral1_800</color>
				<color name="book_item_no_cover_tint">@android:color/system_neutral1_600</color>
				<color name="book_item_selected_check_color">@android:color/system_accent1_800</color>
				<color name="book_item_selected_scrim_color">#77ffffff</color>
				<color name="drawer_background_color">@android:color/system_neutral1_900</color>
				<color name="drawer_item_selected_bg_color">@android:color/system_neutral1_800</color>
				<color name="drawer_scrim_color">#55000000</color>
				<color name="search_status_bar_color">@color/color_surface</color>
				<color name="search_toolbar_color">@color/color_surface</color>
				<color name="reader_search_bg_color">@color/book_list_background_color</color>
				<color name="reader_drawer_tabs_background">@color/color_surface</color>
				<!--<color name="reader_dark_chrome_color">@android:color/system_neutral2_900</color>-->

				<!--<color name="color_circle_stroke_overlay">#32ffffff</color>-->
				<!--<color name="display_settings_disabled_icon_color">#42ffffff</color>-->
				<!--<color name="divider">#1fffffff</color>-->
				<!--<color name="ripple_fallback_color">#1fffffff</color>-->
				<!--<color name="search_recent_icon_tint">#40ffffff</color>-->
				<!--<color name="selection_notes_none_color">#ffb9b9b9</color>-->
				<!--<color name="text_hint">#61ffffff</color>-->
				<!--<color name="themes_new_circle_color">#ff525252</color>-->
			</resources>
			`),
		),
		// flat and more rounded book items, slightly bigger margins
		PatchFiles(
			[]string{
				"res/layout/books_list_item.xml",
				"res/layout/books_grid_item.xml",
			},
			ReplaceString(
				`app:cardCornerRadius="4.0dp"`,
				`app:cardCornerRadius="12.0dp" app:cardElevation="0dp"`,
			),
			ReplaceString(
				`android:layout_margin="5.0dp"`,
				`android:layout_margin="8.0dp"`,
			),
		),
		PatchFile("smali/com/faultexception/reader/BooksFragment.smali",
			MustContain(
				FixIndent("\n"+`
					iget-object p1, p0, Lcom/faultexception/reader/BooksFragment;->mContext:Landroid/content/Context;

					const/16 v0, 0xa

					.line 167
					invoke-static {p1, v0}, Lcom/faultexception/reader/util/Utils;->dpToPx(Landroid/content/Context;I)I

					move-result p1
				`),
			),
			MustContain(
				FixIndent("\n"+`
					invoke-virtual {v2, v3}, Landroid/content/res/Resources;->getDimension(I)F

					move-result v2

					float-to-int v2, v2

					add-int/2addr v2, p1

					.line 168
					invoke-virtual {v0, v2}, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->setSpanWidth(I)V
				`),
			),
			ReplaceString(
				FixIndent("\n"+`
					iget-object p1, p0, Lcom/faultexception/reader/BooksFragment;->mContext:Landroid/content/Context;

					const/16 v0, 0xa
				`),
				FixIndent("\n"+`
					iget-object p1, p0, Lcom/faultexception/reader/BooksFragment;->mContext:Landroid/content/Context;

					const/16 v0, `+strconv.Itoa(8*2)+`
				`),
			),
		),
		// don't darken status bar color in reader
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("applyChromeColor()V",
				ReplaceString(
					FixIndent("\n"+`
						invoke-static {v0}, Landroid/graphics/Color;->blue(I)I

						move-result v4

						int-to-float v4, v4

						mul-float v4, v4, v2

						float-to-int v2, v4

						.line 535
						invoke-static {v1, v3, v2}, Landroid/graphics/Color;->rgb(III)I

						move-result v1
					`),
					FixIndent("\n"+`
						move v1, v0
					`),
				),
			),
		),
		// non-full-width elastic tab indicator for the reader drawer
		WriteFileString("res/drawable/m3_tab_indicator.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<inset xmlns:android="http://schemas.android.com/apk/res/android" android:insetLeft="-10.0dp" android:insetRight="-10.0dp">
				<shape android:shape="rectangle">
					<corners android:topLeftRadius="3.0dp" android:topRightRadius="3.0dp" />
					<solid android:color="@android:color/white" />
					<size android:height="3.0dp" />
				</shape>
			</inset>
			`),
		),
		DefineR("smali/com/faultexception/reader", "drawable", "m3_tab_indicator"),
		PatchFile("res/layout/fragment_reader_drawer.xml",
			ReplaceString(
				`android:layout_height="?actionBarSize" app:tabGravity="fill" />`,
				`android:layout_height="?actionBarSize" app:tabGravity="fill" app:tabIndicator="@drawable/m3_tab_indicator" app:tabIndicatorHeight="3.0dp" app:tabIndicatorFullWidth="false" app:tabIndicatorAnimationMode="elastic" />`,
			),
		),
		// pill-style drawer items
		WriteFileString("res/drawable/drawer_item_background_selector.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<selector xmlns:android="http://schemas.android.com/apk/res/android">
				<item android:state_activated="true">
					<inset android:insetLeft="12.0dp" android:insetRight="12.0dp">
						<shape android:shape="rectangle">
							<corners android:radius="12.0dp" />
							<solid android:color="@color/drawer_item_selected_bg_color" />
						</shape>
					</inset>
				</item>
			</selector>
			`),
		),
		// preserve 16dp icon inset (since the pill is inset by 12dp)
		PatchFiles(
			[]string{
				"res/layout/drawer_item.xml",
				"res/layout-v17/drawer_item.xml",
			},
			ReplaceString(
				`android:paddingLeft="16.0dp" android:paddingRight="16.0dp"`,
				`android:paddingLeft="28.0dp" android:paddingRight="28.0dp"`,
			),
		),
		// remove drawer divider lines (it looks too busy with them)
		PatchFiles(
			[]string{
				"res/layout/drawer_divider.xml",
				"res/layout/categories_list_header.xml",
				"res/layout/folders_list_header.xml",
			},
			ReplaceString(
				`<View android:background="@color/divider"`,
				`<View android:background="@android:color/transparent"`,
			),
		),
		// m3 drawer items don't change color when selected
		WriteFileString("res/color/drawer_text_color_selector.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<selector xmlns:android="http://schemas.android.com/apk/res/android">
				<item android:color="?android:textColorPrimary" />
			</selector>
			`),
		),
		WriteFileString("res/color/drawer_icon_color_selector.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<selector xmlns:android="http://schemas.android.com/apk/res/android">
				<item android:color="?colorControlNormal" />
			</selector>
			`),
		),
	)
}
