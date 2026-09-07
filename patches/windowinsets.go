// # Window insets
//
// Fix window inset issues on some Android versions.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("windowinsets",
		PatchFiles([]string{
			"res/layout/activity_settings.xml",
			"res/layout/activity_settings_licenses.xml",
			"res/layout/activity_backups.xml",
			"res/layout/activity_themes.xml",
			"res/layout-v17/activity_themes.xml",
		},
			ReplaceString(
				"\n"+`<LinearLayout android:orientation="vertical" android:layout_width="match_parent" android:layout_height="match_parent"`+"\n",
				"\n"+`<LinearLayout android:orientation="vertical" android:layout_width="match_parent" android:layout_height="match_parent" android:fitsSystemWindows="true"`+"\n",
			),
		),
	)
}
