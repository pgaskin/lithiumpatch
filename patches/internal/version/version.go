// Package version patches the version string in the UI.
package version

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

func init() {
	Register("version",
		PatchFile("smali/com/faultexception/reader/SettingsActivity$PreferencesFragment.smali",
			InMethod("onCreatePreferences(Landroid/os/Bundle;Ljava/lang/String;)V",
				ReplaceStringPrepend(
					FixIndent("\n"+`
						invoke-virtual {v0, p2}, Landroidx/preference/Preference;->setSummary(Ljava/lang/CharSequence;)V
					`),
					FixIndent("\n"+`
						invoke-static {p2}, Lcom/faultexception/reader/SettingsActivity$PreferencesFragment;->addPatched(Ljava/lang/String;)Ljava/lang/String;
						move-result-object p2
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onCreatePreferences(Landroid/os/Bundle;Ljava/lang/String;)V
				`),
				FixIndent("\n"+`
				.method private static addPatched(Ljava/lang/String;)Ljava/lang/String;
					.locals 1

					const-string v0, " (Patched)"
					invoke-virtual {p0, v0}, Ljava/lang/String;->concat(Ljava/lang/String;)Ljava/lang/String;
					move-result-object p0

					return-object p0
				.end method
				`),
			),
		),
	)
}
