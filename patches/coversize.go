// # Grid view cover size
//
// Change the grid view cover aspect ratio to 1.5.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("coversize",
		PatchFile("res/values-sw364dp/dimens.xml",
			MustContain(`<dimen name="bookshelf_cover_width">160.0dip</dimen>`),
		),
		PatchFile("res/values-sw480dp/dimens.xml",
			MustContain(`<dimen name="bookshelf_cover_width">180.0dip</dimen>`),
		),
		PatchFile("res/layout/books_grid_item.xml",
			ReplaceString(
				"\n"+`        <FrameLayout android:id="@id/cover_container" android:layout_width="@dimen/bookshelf_cover_width" android:layout_height="@dimen/bookshelf_cover_height">`,
				"\n"+`        <com.faultexception.reader.widget.CoverFrameLayout android:id="@id/cover_container" android:layout_width="@dimen/bookshelf_cover_width" android:layout_height="wrap_content">`,
			),
			ReplaceString(
				"\n"+`        </FrameLayout>`,
				"\n"+`        </com.faultexception.reader.widget.CoverFrameLayout>`,
			),
		),
		WriteFileString("smali/com/faultexception/reader/widget/CoverFrameLayout.smali",
			FixIndent(`
			.class public Lcom/faultexception/reader/widget/CoverFrameLayout;
			.super Landroid/widget/FrameLayout;

			.method public constructor <init>(Landroid/content/Context;)V
				.locals 0
				invoke-direct {p0, p1}, Landroid/widget/FrameLayout;-><init>(Landroid/content/Context;)V
				return-void
			.end method

			.method public constructor <init>(Landroid/content/Context;Landroid/util/AttributeSet;)V
				.locals 0
				invoke-direct {p0, p1, p2}, Landroid/widget/FrameLayout;-><init>(Landroid/content/Context;Landroid/util/AttributeSet;)V
				return-void
			.end method

			.method public constructor <init>(Landroid/content/Context;Landroid/util/AttributeSet;I)V
				.locals 0
				invoke-direct {p0, p1, p2, p3}, Landroid/widget/FrameLayout;-><init>(Landroid/content/Context;Landroid/util/AttributeSet;I)V
				return-void
			.end method

			.method protected onMeasure(II)V
				.locals 3

				# get width
				invoke-static {p1}, Landroid/view/View$MeasureSpec;->getSize(I)I
				move-result v0

				# 1.5 aspect ratio
				div-int/lit8 v1, v0, 0x2
				mul-int/lit8 v1, v1, 0x3

				# set height
				const/high16 v2, 0x40000000 # MeasureSpec.EXACTLY
				invoke-static {v1, v2}, Landroid/view/View$MeasureSpec;->makeMeasureSpec(II)I
				move-result p2

				invoke-super {p0, p1, p2}, Landroid/widget/FrameLayout;->onMeasure(II)V
				return-void
			.end method
			`),
		),
	)
}
