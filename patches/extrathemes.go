// # Extra themes
//
// Add additional built-in themes.
package patches

import (
	"io"

	. "github.com/pgaskin/lithiumpatch/patches/patchdef"
)

func init() {
	Register("extrathemes", extrathemes{
		/* sync_id                                dark     bg       text      link    name */
		{"60c8419a-ca5a-427f-b92a-79e69b2745bb", true, 0x000000, 0xC1A23D, 0xF0DE69, "Sepia Dark"},
		{"f20ec453-c903-4dcd-a61c-d06a1567c37b", true, 0x000000, 0x735E13, 0x7D7121, "Sepia Dark Dimmed"},
		//{"83ca1683-f1e2-4408-9afa-e896484ccdf2", true, 0x000000, 0x3D3105, 0x61560E, "Sepia Dark Extra Dimmed"},
		{"397903d7-b301-45c3-8064-fed7a2618e25", true, 0x1D2021, 0xEBDBB2, 0xFF705D, "Ash"},
		{"298fb36c-b91e-4e73-9ef5-3a23b0300666", true, 0x013151, 0xE1FFF0, 0x96F069, "Ocean"},
		{"3ef38676-5531-4351-b20a-a83359e8b546", true, 0xDDF3FF, 0x003193, 0x2D7FFF, "Ice"},
	})
}

type theme struct {
	SyncID     string
	Dark       bool
	Background uint32
	Text       uint32
	Link       uint32
	Name       string
}

type extrathemes []theme

func (ts extrathemes) Do(apk string, diffwriter io.Writer) error {
	builtinThemeIDs := []string{
		"04fd477e-bdbb-4dea-8f38-6bead547a00b",
		"f9715217-d3bb-41e3-974c-71e0ffaeee0b",
		"4948c360-f7cb-42b7-af6a-cf3431145f41",
	}

	// ThemesTable fixedSyncIds
	fixedSyncIDs := builtinThemeIDs
	for _, t := range ts {
		fixedSyncIDs = append(fixedSyncIDs, t.SyncID)
	}
	PatchFile("smali/com/faultexception/reader/db/ThemesTable.smali",
		InMethod("<clinit>()V",
			ReplaceString(
				FixIndent("\n"+`
					const/4 v1, 0x3

					new-array v1, v1, [Ljava/lang/String;

					const/4 v2, 0x0

					const-string v3, "04fd477e-bdbb-4dea-8f38-6bead547a00b"

					aput-object v3, v1, v2

					const/4 v2, 0x1

					const-string v3, "f9715217-d3bb-41e3-974c-71e0ffaeee0b"

					aput-object v3, v1, v2

					const/4 v2, 0x2

					const-string v3, "4948c360-f7cb-42b7-af6a-cf3431145f41"

					aput-object v3, v1, v2

					.line 37
					invoke-virtual {v0, v1}, Lcom/faultexception/reader/sync/SyncDataDefinition;->setFixedSyncIds([Ljava/lang/String;)Lcom/faultexception/reader/sync/SyncDataDefinition;
				`),
				FixIndent(ExecuteTemplate("\n"+`
					const v1, {{len .}}
					new-array v1, v1, [Ljava/lang/String;
					{{- range $i, $x := .}}
					const v2, {{$i}}
					const-string v3, "{{$x}}"
					aput-object v3, v1, v2
					invoke-virtual {v0, v1}, Lcom/faultexception/reader/sync/SyncDataDefinition;->setFixedSyncIds([Ljava/lang/String;)Lcom/faultexception/reader/sync/SyncDataDefinition;
					{{- end}}
				`, fixedSyncIDs)),
			),
		),
	).Do(apk, diffwriter)

	// DatabaseOpenHelper migrations
	// - 23 <= i < 30
	//   what: adds sync IDs to built-in themes, then adds any missing ones as hidden using ThemeManager.addBuiltinThemesAsHidden
	//   note: not needed for us since these themes have sync ids from the beginning
	// - i < 33
	//   what: removes duplicates of built-in themes using ThemeManager.removeDuplicatedBuiltinThemes
	//   note: not needed for us since I assume this was added due to a bug in an earlier version of the previous migration (82 based on similar logic in Backup.restoreThemes?)
	// - note: we don't need a new migration to add themes -- see my comment in updateCustomBuiltinThemes -- we just call our updateCustomBuiltinThemes before getting the themes

	// Backups restoreThemes
	// - note: logic (i.e., adding sync IDs) is basically what's in the DatabaseOpenHelper migrations
	// - note: we don't need to add our themes here -- see my comment in updateCustomBuiltinThemes -- we just call our updateCustomBuiltinThemes before getting the themes

	// ThemeManager removeDuplicatedBuiltinThemes
	// - loops over the builtin themes, deleting any rows for the default themes where said theme's sync id has already been seen
	// - called by Backups theme restore
	// - called by DatabaseOpenHelper migrations
	// - note: not needed for our custom themes (see my comments above)

	// ThemeManager addBuiltinThemesAsHidden
	// - adds each of the three built-in themes if their boolean parameter is set
	// - called by Backups theme restore
	// - called by DatabaseOpenHelper migrations
	// - note: not needed for our custom themes (see my comments above)

	// ThemeManager createBuiltinThemes
	// - creates all builtin themes, sets modified to current date if bool param is true
	// - called by DatabaseOpenHelper initial creation and migrations
	// - called by ThemeManager resetToDefaults -- which is called by ProManager and ThemesActivity)
	PatchFile("smali/com/faultexception/reader/themes/ThemeManager.smali",
		InMethod("createBuiltinThemes(Landroid/database/sqlite/SQLiteDatabase;Z)V",
			ReplaceString(
				FixIndent("\n"+`
					invoke-virtual/range {p0 .. p0}, Landroid/database/sqlite/SQLiteDatabase;->endTransaction()V

					return-void
				`),
				FixIndent("\n"+`
					invoke-virtual/range {p0 .. p0}, Landroid/database/sqlite/SQLiteDatabase;->endTransaction()V
					invoke-static/range {p0 .. p0}, Lcom/faultexception/reader/themes/ThemeManager;->updateCustomBuiltinThemes(Landroid/database/sqlite/SQLiteDatabase;)V
					return-void
				`),
			),
		),
	).Do(apk, diffwriter)

	// ThemeManager updateCustomBuiltinThemes
	// - new method to add custom builtin themes if they're not already there, and to update the colors if they haven't been modified by the user
	// - to update themes when not creating/upgrading the database nor resetting themes, we could either put it in DatabaseOpenHelper.onOpen, or ThemeManager.getThemes...
	//   former may have a minor performance impact on app start, latter when inflating display settings (once per ReaderActivity on first display settings open) or starting theme manager...
	//   probably better with the latter
	// - column values are based on createBuiltinThemes
	// - do an insert to add the rows for the theme sync ids, then do an update to ensure they have the latest colors
	// - will preserve manual theme deletions (i.e., hides), edits, or renames
	// - will delete added custom builtin themes if removed from this file (even if edited by the user)
	// - note: position may overlap with user themes if adding custom themes after init, but this isn't a critical issue (sort order is still consistent since it's based on position then creation time)
	PatchFile("smali/com/faultexception/reader/themes/ThemeManager.smali",
		ReplaceStringPrepend(
			FixIndent("\n"+`
			.method public static createBuiltinThemes(Landroid/database/sqlite/SQLiteDatabase;Z)V
			`),
			FixIndent(ExecuteTemplate("\n"+`
			.method public static updateCustomBuiltinThemes(Landroid/database/sqlite/SQLiteDatabase;)V
				.locals 6
				move-object v0, p0
				invoke-virtual {v0}, Landroid/database/sqlite/SQLiteDatabase;->beginTransaction()V
				:try1_try
				invoke-static {}, Ljava/lang/System;->currentTimeMillis()J
				move-result-wide v3
				invoke-static {v3, v4}, Ljava/lang/String;->valueOf(J)Ljava/lang/String;
				move-result-object v3
				const-string v2, "NOW"
				{{- range $i, $t := .Themes }}
				const-string v1, "INSERT INTO themes (_sync_id, name, builtin, hidden, position, created_date, modified_date, bg_color_timestamp, text_color_timestamp, link_color_timestamp, use_dark_chrome_timestamp) SELECT '{{.SyncID}}', '{{.Name}}', 1, 0, {{AddInt 3 $i}}, NOW, NOW, NOW, NOW, NOW, NOW WHERE NOT EXISTS (SELECT _sync_id FROM themes WHERE _sync_id = '{{.SyncID}}')"
				invoke-virtual {v1, v2, v3}, Ljava/lang/String;->replace(Ljava/lang/CharSequence;Ljava/lang/CharSequence;)Ljava/lang/String;
				move-result-object v1
				invoke-virtual {v0, v1}, Landroid/database/sqlite/SQLiteDatabase;->execSQL(Ljava/lang/String;)V
				{{- end }}
				{{- range $i, $t := .Themes }}
				const-string v1, "UPDATE themes SET position = {{AddInt 3 $i}}, bg_color = {{.Background}}, text_color = {{.Text}}, link_color = {{.Link}}, use_dark_chrome = {{if .Dark}}1{{else}}0{{end}} WHERE _sync_id = '{{.SyncID}}' AND created_date = bg_color_timestamp AND created_date = text_color_timestamp AND created_date = link_color_timestamp AND created_date = use_dark_chrome_timestamp;"
				invoke-virtual {v0, v1}, Landroid/database/sqlite/SQLiteDatabase;->execSQL(Ljava/lang/String;)V
				{{- end }}
				const-string v1, "DELETE FROM themes WHERE builtin = 1 AND _sync_id NOT IN ({{range $i, $x := .SyncIDs}}{{if $i}},{{end}}'{{.}}'{{end}});"
				invoke-virtual {v0, v1}, Landroid/database/sqlite/SQLiteDatabase;->execSQL(Ljava/lang/String;)V
				invoke-virtual {v0}, Landroid/database/sqlite/SQLiteDatabase;->setTransactionSuccessful()V
				:try1_end
				.catchall {:try1_try .. :try1_end} :try1_catch
				invoke-virtual {v0}, Landroid/database/sqlite/SQLiteDatabase;->endTransaction()V
				return-void
				:try1_catch
				move-exception v1
				invoke-virtual {v0}, Landroid/database/sqlite/SQLiteDatabase;->endTransaction()V
				throw v1
			.end method
			`, map[string]any{"Themes": ts, "SyncIDs": fixedSyncIDs})),
		),
		InMethod("getThemes()Ljava/util/List;",
			ReplaceStringAppend(
				FixIndent("\n"+`
					.end annotation

					.line 58
					iget-object v0, p0, Lcom/faultexception/reader/themes/ThemeManager;->mDb:Landroid/database/sqlite/SQLiteDatabase;
				`),
				FixIndent("\n"+`
					invoke-static {v0}, Lcom/faultexception/reader/themes/ThemeManager;->updateCustomBuiltinThemes(Landroid/database/sqlite/SQLiteDatabase;)V
				`),
			),
		),
	).Do(apk, diffwriter)

	return nil
}
