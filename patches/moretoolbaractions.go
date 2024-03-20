// # More toolbar actions
//
// Always show the search and bookmark toolbar actions regardless of orientation
// or screen size.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("moretoolbaractions",
		PatchFile("res/menu/reader.xml",
			ReplaceString(
				`android:title="@string/action_search" app:showAsAction="ifRoom"`,
				`android:title="@string/action_search" app:showAsAction="always"`,
			),
			ReplaceString(
				`android:title="@string/action_add_bookmark" app:showAsAction="ifRoom"`,
				`android:title="@string/action_add_bookmark" app:showAsAction="always"`,
			),
		),
	)
}
