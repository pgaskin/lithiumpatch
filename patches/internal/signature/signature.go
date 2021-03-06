// Package signature fixes the signature check. You still need to buy the app to be able to
// use pro features.
package signature

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

func init() {
	Register("signatures",
		PatchFile("smali/com/faultexception/reader/model/ProManager.smali",
			InMethod("checkIfNecessary(Landroid/app/Activity;)V",
				ReplaceString(
					FixIndent("\n"+`
						invoke-virtual {v1, v3, v2}, Landroid/content/pm/PackageManager;->checkSignatures(Ljava/lang/String;Ljava/lang/String;)I

						move-result v3
					`),
					FixIndent("\n"+`
						const/4 v3, 0x0
					`),
				),
			),
		),
	)
}
