// # Disable animation
//
// Adds an option to disable page turn animation (useful for e-ink devices).
package patches

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("disableanimation",
		PatchFile("res/xml/preferences.xml",
			ReplaceStringAppend(
				"\n"+`    <PreferenceCategory android:title="@string/pref_category_navigation">`,
				"\n"+`        <SwitchPreferenceCompat android:title="Disable page turn animation" android:key="no_page_turn_animation" android:defaultValue="false" />`,
			),
		),
		PatchFile("smali/com/faultexception/reader/content/HtmlContentWebView.smali",
			ReplaceStringAppend(
				"\n"+`.field private mUrl:Ljava/lang/String;`,
				FixIndent("\n"+`
				.field private mNoPageTurnAnimation:Z

				.method private initNoPageTurnAnimation(Landroid/content/Context;)V
					.locals 3

					invoke-static {p1}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "no_page_turn_animation"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v1

					iput-boolean v1, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mNoPageTurnAnimation:Z

					return-void
				.end method
				`),
			),
			InMethod("<init>(Landroid/content/Context;Lcom/faultexception/reader/content/ContentView$ContentClient;Lcom/faultexception/reader/book/EPubBook;)V",
				ReplaceStringAppend(
					"\n"+"    invoke-direct {p0, p1}, Landroid/webkit/WebView;-><init>(Landroid/content/Context;)V",
					"\n"+"    invoke-direct {p0, p1}, Lcom/faultexception/reader/content/HtmlContentWebView;->initNoPageTurnAnimation(Landroid/content/Context;)V",
				),
			),
			InMethod("setPage(IZ)V", // the second argument is true if animating
				ReplaceStringAppend(
					`    .locals 5`,
					FixIndent("\n"+`
						iget-boolean v0, p0, Lcom/faultexception/reader/content/HtmlContentWebView;->mNoPageTurnAnimation:Z
						if-eqz v0, :page_turn_animation_ok
						const/4 p2, 0x0
						:page_turn_animation_ok
					`),
				),
			),
		),
		PatchFile("smali/com/faultexception/reader/widget/OverScrollView.smali",
			ReplaceStringAppend(
				"\n"+`.field private mVerticalRestricted:Z`,
				FixIndent("\n"+`
				.field private mNoPageTurnAnimation:Z

				.method private initNoPageTurnAnimation(Landroid/content/Context;)V
					.locals 3

					invoke-static {p1}, Landroid/preference/PreferenceManager;->getDefaultSharedPreferences(Landroid/content/Context;)Landroid/content/SharedPreferences;
					move-result-object v0

					const-string v1, "no_page_turn_animation"
					const/4 v2, 0x0
					invoke-interface {v0, v1, v2}, Landroid/content/SharedPreferences;->getBoolean(Ljava/lang/String;Z)Z
					move-result v1

					iput-boolean v1, p0, Lcom/faultexception/reader/widget/OverScrollView;->mNoPageTurnAnimation:Z

					return-void
				.end method
				`),
			),
			InMethod("<init>(Landroid/content/Context;Landroid/util/AttributeSet;)V",
				ReplaceStringAppend(
					"\n"+"    invoke-direct {p0, p1, p2}, Landroid/widget/FrameLayout;-><init>(Landroid/content/Context;Landroid/util/AttributeSet;)V",
					"\n"+"    invoke-direct {p0, p1}, Lcom/faultexception/reader/widget/OverScrollView;->initNoPageTurnAnimation(Landroid/content/Context;)V",
				),
			),
			InMethod("doOverScroll(I)V",
				// this is a terrible hack, but it works well enough and reduces code duplication
				MustContain(FixIndent("\n"+`
					.line 196
					:goto_0
					iget-object v2, p0, Lcom/faultexception/reader/widget/OverScrollView;->mOverScrollPullAnimation:Landroid/animation/ValueAnimator;

					new-instance v3, Lcom/faultexception/reader/widget/OverScrollView$2;

					invoke-direct {v3, p0, v0}, Lcom/faultexception/reader/widget/OverScrollView$2;-><init>(Lcom/faultexception/reader/widget/OverScrollView;I)V

					invoke-virtual {v2, v3}, Landroid/animation/ValueAnimator;->addListener(Landroid/animation/Animator$AnimatorListener;)V

					.line 207
					iget-object v0, p0, Lcom/faultexception/reader/widget/OverScrollView;->mOverScrollPullAnimation:Landroid/animation/ValueAnimator;
				`)),
				ReplaceStringPrepend(
					"\n"+`    .line 207`,
					FixIndent("\n"+`
						iget-boolean v0, p0, Lcom/faultexception/reader/widget/OverScrollView;->mNoPageTurnAnimation:Z
						if-eqz v0, :page_turn_animation_ok
						const v0, 0x0
						invoke-virtual {v3, v0}, Lcom/faultexception/reader/widget/OverScrollView$2;->onAnimationEnd(Landroid/animation/Animator;)V
						return-void
						:page_turn_animation_ok
					`),
				),
			),
		),
	)
}
