// # Invert rotation
//
// Add a toolbar icon to invert the screen rotation.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("invertrotation",
		WriteFileString("res/drawable-anydpi-v21/ic_rotate_alt.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<vector xmlns:android="http://schemas.android.com/apk/res/android" android:height="24dp" android:width="24dp" android:viewportWidth="24" android:viewportHeight="24">
				<path android:fillColor="#ffffffff" android:pathData="M4,7.59l5-5c0.78-0.78,2.05-0.78,2.83,0L20.24,11h-2.83L10.4,4L5.41,9H8v2H2V5h2V7.59z M20,19h2v-6h-6v2h2.59l-4.99,5 l-7.01-7H3.76l8.41,8.41c0.78,0.78,2.05,0.78,2.83,0l5-5V19z" />
			</vector>
			`),
		),
		DefineR("smali/com/faultexception/reader", "drawable", "ic_rotate_alt"),
		PatchFile("res/values/ids.xml",
			ReplaceStringPrepend(
				"\n"+`</resources>`,
				"\n"+`    <item type="id" name="invert_rotation" />`,
			),
		),
		DefineR("smali/com/faultexception/reader", "id", "invert_rotation"),
		PatchFile("res/menu/reader.xml",
			ReplaceStringPrepend(
				"\n"+`    <item android:id="@id/settings"`,
				"\n"+`    <item android:icon="@drawable/ic_rotate_alt" android:id="@id/invert_rotation" android:title="Invert Rotation" app:showAsAction="ifRoom" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("onOptionsItemSelected(Landroid/view/MenuItem;)Z",
				ReplaceStringPrepend(
					FixIndent("\n"+`
						const/4 v2, 0x1

						sparse-switch v0, :sswitch_data_0
					`),
					FixIndent("\n"+`
						sget v2, Lcom/faultexception/reader/R$id;->invert_rotation:I
						if-ne v0, v2, :not_invert_rotation
						invoke-direct {p0}, Lcom/faultexception/reader/ReaderActivity;->invertRotation()V
						:not_invert_rotation
					`),
				),
			),
			ReplaceStringPrepend(
				FixIndent("\n"+`
				.method public onOptionsItemSelected(Landroid/view/MenuItem;)Z
				`),
				FixIndent("\n"+`
				.method private invertRotation()V
					.locals 1

					# attempt to get the last requested orientation
					invoke-virtual {p0}, Lcom/faultexception/reader/ReaderActivity;->getRequestedOrientation()I
					move-result v0
					sparse-switch v0, :screen_orientation

					# unknown or none, so attempt to do it based on if we're currently using portrait or landscape resources
					invoke-virtual {p0}, Lcom/faultexception/reader/ReaderActivity;->getResources()Landroid/content/res/Resources;
					move-result-object v0
					invoke-virtual {v0}, Landroid/content/res/Resources;->getConfiguration()Landroid/content/res/Configuration;
					move-result-object v0
					iget v0, v0, Landroid/content/res/Configuration;->orientation:I
					sparse-switch v0, :resources_orientation

					:orientation_rp
					const v0, 9 # SCREEN_ORIENTATION_REVERSE_PORTRAIT
					goto :try0s

					:orientation_rl
					const v0, 8 # SCREEN_ORIENTATION_REVERSE_LANDSCAPE
					goto :try0s

					:orientation_np
					const v0, 1 # SCREEN_ORIENTATION_PORTRAIT
					goto :try0s

					:orientation_nl
					const v0, 0 # SCREEN_ORIENTATION_LANDSCAPE
					goto :try0s

					:try0s
					invoke-virtual {p0, v0}, Lcom/faultexception/reader/ReaderActivity;->setRequestedOrientation(I)V
					:try0e
					.catch Ljava/lang/IllegalStateException; {:try0s .. :try0e} :try0c
					:try0c

					return-void

					:screen_orientation
					.sparse-switch
						0  -> :orientation_rl # SCREEN_ORIENTATION_LANDSCAPE
						1  -> :orientation_rp # SCREEN_ORIENTATION_PORTRAIT
						6  -> :orientation_rl # SCREEN_ORIENTATION_SENSOR_LANDSCAPE
						7  -> :orientation_rp # SCREEN_ORIENTATION_SENSOR_PORTRAIT
						8  -> :orientation_nl # SCREEN_ORIENTATION_REVERSE_LANDSCAPE
						9  -> :orientation_np # SCREEN_ORIENTATION_REVERSE_PORTRAIT
						11 -> :orientation_rl # SCREEN_ORIENTATION_USER_LANDSCAPE
						12 -> :orientation_rp # SCREEN_ORIENTATION_USER_PORTRAIT
					.end sparse-switch

					:resources_orientation
					.sparse-switch
						1 -> :orientation_rp # ORIENTATION_PORTRAIT
						2 -> :orientation_rl # ORIENTATION_LANDSCAPE
					.end sparse-switch
				.end method
				`),
			),
		),
	)
}
