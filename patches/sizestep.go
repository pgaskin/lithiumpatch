// # Font size step
//
// Change the font size increment from 10% to 2%.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("sizestep",
		PatchFile("smali/com/faultexception/reader/DisplaySettingsFragment.smali",
			InMethod("onClick(Landroid/view/View;)V",
				// note: need to change v2 on both branches since v2 is reused
				// for the maximum margin
				MustContain(FixIndent("\n"+`
					const/16 v2, 0xa
				`)),
				MustContain(FixIndent("\n"+`
					if-ge p1, v2, :cond_1a
				`)),
				ReplaceString(
					FixIndent("\n"+`
						if-ne p1, v5, :cond_1c

						goto :goto_d

						:cond_1c
						const/16 v2, -0xa

						:goto_d
						add-int/2addr v0, v2
					`),
					FixIndent("\n"+`
						if-ne p1, v5, :cond_1c

						const/16 v2, 0x2

						goto :goto_d

						:cond_1c
						const/16 v2, -0x2

						:goto_d
						add-int/2addr v0, v2
					`),
				),
			),
		),
	)
}
