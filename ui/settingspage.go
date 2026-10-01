package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/nso"
	"github.com/voxelprismatic/richpresenceu/svc"
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

	_, locLay := newSettingsPane(a.tr.T("SETTINGS_LOCALIZATION"))
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
	addSettingsField(locLay, a.tr.T("LANGUAGE_TITLE"), a.langCombo.QWidget, nil)

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
			a.refillGameCombo()
			a.refreshGameUI()
			a.updateApply()
		}
	})
	addSettingsField(locLay, a.tr.T("REGION_TITLE"), a.prefRegion.QWidget, nil)

	_, presLay := newSettingsPane(a.tr.T("SETTINGS_PRESENCE"))
	a.autoConn = qt6.NewQCheckBox2()
	a.autoConn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.AutoConnect = on
		}
	})
	addSettingsCheck(presLay, a.tr.T("AUTOCONNECT_TITLE"), a.autoConn, helpButton(a.tr.T("AUTOCONNECT_HINT")))

	a.keepOn = qt6.NewQCheckBox2()
	a.keepOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.KeepOn = on
			a.refreshScreensaver()
		}
	})
	addSettingsCheck(presLay, a.tr.T("KEEPON_TITLE"), a.keepOn, helpButton(a.tr.T("KEEPON_HINT")))

	a.debugOn = qt6.NewQCheckBox2()
	a.debugOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.DebugLog = on
			a.log.SetEnabled(on)
		}
	})
	addSettingsCheck(presLay, a.tr.T("DEBUG_TITLE"), a.debugOn, helpButton(a.tr.T("DEBUG_HINT")))

	dataWrap := qt6.NewQWidget2()
	dw := qt6.NewQHBoxLayout(dataWrap)
	dw.SetContentsMargins(0, 0, 0, 0)
	a.dataCombo = qt6.NewQComboBox2()
	a.dataCombo.SetSizePolicy2(qt6.QSizePolicy__Expanding, qt6.QSizePolicy__Fixed)
	a.dataCombo.AddItem4(iconNamed("action-unavailable", "action-unavailable-symbolic"), a.tr.T("DATA_SELECT"), qt6.NewQVariant11(""))
	a.dataCombo.AddItem4(iconNamed("edit-clear-history", "edit-clear-history"), a.tr.T("RESET_CACHE_TITLE"), qt6.NewQVariant11("cache"))
	a.dataCombo.AddItem4(iconNamed("user-trash", "albumfolder-user-trash"), a.tr.T("RESET_ALL_TITLE"), qt6.NewQVariant11("all"))
	a.dataBtn = qt6.NewQPushButton2()
	a.dataBtn.SetIcon(iconNamed("action-unavailable", "action-unavailable-symbolic"))
	a.dataCombo.OnCurrentIndexChanged(func(i int) {
		a.dataBtn.SetIcon(a.dataCombo.ItemIcon(i))
	})
	a.dataBtn.OnClicked(func() { a.onDataAction() })
	dw.AddWidget2(a.dataCombo.QWidget, 1)
	dw.AddWidget(a.dataBtn.QWidget)
	addSettingsField(presLay, a.tr.T("DATA_TITLE"), dataWrap, helpButton(a.tr.T("DATA_HINT")))

	_, igdbLay := newSettingsPane(a.tr.T("SETTINGS_GAME_SEARCH"))
	steps := linkLabel(a.tr.T("IGDB_INSTRUCTIONS"))
	steps.SetWordWrap(true)
	steps.SetAlignment(qt6.AlignLeft | qt6.AlignTop)
	igdbLay.AddWidget(steps.QWidget)
	a.igdbID = qt6.NewQLineEdit2()
	a.igdbID.SetPlaceholderText(a.tr.T("IGDB_CLIENT_ID"))
	a.igdbID.OnEditingFinished(func() { a.saveIGDBCredentials() })
	a.igdbSecret = qt6.NewQLineEdit2()
	a.igdbSecret.SetPlaceholderText(a.tr.T("IGDB_CLIENT_SECRET"))
	a.igdbSecret.SetEchoMode(qt6.QLineEdit__Password)
	a.igdbSecret.OnEditingFinished(func() { a.saveIGDBCredentials() })
	igdbLay.AddWidget(a.igdbID.QWidget)
	igdbLay.AddWidget(a.igdbSecret.QWidget)
	igdbBtns := qt6.NewQHBoxLayout2()
	igdbBtns.SetContentsMargins(0, 0, 0, 0)
	testBtn := qt6.NewQPushButton3(a.tr.T("IGDB_TEST"))
	testBtn.OnClicked(func() { a.testIGDB() })
	consoleBtn := qt6.NewQPushButton3(a.tr.T("IGDB_CONSOLE"))
	consoleBtn.OnClicked(func() { openURL("https://dev.twitch.tv/console/apps") })
	igdbBtns.AddWidget(testBtn.QWidget)
	igdbBtns.AddWidget(consoleBtn.QWidget)
	igdbBtns.AddStretch()
	igdbLay.AddLayout(igdbBtns.QLayout)

	_, creditsLay := newSettingsPane(a.tr.T("ABOUT_CREDITS"))
	creditsLay.AddWidget(a.buildAboutTable())

	locLay.AddStretch()
	presLay.AddStretch()
	igdbLay.AddStretch()
	creditsLay.AddStretch()

	names := []string{
		a.tr.T("SETTINGS_LOCALIZATION"),
		a.tr.T("SETTINGS_PRESENCE"),
		a.tr.T("SETTINGS_GAME_SEARCH"),
		a.tr.T("ABOUT_CREDITS"),
	}
	panes := []*qt6.QVBoxLayout{locLay, presLay, igdbLay, creditsLay}
	list := qt6.NewQListWidget2()
	list.SetFixedWidth(220)
	list.SetHorizontalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	list.SetSizePolicy2(qt6.QSizePolicy__Fixed, qt6.QSizePolicy__Expanding)
	pages := qt6.NewQStackedWidget2()
	for i, name := range names {
		list.AddItem(name)
		pages.AddWidget(settingsScroll(panes[i]).QWidget)
	}
	list.SetCurrentRow(0)
	list.OnCurrentRowChanged(func(row int) {
		if row >= 0 {
			pages.SetCurrentIndex(row)
		}
	})

	split := qt6.NewQHBoxLayout2()
	split.SetContentsMargins(0, 0, 0, 0)
	split.SetSpacing(12)
	split.AddWidget(list.QWidget)
	split.AddWidget2(pages.QWidget, 1)
	box.AddLayout2(split.QLayout, 1)

	a.fillAboutLinks()
	return page
}

func newSettingsPane(title string) (*qt6.QWidget, *qt6.QVBoxLayout) {
	w := qt6.NewQWidget2()
	lay := qt6.NewQVBoxLayout(w)
	lay.SetContentsMargins(12, 4, 8, 8)
	lay.SetSpacing(8)
	head := qt6.NewQLabel3(title)
	setBigFont(head.QWidget, 2)
	lay.AddWidget(head.QWidget)
	return w, lay
}

func settingsScroll(lay *qt6.QVBoxLayout) *qt6.QScrollArea {
	scroll := qt6.NewQScrollArea2()
	scroll.SetWidgetResizable(true)
	scroll.SetFrameShape(qt6.QFrame__NoFrame)
	scroll.SetWidget(lay.ParentWidget())
	return scroll
}

func addSettingsField(parent *qt6.QVBoxLayout, title string, control *qt6.QWidget, help *qt6.QToolButton) {
	row := qt6.NewQHBoxLayout2()
	row.SetContentsMargins(0, 0, 0, 0)
	row.AddWidget(qt6.NewQLabel3(title).QWidget)
	if help != nil {
		row.AddWidget(help.QWidget)
	}
	row.AddStretch()
	parent.AddLayout(row.QLayout)
	parent.AddWidget(control)
}

func addSettingsCheck(parent *qt6.QVBoxLayout, title string, box *qt6.QCheckBox, help *qt6.QToolButton) {
	box.SetText(title)
	row := qt6.NewQHBoxLayout2()
	row.SetContentsMargins(0, 0, 0, 0)
	row.AddWidget(box.QWidget)
	if help != nil {
		row.AddWidget(help.QWidget)
	}
	row.AddStretch()
	parent.AddLayout(row.QLayout)
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
	changelog := meta.BinURL
	if changelog == "" {
		changelog = home
	}

	t.ClearContents()
	t.SetItem(0, 0, aboutKey(a.tr.T("ABOUT_VERSION")))
	t.SetCellWidget(0, 1, aboutLink(svc.VERSION, changelog).QWidget)

	t.SetItem(1, 0, aboutKey(a.tr.T("ABOUT_INFO")))
	t.SetCellWidget(1, 1, aboutLink(a.tr.T("ABOUT_BLOG"), home).QWidget)

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
	a.autoConn.SetChecked(a.settings.AutoConnect)
	a.keepOn.SetChecked(a.settings.KeepOn)
	a.debugOn.SetChecked(a.settings.DebugLog)
	if a.igdbID != nil {
		a.igdbID.SetText(a.settings.IGDBClientID)
	}
	if a.igdbSecret != nil {
		a.igdbSecret.SetText(a.settings.IGDBClientSecret)
	}
	a.silent = false
}

func (a *App) saveIGDBCredentials() {
	if a.silent {
		return
	}
	id, secret := "", ""
	if a.igdbID != nil {
		id = strings.TrimSpace(a.igdbID.Text())
	}
	if a.igdbSecret != nil {
		secret = strings.TrimSpace(a.igdbSecret.Text())
	}
	a.settings.IGDBClientID = id
	a.settings.IGDBClientSecret = secret
	if a.igdbAPI != nil {
		a.igdbAPI.SetCredentials(id, secret)
	}
	a.persist()
}

func (a *App) testIGDB() {
	a.saveIGDBCredentials()
	if a.igdbAPI == nil || !a.igdbAPI.Configured() {
		qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("IGDB_TITLE"), popupText(a.tr.T("IGDB_TEST_FAIL")))
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		err := a.igdbAPI.Ping(ctx)
		mainthread.Start(func() {
			if err != nil {
				a.debug("igdb ping: %v", err)
				qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("IGDB_TITLE"), popupText(a.tr.T("IGDB_TEST_FAIL")))
				return
			}
			qt6.QMessageBox_Information(a.win.QWidget, a.tr.T("IGDB_TITLE"), popupText(a.tr.T("IGDB_TEST_OK")))
		})
	}()
}
