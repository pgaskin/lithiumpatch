// # Adaptive app icon
//
// Add a monochrome adaptive app icon for Android 13+.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("adaptiveicon",
		PatchFile("res/mipmap-anydpi-v26/ic_launcher.xml",
			ReplaceStringAppend(
				"\n"+`    <foreground android:drawable="@mipmap/ic_launcher_foreground" />`,
				"\n"+`    <monochrome android:drawable="@drawable/ic_launcher_monochrome" />`,
			),
		),
		WriteFileString("res/drawable/ic_launcher_monochrome.xml",
			FixIndent(`
			<?xml version="1.0" encoding="utf-8"?>
			<vector xmlns:android="http://schemas.android.com/apk/res/android"
				android:width="90dp"
				android:height="90dp"
				android:viewportWidth="90"
				android:viewportHeight="90">
				<path
					android:pathData="M 58.621 57.5 L 31.4 57.5 C 30.06 57.5 29 56.422 29 55.108 L 29 34.892 C 29 33.556 30.081 32.5 31.4 32.5 L 58.6 32.5 C 59.94 32.5 61 33.578 61 34.892 L 61 55.086 C 61.021 56.422 59.94 57.5 58.621 57.5 Z M 45.021 32.5 L 45.021 57.5"
					android:strokeWidth="3"
					android:strokeColor="#FFFFFFFF"
					android:strokeLineJoin="round"
					android:fillColor="#00000000"
					android:fillAlpha="0"/>
				<path
					android:pathData="M 52.3 32.5 L 52.3 43.362 L 54.45 39.978 L 56.6 43.362 L 56.6 32.5 L 52.3 32.5 Z"
					android:strokeWidth="1.5"
					android:strokeLineJoin="round"
					android:strokeColor="#FFFFFFFF"
					android:fillColor="#FFFFFFFF"/>
			</vector>
			`),
		),
		DefineR("smali/com/faultexception/reader", "drawable", "ic_launcher_monochrome"),
	)
}

/* https://boxy-svg.com/app
<?xml version="1.0" encoding="utf-8"?>
	<svg viewBox="0 0 90 90" stroke="#000" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg">
	<path fill="none" d="M 58.621 57.5 L 31.4 57.5 C 30.06 57.5 29 56.422 29 55.108 L 29 34.892 C 29 33.556 30.081 32.5 31.4 32.5 L 58.6 32.5 C 59.94 32.5 61 33.578 61 34.892 L 61 55.086 C 61.021 56.422 59.94 57.5 58.621 57.5 Z M 45.021 32.5 L 45.021 57.5" style="stroke-width: 3px;" />
	<path d="M 52.3 32.5 L 52.3 43.362 L 54.45 39.978 L 56.6 43.362 L 56.6 32.5 L 52.3 32.5 Z" style="stroke-width: 1.5px;"/>
</svg>
*/
