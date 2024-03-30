// # Highlight fix
//
// Work around bug when highlighting over a gap in the columns (i.e.,
// top/bottom, but also left/right if modifying body to have zoom:0.5 for
// testing) of page causing the view to scroll unpredictably. This issue seems
// to be even more common for some people based on the recent Google Play
// reviews.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("highlightfix",
		PatchFile("assets/js/epub.js"), // TODO: I might need to entirely rework the pagination to use css transforms instead of scrolling to fix this
	)
}
