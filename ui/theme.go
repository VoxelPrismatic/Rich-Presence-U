package ui

import "github.com/mappu/miqt/qt6"

const (
	schemeLight = "light"
	schemeDark  = "dark"
)

// iconScheme selects assets/light or assets/dark. Qt UI work stays on the main thread.
var iconScheme = schemeLight

// schemeForLightness reports dark when the window background is darker than the text.
func schemeForLightness(bg, fg int) string {
	if bg < fg {
		return schemeDark
	}
	return schemeLight
}

func paletteScheme(pal *qt6.QPalette) string {
	if pal == nil {
		return schemeLight
	}
	bg := pal.ColorWithCr(qt6.QPalette__Window)
	fg := pal.ColorWithCr(qt6.QPalette__WindowText)
	if bg == nil || fg == nil || !bg.IsValid() || !fg.IsValid() {
		return schemeLight
	}
	return schemeForLightness(bg.Lightness(), fg.Lightness())
}

// applyIconScheme stores the scheme for pal. It reports whether the scheme changed.
func applyIconScheme(pal *qt6.QPalette) bool {
	next := paletteScheme(pal)
	if next == iconScheme {
		return false
	}
	iconScheme = next
	clearIconCache()
	return true
}

func (a *App) watchIconScheme() {
	applyIconScheme(qt6.QApplication_PaletteWithClassName("QWidget"))
	if a.win == nil {
		return
	}
	a.win.OnChangeEvent(func(super func(ev *qt6.QEvent), ev *qt6.QEvent) {
		if ev != nil {
			switch ev.Type() {
			case qt6.QEvent__ApplicationPaletteChange, qt6.QEvent__PaletteChange, qt6.QEvent__ThemeChange:
				super(ev)
				if applyIconScheme(a.win.Palette()) {
					a.refreshThemeIcons()
				}
				return
			}
		}
		super(ev)
	})
}

func (a *App) refreshThemeIcons() {
	if a.settingsBack != nil {
		a.settingsBack.SetIcon(iconNamed("go-previous", "go-previous"))
	}
	if a.descIcon != nil {
		a.descIcon.SetPixmap(iconPublic().Pixmap2(16, 16))
	}
	if a.clk.icon != nil {
		a.clk.icon.SetPixmap(iconGames().Pixmap2(16, 16))
	}
	if a.clk.reset != nil {
		a.clk.reset.SetIcon(iconNamed("view-refresh", "view-refresh"))
	}
	a.refreshClockMode()
	if a.cfgBtn != nil {
		a.cfgBtn.SetIcon(iconNamed("configure", "configure"))
	}
	if a.platBack != nil {
		a.platBack.SetIcon(iconNamed("go-previous", "go-previous"))
	}
	if a.platFind != nil {
		a.platFind.SetIcon(iconFind())
	}
	for _, b := range a.helpBtns {
		if b != nil {
			b.SetIcon(iconHelp())
		}
	}
	if a.dataCombo != nil {
		a.dataCombo.SetItemIcon(0, iconNamed("action-unavailable", "action-unavailable-symbolic"))
		a.dataCombo.SetItemIcon(1, iconNamed("edit-clear-history", "edit-clear-history"))
		a.dataCombo.SetItemIcon(2, iconNamed("user-trash", "albumfolder-user-trash"))
		if a.dataBtn != nil {
			a.dataBtn.SetIcon(a.dataCombo.ItemIcon(a.dataCombo.CurrentIndex()))
		}
	}
	a.updateApply()
	a.updateInfoButton()
}
