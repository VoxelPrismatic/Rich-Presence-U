package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/igdb"
	"github.com/voxelprismatic/richpresenceu/locales"
	"github.com/voxelprismatic/richpresenceu/nso"
	"github.com/voxelprismatic/richpresenceu/svc"
)

func (a *App) buildSettings() *qt6.QWidget {
	page := qt6.NewQWidget2()
	box := qt6.NewQVBoxLayout(page)
	box.SetContentsMargins(16, 12, 16, 12)

	head := qt6.NewQHBoxLayout2()
	back := qt6.NewQPushButton4(iconNamed("go-previous", "go-previous"), a.tr.T("SETTINGS_BACK"))
	a.settingsBack = back
	back.OnClicked(func() { a.stack.SetCurrentIndex(0) })
	head.AddWidget(back.QWidget)
	head.AddStretch()
	box.AddLayout(head.QLayout)

	_, locLay := newSettingsPane(a.tr.T("SETTINGS_LOCALIZATION"))
	locForm := settingsForm(locLay)
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
	addFormRow(locForm, 0, a.tr.T("LANGUAGE_TITLE"), a.langCombo.QWidget, nil)

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
	addFormRow(locForm, 1, a.tr.T("REGION_TITLE"), a.prefRegion.QWidget, nil)

	_, presLay := newSettingsPane(a.tr.T("SETTINGS_PRESENCE"))
	presForm := settingsForm(presLay)
	a.autoConn = qt6.NewQComboBox2()
	addAutoConnectItem(a.autoConn, a.tr.T("AUTOCONNECT_NEVER"), string(connectNever), a.tr.T("AUTOCONNECT_HINT_NEVER"))
	addAutoConnectItem(a.autoConn, a.tr.T("AUTOCONNECT_STARTUP"), string(connectStartup), a.tr.T("AUTOCONNECT_HINT_STARTUP"))
	addAutoConnectItem(a.autoConn, a.tr.T("AUTOCONNECT_POLL"), string(connectPoll), a.tr.T("AUTOCONNECT_HINT_POLL"))
	if view := a.autoConn.View(); view != nil {
		view.SetMouseTracking(true)
	}
	a.autoConnHelp = a.helpButton("")
	a.autoConn.OnCurrentIndexChanged(func(i int) {
		a.refreshAutoConnectHint()
		if a.silent {
			return
		}
		a.settings.AutoConnect = connectMode(a.autoConn.ItemData(i).ToString())
		a.syncConnectPoll()
	})
	a.autoConn.OnHighlighted(func(i int) {
		if i < 0 {
			return
		}
		a.autoConn.SetToolTip(a.tr.T(autoConnectHint(connectMode(a.autoConn.ItemData(i).ToString()))))
	})
	a.autoConn.OnHidePopup(func(super func()) {
		super()
		a.refreshAutoConnectHint()
	})
	addFormRow(presForm, 0, a.tr.T("AUTOCONNECT_TITLE"), a.autoConn.QWidget, a.autoConnHelp)

	a.keepOn = qt6.NewQCheckBox2()
	a.keepOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.KeepOn = on
			a.refreshScreensaver()
		}
	})
	addFormRow(presForm, 1, a.tr.T("KEEPON_TITLE"), a.keepOn.QWidget, a.helpButton(a.tr.T("KEEPON_HINT")))

	a.debugOn = qt6.NewQCheckBox2()
	a.debugOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.DebugLog = on
			a.log.SetEnabled(on)
		}
	})
	addFormRow(presForm, 2, a.tr.T("DEBUG_TITLE"), a.debugOn.QWidget, a.helpButton(a.tr.T("DEBUG_HINT")))

	hideWrap := qt6.NewQWidget2()
	hideLay := qt6.NewQVBoxLayout(hideWrap)
	hideLay.SetContentsMargins(0, 0, 0, 0)
	hideLay.SetSpacing(2)
	a.hidePause = qt6.NewQCheckBox4(a.tr.T("HIDE_PAUSE"), hideWrap)
	a.hideDiscord = qt6.NewQCheckBox4(a.tr.T("HIDE_DISCORD"), hideWrap)
	a.autoUnhide = qt6.NewQCheckBox4(a.tr.T("HIDE_AUTO_UNHIDE"), hideWrap)
	a.hidePause.OnToggled(func(on bool) { a.setPauseTimer(on) })
	a.hideDiscord.OnToggled(func(on bool) { a.setHideDiscord(on) })
	a.autoUnhide.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.AutoUnhide = on
		}
	})
	hideLay.AddWidget(a.hidePause.QWidget)
	hideLay.AddWidget(a.hideDiscord.QWidget)
	hideLay.AddWidget(a.autoUnhide.QWidget)
	addFormRow(presForm, 3, a.tr.T("HIDE_BEHAVIOR"), hideWrap, nil)

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
	addFormRow(presForm, 4, a.tr.T("DATA_TITLE"), dataWrap, a.helpButton(a.tr.T("DATA_HINT")))

	_, igdbLay := newSettingsPane(a.tr.T("SETTINGS_GAME_SEARCH"))
	intro := qt6.NewQLabel3(a.tr.T("IGDB_INTRO"))
	intro.SetWordWrap(true)
	igdbLay.AddWidget(intro.QWidget)
	steps := linkLabel(locales.GameSearchInstructions(a.tr.T("IGDB_INSTRUCTIONS"), ""))
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
	a.igdbStatus = qt6.NewQLabel3(a.tr.T("IGDB_STATUS_UNTESTED"))
	igdbBtns.AddWidget(testBtn.QWidget)
	igdbBtns.AddWidget(a.igdbStatus.QWidget)
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

func addAutoConnectItem(combo *qt6.QComboBox, label, value, hint string) {
	combo.AddItem3(label, qt6.NewQVariant11(value))
	combo.SetItemData2(combo.Count()-1, qt6.NewQVariant11(hint), int(qt6.ToolTipRole))
}

func (a *App) refreshAutoConnectHint() {
	if a.autoConn == nil {
		return
	}
	mode := connectMode(a.autoConn.ItemData(a.autoConn.CurrentIndex()).ToString())
	tip := a.tr.T(autoConnectHint(mode))
	a.autoConn.SetToolTip(tip)
	if a.autoConnHelp != nil {
		a.autoConnHelp.SetToolTip(tip)
	}
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

func settingsForm(parent *qt6.QVBoxLayout) *qt6.QGridLayout {
	g := newFormGrid(nil)
	parent.AddLayout(g.QLayout)
	return g
}

func settingsScroll(lay *qt6.QVBoxLayout) *qt6.QScrollArea {
	scroll := qt6.NewQScrollArea2()
	scroll.SetWidgetResizable(true)
	scroll.SetFrameShape(qt6.QFrame__NoFrame)
	scroll.SetWidget(lay.ParentWidget())
	return scroll
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
	a.selectComboData(a.autoConn, string(a.settings.AutoConnect))
	a.keepOn.SetChecked(a.settings.KeepOn)
	a.debugOn.SetChecked(a.settings.DebugLog)
	if a.hidePause != nil {
		a.hidePause.SetChecked(a.settings.PauseTimer)
	}
	if a.hideDiscord != nil {
		a.hideDiscord.SetChecked(a.settings.HideDiscord)
	}
	if a.autoUnhide != nil {
		a.autoUnhide.SetChecked(a.settings.AutoUnhide)
	}
	a.refreshAutoConnectHint()
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
	if id == a.settings.IGDBClientID && secret == a.settings.IGDBClientSecret {
		return
	}
	a.settings.IGDBClientID = id
	a.settings.IGDBClientSecret = secret
	if a.igdbAPI != nil {
		a.igdbAPI.SetCredentials(id, secret)
	}
	a.setIGDBStatus(a.tr.T("IGDB_STATUS_UNTESTED"))
	a.persist()
}

func (a *App) setIGDBStatus(text string) {
	if a.igdbStatus != nil {
		a.igdbStatus.SetText(text)
	}
}

func (a *App) igdbFailText(err error) string {
	switch igdb.Classify(err) {
	case igdb.FailTimeout:
		return a.tr.T("IGDB_TEST_TIMEOUT")
	case igdb.FailUnauthorized:
		return a.tr.T("IGDB_TEST_UNAUTHORIZED")
	default:
		return a.tr.T("IGDB_TEST_FAIL")
	}
}

func (a *App) finishIGDBTest(err error) {
	if err == nil {
		a.setIGDBStatus(a.tr.T("IGDB_STATUS_OK"))
		return
	}
	a.debug("igdb ping: %v", err)
	a.setIGDBStatus(a.tr.T("IGDB_STATUS_FAIL"))
	qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("IGDB_TITLE"), popupText(a.igdbFailText(err)))
}

func (a *App) testIGDB() {
	a.saveIGDBCredentials()
	if a.igdbAPI == nil {
		a.finishIGDBTest(igdb.ErrNoCredentials)
		return
	}
	a.igdbTestGen++
	gen := a.igdbTestGen
	api := a.igdbAPI
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		err := api.Ping(ctx)
		mainthread.Start(func() {
			if gen != a.igdbTestGen {
				return
			}
			a.finishIGDBTest(err)
		})
	}()
}
