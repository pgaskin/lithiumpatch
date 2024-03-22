//go:build renamepkg

// # Rename package
//
// Optionally rename the package based on the renamepkg build tag.
package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	const (
		newpkg = "com.faultexception.reader.patched"
		bkpdir = "LithiumPatchedBackups"
	)
	Register("renamepkg",
		PatchFile("apktool.yml",
			ReplaceString(
				`renameManifestPackage: null`,
				`renameManifestPackage: "`+newpkg+`"`,
			),
		),
		PatchFile("AndroidManifest.xml",
			ReplaceString(
				`package="com.faultexception.reader"`,
				`package="`+newpkg+`"`,
			),
			ReplaceString(
				`android:authorities="com.faultexception.reader.fileprovider"`,
				`android:authorities="`+newpkg+`.fileprovider"`,
			),
			ReplaceString(
				`android:authorities="com.faultexception.reader.androidx-startup"`,
				`android:authorities="`+newpkg+`.androidx-startup"`,
			),
		),
		PatchFile("res/xml/fileprovider_paths.xml",
			ReplaceString(
				`path="LithiumBackups/"`,
				`path="`+bkpdir+`/"`,
			),
		),
		PatchFile("smali/com/faultexception/reader/backup/BackupsActivity.smali",
			ReplaceString(
				`"LithiumBackups"`,
				`"`+bkpdir+`"`,
			),
		),
		PatchFile("smali/com/faultexception/reader/BuildConfig.smali",
			ReplaceString(
				`.field public static final APPLICATION_ID:Ljava/lang/String; = "com.faultexception.reader"`,
				`.field public static final APPLICATION_ID:Ljava/lang/String; = "`+newpkg+`"`,
			),
		),
		PatchFile("smali/com/faultexception/reader/model/ProManager.smali",
			ReplaceString(
				`"com.faultexception.reader"`,
				`"`+newpkg+`"`,
			),
		),
	)
}
