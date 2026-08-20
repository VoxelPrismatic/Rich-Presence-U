package ui

import (
	"fmt"

	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/nso"
)

func (a *App) buildSettings() *qt6.QWidget {
	page := qt6.NewQWidget2()
	box := qt6.NewQVBoxLayout(page)
	box.SetContentsMargins(16, 12, 16, 12)

	head := qt6.NewQHBoxLayout2()
	back := qt6.NewQPushButton4(iconNamed("go-previous", "go-previous"), a.tr.T("SETTINGS_BACK"))
	back.OnClicked(func() { a.stack.SetCurrentIndex(0) })
	head.AddWidget(back.QWidget)
	head.AddStretch()
	box.AddLayout(head.QLayout)

	form := qt6.NewQWidget2()
	g := newFormGrid(form)
	row := 0

	a.langCombo = qt6.NewQComboBox2()
	a.langCombo.AddItem3(a.tr.T("LANGUAGE_AUTO"), qt6.NewQVariant11(""))
	for _, code := range localeCodes() {
		a.langCombo.AddItem3(localeNames[code], qt6.NewQVariant11(code))
	}
	a.langCombo.OnCurrentIndexChanged(func(i int) {
		if a.silent {
			return
		}
		a.settings.Language = a.langCombo.ItemData(i).ToString()
		a.tr.Set(a.settings.Language)
	})
	addFormRow(g, row, a.tr.T("LANGUAGE_TITLE"), a.langCombo.QWidget, nil)
	row++

	a.prefRegion = qt6.NewQComboBox2()
	a.prefRegion.AddItem3(a.tr.T("REGION_US"), qt6.NewQVariant11("US"))
	a.prefRegion.AddItem3(a.tr.T("REGION_EU"), qt6.NewQVariant11("EU"))
	a.prefRegion.AddItem3(a.tr.T("REGION_JP"), qt6.NewQVariant11("JP"))
	a.prefRegion.OnCurrentIndexChanged(func(i int) {
		if a.silent {
			return
		}
		if r, ok := nso.ParseRegion(a.prefRegion.ItemData(i).ToString()); ok {
			a.settings.Region = r
			a.refillCompleter()
			a.refreshGameUI()
			a.updateApply()
		}
	})
	addFormRow(g, row, a.tr.T("REGION_TITLE"), a.prefRegion.QWidget, nil)
	row++

	refreshWrap := qt6.NewQWidget2()
	rw := qt6.NewQHBoxLayout(refreshWrap)
	rw.SetContentsMargins(0, 0, 0, 0)
	now := qt6.NewQPushButton2()
	now.SetIcon(iconNamed("view-refresh", "view-refresh"))
	now.SetToolTip(a.tr.T("REFRESH_NOW"))
	now.OnClicked(func() { a.refreshTitles(true) })
	a.refreshCombo = qt6.NewQComboBox2()
	a.refreshCombo.AddItem3(a.tr.T("REFRESH_12H"), qt6.NewQVariant4(43200))
	a.refreshCombo.AddItem3(a.tr.T("REFRESH_DAILY"), qt6.NewQVariant4(86400))
	a.refreshCombo.AddItem3(a.tr.T("REFRESH_WEEKLY"), qt6.NewQVariant4(604800))
	a.refreshCombo.AddItem3(a.tr.T("REFRESH_DISABLED"), qt6.NewQVariant4(-1))
	a.refreshCombo.OnCurrentIndexChanged(func(i int) {
		if a.silent {
			return
		}
		a.settings.Refresh = a.refreshCombo.ItemData(i).ToInt()
	})
	rw.AddWidget(now.QWidget)
	rw.AddWidget(a.refreshCombo.QWidget)
	rw.AddStretch()
	addFormRow(g, row, a.tr.T("REFRESH_TITLE"), refreshWrap, nil)
	row++

	a.autoConn = qt6.NewQCheckBox2()
	a.autoConn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.AutoConnect = on
		}
	})
	addFormRow(g, row, a.tr.T("AUTOCONNECT_TITLE"), a.autoConn.QWidget, helpButton(a.tr.T("AUTOCONNECT_HINT")))
	row++

	a.keepOn = qt6.NewQCheckBox2()
	a.keepOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.KeepOn = on
			a.refreshScreensaver()
		}
	})
	addFormRow(g, row, a.tr.T("KEEPON_TITLE"), a.keepOn.QWidget, helpButton(a.tr.T("KEEPON_HINT")))
	row++

	a.debugOn = qt6.NewQCheckBox2()
	a.debugOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.DebugLog = on
			a.log.SetEnabled(on)
		}
	})
	addFormRow(g, row, a.tr.T("DEBUG_TITLE"), a.debugOn.QWidget, helpButton(a.tr.T("DEBUG_HINT")))
	row++

	dataWrap := qt6.NewQWidget2()
	dw := qt6.NewQHBoxLayout(dataWrap)
	dw.SetContentsMargins(0, 0, 0, 0)
	a.dataCombo = qt6.NewQComboBox2()
	a.dataCombo.AddItem4(iconNamed("action-unavailable", "action-unavailable-symbolic"), a.tr.T("DATA_SELECT"), qt6.NewQVariant11(""))
	a.dataCombo.AddItem4(iconNamed("edit-clear-history", "edit-clear-history"), a.tr.T("RESET_CACHE_TITLE"), qt6.NewQVariant11("cache"))
	a.dataCombo.AddItem4(iconNamed("user-trash", "albumfolder-user-trash"), a.tr.T("RESET_ALL_TITLE"), qt6.NewQVariant11("all"))
	a.dataBtn = qt6.NewQPushButton2()
	a.dataBtn.SetIcon(iconNamed("action-unavailable", "action-unavailable-symbolic"))
	a.dataCombo.OnCurrentIndexChanged(func(i int) {
		a.dataBtn.SetIcon(a.dataCombo.ItemIcon(i))
	})
	a.dataBtn.OnClicked(func() { a.onDataAction() })
	dw.AddWidget(a.dataCombo.QWidget)
	dw.AddWidget(a.dataBtn.QWidget)
	dw.AddStretch()
	addFormRow(g, row, a.tr.T("DATA_TITLE"), dataWrap, helpButton(a.tr.T("DATA_HINT")))

	box.AddWidget(form)
	box.AddWidget(hline().QWidget)
	box.AddWidget(a.buildAboutTable())
	box.AddStretch()

	a.fillAboutLinks()
	return page
}

func (a *App) buildAboutTable() *qt6.QWidget {
	t := qt6.NewQTableWidget3(4, 2)
	a.aboutTable = t
	t.SetShowGrid(true)
	t.SetFocusPolicy(qt6.NoFocus)
	t.SetSelectionMode(qt6.QAbstractItemView__NoSelection)
	t.SetEditTriggers(qt6.QAbstractItemView__NoEditTriggers)
	t.SetHorizontalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	t.SetVerticalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	t.VerticalHeader().Hide()
	t.HorizontalHeader().Hide()
	t.HorizontalHeader().SetSectionResizeMode2(0, qt6.QHeaderView__ResizeToContents)
	t.HorizontalHeader().SetSectionResizeMode2(1, qt6.QHeaderView__Stretch)
	t.SetSizePolicy2(qt6.QSizePolicy__Expanding, qt6.QSizePolicy__Minimum)
	return t.QWidget
}

func aboutKey(text string) *qt6.QTableWidgetItem {
	it := qt6.NewQTableWidgetItem2(text)
	it.SetFlags(qt6.ItemIsEnabled)
	return it
}

func aboutLink(text, url string) *qt6.QLabel {
	l := linkLabel(fmt.Sprintf(`<a href="%s">%s</a>`, url, text))
	l.SetMargin(4)
	return l
}

func aboutInfoCell(blogURL, helpURL string) *qt6.QWidget {
	w := qt6.NewQWidget2()
	lay := qt6.NewQHBoxLayout(w)
	lay.SetContentsMargins(4, 0, 4, 0)
	lay.SetSpacing(8)
	lay.AddWidget(aboutLink("Blog", blogURL).QWidget)
	lay.AddWidget(vline().QWidget)
	lay.AddWidget(aboutLink("Help", helpURL).QWidget)
	lay.AddStretch()
	return w
}

func (a *App) fillAboutLinks() {
	t := a.aboutTable
	if t == nil {
		return
	}
	meta := a.nso.Meta
	home := meta.Home
	if home == "" {
		home = "https://ninstars.blogspot.com/rpc"
	}
	help := meta.HelpURL(a.tr.Lang())
	changelog := meta.BinURL
	if changelog == "" {
		changelog = home
	}

	t.ClearContents()
	t.SetItem(0, 0, aboutKey(a.tr.T("ABOUT_VERSION")))
	t.SetCellWidget(0, 1, aboutLink(Version, changelog).QWidget)

	t.SetItem(1, 0, aboutKey(a.tr.T("ABOUT_INFO")))
	t.SetCellWidget(1, 1, aboutInfoCell(home, help))

	t.SetItem(2, 0, aboutKey(a.tr.T("ABOUT_CORE")))
	t.SetCellWidget(2, 1, aboutLink("NinStar", "https://github.com/ninstar/Rich-Presence-U").QWidget)

	t.SetItem(3, 0, aboutKey(a.tr.T("ABOUT_QT")))
	t.SetCellWidget(3, 1, aboutLink("VoxelPrismatic", "https://github.com/VoxelPrismatic/Rich-Presence-U").QWidget)

	t.ResizeRowsToContents()
	t.ResizeColumnToContents(0)
	h := t.FrameWidth() * 2
	for i := 0; i < t.RowCount(); i++ {
		h += t.RowHeight(i)
	}
	t.SetFixedHeight(h + 2)
}

func (a *App) loadSettingsIntoUI() {
	a.silent = true
	a.selectComboData(a.langCombo, a.settings.Language)
	a.selectComboData(a.prefRegion, string(a.settings.Region))
	found := false
	for i := 0; i < a.refreshCombo.Count(); i++ {
		if a.refreshCombo.ItemData(i).ToInt() == a.settings.Refresh {
			a.refreshCombo.SetCurrentIndex(i)
			found = true
			break
		}
	}
	if !found {
		a.refreshCombo.SetCurrentIndex(2)
	}
	a.autoConn.SetChecked(a.settings.AutoConnect)
	a.keepOn.SetChecked(a.settings.KeepOn)
	a.debugOn.SetChecked(a.settings.DebugLog)
	a.silent = false
}
