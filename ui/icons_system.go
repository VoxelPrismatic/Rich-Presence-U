//go:build !portable

package ui

func useSystemIcons() bool { return true }

func iconFile(kind, name string) string {
	return "/usr/share/icons/breeze/" + kind + "/16/" + name + ".svg"
}
