// # Grid view cover size
//
// Change the grid view cover aspect ratio from 1.33 to 1.5. Expand it to fill
// the width of a column rather than adding margins to the sides of the grid
// view (after splitting by the default width).
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("coversize",
		// size breakpoints for cover grid
		// note: padding/gap is 5
		PatchFile("res/values-sw364dp/dimens.xml",
			ReplaceString(
				`<dimen name="bookshelf_cover_width">160.0dip</dimen>`,
				`<dimen name="bookshelf_cover_width">115.0dip</dimen>`,
			),
		),
		PatchFile("res/values-sw480dp/dimens.xml",
			ReplaceString(
				`<dimen name="bookshelf_cover_width">180.0dip</dimen>`,
				`<dimen name="bookshelf_cover_width">115.0dip</dimen>`,
			),
		),
		// the base width to determine the number of columns is set in code like AutoFitRecyclerView.setSpanWidth(bookshelf_cover_width)
		PatchFile("smali/com/faultexception/reader/BooksFragment.smali",
			MustContain(
				FixIndent("\n"+`
					.line 167
					invoke-virtual {v2, v3}, Landroid/content/res/Resources;->getDimension(I)F
				
					move-result v2
				
					float-to-int v2, v2
				
					add-int/2addr v2, p1
				
					.line 166
					invoke-virtual {v0, v2}, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->setSpanWidth(I)V
				`),
			),
		),
		// set the cover width on the book card itself, then let the cover drive the height
		PatchFile("res/layout/books_grid_item.xml",
			ReplaceString(
				`<androidx.cardview.widget.CardView android:layout_width="wrap_content" android:layout_height="wrap_content"`,
				`<androidx.cardview.widget.CardView android:layout_width="fill_parent" android:layout_height="wrap_content"`,
			),
			ReplaceString(
				"\n"+`        <FrameLayout android:id="@id/cover_container" android:layout_width="@dimen/bookshelf_cover_width" android:layout_height="@dimen/bookshelf_cover_height">`,
				"\n"+`        <com.faultexception.reader.widget.CoverFrameLayout android:id="@id/cover_container" android:layout_width="fill_parent" android:layout_height="wrap_content">`,
			),
			ReplaceString(
				"\n"+`        </FrameLayout>`,
				"\n"+`        </com.faultexception.reader.widget.CoverFrameLayout>`,
			),
		),
		// frame layout, but setting the height from the width and an aspect ratio
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
		// AutoFitRecyclerView is only used by fragment_books, so it's much easier to change it there than to change all references
		PatchFile("smali/com/faultexception/reader/widget/AutoFitRecyclerView.smali",
			InMethod("onMeasure(II)V",
				// never add padding to the sides of the grid layout (rather than padding it based on the leftover after fitting in as many as possible)
				MustContain(
					FixIndent("\n"+`
						:cond_2
						iget p1, p0, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->mPaddingLeft:I

						invoke-virtual {p0}, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->getPaddingTop()I

						move-result p2

						iget v0, p0, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->mPaddingRight:I

						invoke-virtual {p0}, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->getPaddingRight()I

						move-result v1

						invoke-virtual {p0, p1, p2, v0, v1}, Lcom/faultexception/reader/widget/AutoFitRecyclerView;->setPadding(IIII)V

						:goto_0
						return-void
					`),
				),
				ReplaceString(
					"\n"+`    if-lez p1, :cond_2`,
					"\n"+`    goto :cond_2`,
				),
			),
		),
	)
}
