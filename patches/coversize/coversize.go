// Package coversize changes the grid view cover aspect ration to 1.5 and adds
// values for sw.
package coversize

import . "github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

func init() {
	Register("coversize",
		PatchFile("res/values-sw364dp/dimens.xml",
			MustContain(`<dimen name="bookshelf_cover_width">160.0dip</dimen>`),
			ReplaceString(
				`<dimen name="bookshelf_cover_height">212.0dip</dimen>`,
				`<dimen name="bookshelf_cover_height">240.0dip</dimen>`,
			),
		),
		PatchFile("res/values-sw480dp/dimens.xml",
			MustContain(`<dimen name="bookshelf_cover_width">180.0dip</dimen>`),
			ReplaceString(
				`<dimen name="bookshelf_cover_height">239.0dip</dimen>`,
				`<dimen name="bookshelf_cover_height">270.0dip</dimen>`,
			),
		),
	)
}
