// # HtmlContentWebView visibility
//
// Make the HtmlContentWebView Context public.
package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("public_hcwv_ctx",
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceString(
				`.field private mContext:Landroid/content/Context;`,
				`.field public mContext:Landroid/content/Context;`,
			),
		),
	)
}
