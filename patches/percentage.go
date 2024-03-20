// # Reader footer
//
// Add chapter progress and percentage to the reader footer.
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("percentage",
		PatchFile("res/values/strings.xml",
			ReplaceString(
				`<string name="page_number_label">%1$d/%2$d</string>`,
				`<string name="page_number_label">%3$d%% chapter&#160;&#160;|&#160;&#160;%1$d/%2$d %4$d%%</string>`,
			),
		),
		PatchFile("smali/com/faultexception/reader/ReaderActivity.smali",
			InMethod("updateReadingProgress()V",
				// make new v9, v10 vars to avoid conflicts and ease updating
				ReplaceString(
					".locals 9",
					".locals 11",
				),
				ReplaceString(
					FixIndent("\n"+`
						new-array v2, v2, [Ljava/lang/Object;
					`),
					FixIndent("\n"+`
						const/4 v9, 0x4
						new-array v2, v9, [Ljava/lang/Object;
					`),
				),
				ReplaceStringAppend(
					FixIndent("\n"+`
						aput-object v8, v2, v4
					`),
					FixIndent("\n"+`
						iget-object v9, p0, Lcom/faultexception/reader/ReaderActivity;->mBookView:Lcom/faultexception/reader/content/BookView;
						invoke-virtual {v9}, Lcom/faultexception/reader/content/BookView;->getScrollPosition()F
						move-result v9
						const/high16 v10, 0x42c80000     # 100.0f
						mul-float/2addr v9, v10
						invoke-static {v9}, Ljava/lang/Math;->round(F)I
						move-result v9
						invoke-static {v9}, Ljava/lang/Integer;->valueOf(I)Ljava/lang/Integer;
						move-result-object v9
						const/4 v10, 0x2
						aput-object v9, v2, v10

						const/4 v9, 0x0
						const/4 v10, 0x1
						aget-object v9, v2, v9
						aget-object v10, v2, v10
						check-cast v9, Ljava/lang/Integer;
						check-cast v10, Ljava/lang/Integer;
						invoke-virtual {v9}, Ljava/lang/Integer;->intValue()I
						move-result v9
						invoke-virtual {v10}, Ljava/lang/Integer;->intValue()I
						move-result v10
						int-to-float v9, v9 # current page (from earlier, arr[0])
						int-to-float v10, v10 # total pages (from earlier, arr[1])

						div-float/2addr v9, v10
						const/high16 v10, 0x42c80000 # 100.0f
						mul-float/2addr v9, v10
						invoke-static {v9}, Ljava/lang/Math;->round(F)I
						move-result v9
						invoke-static {v9}, Ljava/lang/Integer;->valueOf(I)Ljava/lang/Integer;
						move-result-object v9
						const/4 v10, 0x3
						aput-object v9, v2, v10
					`),
				),
			),
		),
	)
}
