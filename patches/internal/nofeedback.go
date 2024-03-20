// # Remove feedback
//
// Removes the feedback option (since this is a patched version).
package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("nofeedback",
		PatchFile("res/menu/reader.xml",
			ReplaceString(
				`<item android:id="@id/feedback" android:title="@string/action_feedback" app:showAsAction="never" />`,
				`<item android:id="@id/feedback" android:title="@string/action_feedback" app:showAsAction="never" android:visible="false" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/util/adapters/DrawerFooterAdapter.smali",
			// hack: we assume the last item is feedback
			InMethod("getCount()I",
				ReplaceString(
					`.locals 1`,
					`.locals 2`,
				),
				ReplaceStringAppend(
					FixIndent("\n"+`
						array-length v0, v0
					`),
					FixIndent("\n"+`
						const/4 v1, 0x1
						sub-int v0, v0, v1
					`),
				),
			),
		),
	)
}
