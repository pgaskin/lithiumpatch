// # App icon color
//
// Change the app icon color from purple to pale dark blue.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("color",
		PatchFile("res/values/colors.xml",
			ReplaceString(
				`<color name="app_primary">#5f2deb</color>`,
				`<color name="app_primary">#104068</color>`,
			),
			ReplaceString(
				`<color name="app_primary_dark">#4a1cc9</color>`,
				`<color name="app_primary_dark">#002b5a</color>`,
			),
			ReplaceString(
				`<color name="ic_launcher_background">#784ef1</color>`,
				`<color name="ic_launcher_background">#466a96</color>`,
			),
		),
	)
}
