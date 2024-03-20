// # Signature check
//
// Fix the pro signature check. You still need to buy the app to be able to use
// pro features.
package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("signatures",
		PatchFile("smali/com/faultexception/reader/model/ProManager.smali",
			InMethod("checkIfNecessary(Landroid/app/Activity;)V",
				ReplaceString(
					FixIndent("\n"+`
						invoke-virtual {v1, v2, v3}, Landroid/content/pm/PackageManager;->checkSignatures(Ljava/lang/String;Ljava/lang/String;)I

						move-result v2
					`),
					FixIndent("\n"+`
						const/4 v2, 0x0
					`),
				),
			),
		),
	)
}
