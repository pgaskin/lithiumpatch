// # Debuggable webview
//
// Make the reader webview debuggable with chrome://inspect.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("debuggable",
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			InMethod("<clinit>()V",
				ReplaceStringPrepend(
					FixIndent("\n"+`
						return-void
					`),
					FixIndent("\n"+`
						const/4 v0, 0x1
						invoke-static {v0}, Landroid/webkit/WebView;->setWebContentsDebuggingEnabled(Z)V
					`),
				),
			),
		),
	)
}
