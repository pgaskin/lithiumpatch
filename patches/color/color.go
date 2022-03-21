// Package color changes the app color from purple to pale dark blue.
package color

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

func init() {
	Register("color",
		PatchFile("res/values/colors.xml",
			ReplaceString(
				`<color name="app_primary">#ff5f2deb</color>`,
				`<color name="app_primary">#ff104068</color>`,
			),
			ReplaceString(
				`<color name="app_primary_dark">#ff4a1cc9</color>`,
				`<color name="app_primary_dark">#ff002b5a</color>`,
			),
			ReplaceString(
				`<color name="ic_launcher_background">#ff784ef1</color>`,
				`<color name="ic_launcher_background">#ff466a96</color>`,
			),
		),
	)
}
