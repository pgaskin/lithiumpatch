// # Disable sync
//
// If built with the default keystore, disables the sync functionality since it
// isn't available unless a signing key is registered with Google APIs.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func NoSync() {
	Register("nosync",
		PatchFile("res/xml/sync.xml",
			ReplaceString(
				"@string/pref_sync_summary",
				"Sync requires the patched Lithium APK to be signed with a custom key, which must be registered with the Google Drive API. See https://github.com/pgaskin/lithiumpatch for more information.",
			),
		),
		PatchFile("smali/com/faultexception/reader/sync/SyncSettingsFragment.smali",
			InMethod("onCreatePreferences(Landroid/os/Bundle;Ljava/lang/String;)V",
				ReplaceStringAppend(
					FixIndent(`
						:cond_0
						iget-object p1, p0, Lcom/faultexception/reader/sync/SyncSettingsFragment;->mProRequiredPref:Landroidx/preference/Preference;

						invoke-virtual {p1, v2}, Landroidx/preference/Preference;->setVisible(Z)V
					`),
					FixIndent(`

						iget-object p1, p0, Lcom/faultexception/reader/sync/SyncSettingsFragment;->mEnabledPref:Landroidx/preference/SwitchPreferenceCompat;

						invoke-virtual {p1, v2}, Landroidx/preference/SwitchPreferenceCompat;->setEnabled(Z)V
					`),
				),
			),
		),
	)
}
