// # Increase minimum API level
//
// Set the minimum API level to 26 (8: Oreo). Ships with webview 58 (but
// upgradeable -- note that we want ~73+ for some of our patches). Ships with
// SQLite 3.18. Supports APK signatures v2.
package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("minsdk",
		PatchFile("apktool.yml",
			ReplaceString(
				`minSdkVersion: '16'`,
				`minSdkVersion: '26'`,
			),
		),
	)
}
