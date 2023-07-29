// Package moretoolbaractions makes more of the reader toolbar actions always visible.
package moretoolbaractions

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

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
