package ui

import (
	"github.com/mappu/miqt/qt6"
)

func fileIcon(path string) *qt6.QIcon {
	ic := qt6.NewQIcon4(path)
	if ic != nil && !ic.IsNull() {
		return ic
	}
	return qt6.NewQIcon()
}

func pickIcon(kind, name string, _ ...string) *qt6.QIcon {
	return fileIcon(iconFile(kind, name))
}

func iconPublic() *qt6.QIcon {
	return pickIcon("places", "folder-public", "folder-public")
}

func iconGames() *qt6.QIcon {
	return pickIcon("places", "folder-games", "folder-games")
}

func iconHelp() *qt6.QIcon {
	return pickIcon("actions", "help-contextual", "help-contextual", "help-hint")
}

func iconNamed(themeName, breezeName string) *qt6.QIcon {
	return pickIcon("actions", breezeName, themeName)
}

var goNextIcon *qt6.QIcon

func iconGoNext() *qt6.QIcon {
	if goNextIcon == nil {
		goNextIcon = pickIcon("actions", "go-next", "go-next", "arrow-right")
	}
	return goNextIcon
}

func clearIconCache() {
	goNextIcon = nil
}

func iconShop() *qt6.QIcon {
	return pickIcon("actions", "amarok_cart_view", "amarok_cart_view", "amarok-cart-view")
}

func iconIGDB() *qt6.QIcon {
	return pickIcon("actions", "compass", "compass")
}

func iconFind() *qt6.QIcon {
	return pickIcon("actions", "edit-find", "edit-find")
}

func (a *App) helpButton(tip string) *qt6.QToolButton {
	b := qt6.NewQToolButton2()
	b.SetIcon(iconHelp())
	b.SetAutoRaise(true)
	b.SetToolTip(tip)
	b.SetFocusPolicy(qt6.NoFocus)
	b.SetCursor(qt6.NewQCursor2(qt6.WhatsThisCursor))
	a.helpBtns = append(a.helpBtns, b)
	return b
}
