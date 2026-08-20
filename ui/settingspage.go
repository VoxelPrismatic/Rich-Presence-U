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
	a.addSettingRow(box, a.tr.T("LANGUAGE_TITLE"), nil, a.langCombo.QWidget)

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
	a.addSettingRow(box, a.tr.T("REGION_TITLE"), nil, a.prefRegion.QWidget)

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
	a.addSettingRow(box, a.tr.T("REFRESH_TITLE"), nil, refreshWrap)

	a.autoConn = qt6.NewQCheckBox2()
	a.autoConn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.AutoConnect = on
		}
	})
	a.addSettingRow(box, a.tr.T("AUTOCONNECT_TITLE"), helpButton(a.tr.T("AUTOCONNECT_HINT")), a.autoConn.QWidget)

	a.keepOn = qt6.NewQCheckBox2()
	a.keepOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.KeepOn = on
			a.refreshScreensaver()
		}
	})
	a.addSettingRow(box, a.tr.T("KEEPON_TITLE"), helpButton(a.tr.T("KEEPON_HINT")), a.keepOn.QWidget)

	a.debugOn = qt6.NewQCheckBox2()
	a.debugOn.OnToggled(func(on bool) {
		if !a.silent {
			a.settings.DebugLog = on
			a.log.SetEnabled(on)
		}
	})
	a.addSettingRow(box, a.tr.T("DEBUG_TITLE"), helpButton(a.tr.T("DEBUG_HINT")), a.debugOn.QWidget)

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
	a.addSettingRow(box, a.tr.T("DATA_TITLE"), helpButton(a.tr.T("DATA_HINT")), dataWrap)

	box.AddWidget(hline().QWidget)

	a.verLabel = qt6.NewQLabel2()
	a.infoLabel = qt6.NewQLabel2()
	box.AddWidget(a.verLabel.QWidget)
	box.AddWidget(a.infoLabel.QWidget)
	core := linkLabel(fmt.Sprintf("%s | <a href=\"https://github.com/ninstar/Rich-Presence-U\">NinStar</a>", a.tr.T("ABOUT_CORE")))
	qtui := linkLabel(fmt.Sprintf("%s | <a href=\"https://github.com/VoxelPrismatic/Rich-Presence-U\">VoxelPrismatic</a>", a.tr.T("ABOUT_QT")))
	box.AddWidget(core.QWidget)
	box.AddWidget(qtui.QWidget)
	box.AddStretch()

	a.fillAboutLinks()
	return page
}

func (a *App) addSettingRow(parent *qt6.QVBoxLayout, title string, help *qt6.QToolButton, right *qt6.QWidget) {
	row := qt6.NewQHBoxLayout2()
	row.AddWidget(qt6.NewQLabel3(title).QWidget)
	row.AddStretch()
	if help != nil {
		row.AddWidget(help.QWidget)
	}
	row.AddWidget(right)
	parent.AddLayout(row.QLayout)
}

func (a *App) fillAboutLinks() {
	meta := a.nso.Meta
	ver := fmt.Sprintf("%s | %s", a.tr.T("ABOUT_VERSION"), Version)
	if meta.BinURL != "" {
		ver = fmt.Sprintf("%s | <a href=\"%s\">%s</a>", a.tr.T("ABOUT_VERSION"), meta.BinURL, Version)
	}
	a.verLabel.SetText(ver)
	a.verLabel.SetOpenExternalLinks(true)
	home := meta.Home
	if home == "" {
		home = "https://ninstars.blogspot.com/rpc"
	}
	help := meta.HelpURL(a.tr.Lang())
	a.infoLabel.SetText(fmt.Sprintf("%s | <a href=\"%s\">Rich Presence U</a> | <a href=\"%s\">%s</a>",
		a.tr.T("ABOUT_INFO"), home, help, a.tr.T("ABOUT_HELP")))
	a.infoLabel.SetOpenExternalLinks(true)
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
