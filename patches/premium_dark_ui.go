// # Premium Dark UI
//
// Implements a darker, high-contrast, beautiful dark theme.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("premium_dark_ui",
		// Overwrite the night colors with our premium palette
		WriteFileString("res/values-night/colors.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<resources>
				<!-- Premium Dark Palette -->
				
				<!-- Main Brand Colors: Electric Indigo & Teal -->
				<color name="app_primary">#ff5c6bc0</color>
				<color name="app_primary_dark">#ff050505</color> <!-- Blends with window background -->
				<color name="app_secondary">#ff26a69a</color>

				<!-- Backgrounds: Deep Black / Dark Grey -->
				<color name="window_background">#ff050505</color>
				<color name="color_surface">#ff121212</color>
				
				<!-- System Bars -->
				<color name="navigation_bar_color">#ff050505</color>
				<color name="surface_status_bar_color">#ff050505</color>
				<color name="launch_toolbar_color">#ff050505</color>

				<!-- Books List -->
				<color name="book_list_background_color">#ff050505</color>
				<color name="book_item_background_color">#ff1e1e1e</color>
				<color name="book_item_no_cover_tint">#ff424242</color>
				<color name="book_item_selected_check_color">#ff5c6bc0</color>
				<color name="book_item_selected_scrim_color">#99000000</color>

				<!-- Drawer -->
				<color name="drawer_background_color">#ff121212</color>
				<color name="drawer_item_selected_bg_color">#ff2c2c2c</color>
				<color name="drawer_scrim_color">#cc000000</color>

				<!-- Search & Reader -->
				<color name="search_status_bar_color">#ff050505</color>
				<color name="search_toolbar_color">#ff121212</color>
				<color name="reader_search_bg_color">#ff050505</color>
				<color name="reader_drawer_tabs_background">#ff121212</color>

				<!-- Override standard colors often used -->
				<color name="ic_launcher_background">#ff121212</color>
				
				<!-- Reader Chrome & Popups -->
				<color name="reader_dark_chrome_color">#ff050505</color>
				<color name="background_floating_material_dark">#ff121212</color> <!-- Popup menus -->
				<color name="tooltip_background_dark">#ff121212</color>
			</resources>
			`),
		),
		// Consistent branding for Light Mode (modifies res/values/colors.xml)
		// Replaces defaults with our Indigo theme, but keeps them appropriate for Light Mode if needed,
		// or just forces the brand color.
		PatchFile("res/values/colors.xml",
			ReplaceString(
				`<color name="app_primary">#5f2deb</color>`,
				`<color name="app_primary">#ff5c6bc0</color>`, // Indigo 400
			),
			ReplaceString(
				`<color name="app_primary_dark">#4a1cc9</color>`,
				`<color name="app_primary_dark">#ff3949ab</color>`, // Indigo 600
			),
			ReplaceString(
				`<color name="ic_launcher_background">#784ef1</color>`,
				`<color name="ic_launcher_background">#ff121212</color>`, // Keep icon background dark/neutral
			),
		),
		WriteFileString("res/drawable/bg_premium_popup.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<shape xmlns:android="http://schemas.android.com/apk/res/android">
				<solid android:color="#ff1a1a1a" />
				<corners android:radius="16dp" />
				<stroke android:width="1dp" android:color="#ff333333" />
			</shape>
			`),
		),
		// Force overwrite standard AppCompat popup background to ensure style is applied
		WriteFileString("res/drawable/abc_popup_background_mtrl_mult.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<shape xmlns:android="http://schemas.android.com/apk/res/android">
				<solid android:color="#ff1a1a1a" />
				<corners android:radius="16dp" />
				<stroke android:width="1dp" android:color="#ff333333" />
			</shape>
			`),
		),
		WriteFileString("res/drawable-v21/abc_popup_background_mtrl_mult.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<shape xmlns:android="http://schemas.android.com/apk/res/android">
				<solid android:color="#ff1a1a1a" />
				<corners android:radius="16dp" />
				<stroke android:width="1dp" android:color="#ff333333" />
			</shape>
			`),
		),
		WriteFileString("res/drawable-night/abc_popup_background_mtrl_mult.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<shape xmlns:android="http://schemas.android.com/apk/res/android">
				<solid android:color="#ff1a1a1a" />
				<corners android:radius="16dp" />
				<stroke android:width="1dp" android:color="#ff333333" />
			</shape>
			`),
		),
		WriteFileString("res/drawable/popup_background_material.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<shape xmlns:android="http://schemas.android.com/apk/res/android">
				<solid android:color="#ff1a1a1a" />
				<corners android:radius="16dp" />
				<stroke android:width="1dp" android:color="#ff333333" />
			</shape>
			`),
		),
		WriteFileString("res/drawable-night/popup_background_material.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<shape xmlns:android="http://schemas.android.com/apk/res/android">
				<solid android:color="#ff1a1a1a" />
				<corners android:radius="16dp" />
				<stroke android:width="1dp" android:color="#ff333333" />
			</shape>
			`),
		),
		DefineR("smali/com/faultexception/reader", "drawable", "bg_premium_popup"),
	)
}
