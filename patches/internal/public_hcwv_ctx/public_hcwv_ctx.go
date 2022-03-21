// Package public_hcwv_ctx makes the HtmlContentWebView Context public.
package public_hcwv_ctx

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

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
