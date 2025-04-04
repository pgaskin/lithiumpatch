package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

// Lithium isn't available on the Play Store anymore as of April 2025, so always
// unlock pro features.

func init() {
	Register("prv",
		PatchFile("smali/com/faultexception/reader/model/ProManager.smali",
			InMethod("setUnlockedState(Landroid/app/Activity;Z)V",
				ReplaceWith(FixIndent(`
							.locals 3
							invoke-static {p0}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
							move-result-object v0

							const-string v1, "pro_unlocked"
							invoke-interface {v0}, Landroid/content/SharedPreferences;->edit()Landroid/content/SharedPreferences$Editor;
							move-result-object v0

							const/4 v2, 0x1
							invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences$Editor;->putBoolean(Ljava/lang/String;Z)Landroid/content/SharedPreferences$Editor;
							move-result-object v0

							invoke-interface {v0}, Landroid/content/SharedPreferences$Editor;->apply()V
							return-void
					`)),
			),
		),
	)
}
