// # Smaller font sizes
//
// Allow smaller font sizes to be selected.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("minsize",
		PatchFile("smali/com/faultexception/reader/DisplaySettingsFragment.smali",
			InConstant("TEXT_SIZE_MIN:I",
				ReplaceString("0x50", "0x3c"),
			),
			InMethod("update()V",
				ReplaceString("0x50", "0x3c"),
			),
			InMethod("onClick(Landroid/view/View;)V",
				ReplaceString("0x50", "0x3c"),
			),
		),
	)
}
