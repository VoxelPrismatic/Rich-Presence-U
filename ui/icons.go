package ui

import "github.com/mappu/miqt/qt6"

func breeze(kind, name string) string {
	return "/usr/share/icons/breeze/" + kind + "/16/" + name + ".svg"
}

func themeIcon(names ...string) *qt6.QIcon {
	for _, name := range names {
		ic := qt6.QIcon_FromTheme(name)
		if ic != nil && !ic.IsNull() {
			return ic
		}
	}
	return qt6.NewQIcon()
}

func fileIcon(path string) *qt6.QIcon {
	ic := qt6.NewQIcon4(path)
	if ic != nil && !ic.IsNull() {
		return ic
	}
	return qt6.NewQIcon()
}

func iconPublic() *qt6.QIcon {
	ic := themeIcon("folder-public")
	if ic.IsNull() {
		return fileIcon(breeze("places", "folder-public"))
	}
	return ic
}

func iconGames() *qt6.QIcon {
	ic := themeIcon("folder-games")
	if ic.IsNull() {
		return fileIcon(breeze("places", "folder-games"))
	}
	return ic
}

func iconHelp() *qt6.QIcon {
	ic := themeIcon("help-contextual", "help-hint")
	if ic.IsNull() {
		return fileIcon(breeze("actions", "help-contextual"))
	}
	return ic
}

func iconNamed(themeName, breezeName string) *qt6.QIcon {
	ic := themeIcon(themeName)
	if ic.IsNull() {
		return fileIcon(breeze("actions", breezeName))
	}
	return ic
}

func iconGoNext() *qt6.QIcon {
	ic := themeIcon("go-next", "arrow-right")
	if ic.IsNull() {
		return fileIcon(breeze("actions", "go-next"))
	}
	return ic
}

func iconShop() *qt6.QIcon {
	ic := themeIcon("amarok_cart_view", "amarok-cart-view")
	if ic.IsNull() {
		return fileIcon(breeze("actions", "amarok_cart_view"))
	}
	return ic
}

func iconIGDB() *qt6.QIcon {
	ic := themeIcon("compass")
	if ic.IsNull() {
		return fileIcon(breeze("actions", "compass"))
	}
	return ic
}

func iconFind() *qt6.QIcon {
	ic := themeIcon("edit-find")
	if ic.IsNull() {
		return fileIcon(breeze("actions", "edit-find"))
	}
	return ic
}

func helpButton(tip string) *qt6.QToolButton {
	b := qt6.NewQToolButton2()
	b.SetIcon(iconHelp())
	b.SetAutoRaise(true)
	b.SetToolTip(tip)
	b.SetFocusPolicy(qt6.NoFocus)
	b.SetCursor(qt6.NewQCursor2(qt6.WhatsThisCursor))
	return b
}
