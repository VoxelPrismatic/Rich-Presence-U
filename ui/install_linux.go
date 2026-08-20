//go:build linux

package ui

import (
	"context"
	"os"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/svc"
)

func (a *App) maybeInstall() {
	qt6.QGuiApplication_SetDesktopFileName("rich-presence-u")
	if a == nil || a.nso == nil {
		return
	}
	configDir := a.nso.ConfigDir
	if svc.Installed(configDir) {
		return
	}
	if a.settings.InstallDeclined == svc.VERSION {
		return
	}
	parent := a.dialogParent()
	ans := qt6.QMessageBox_Question6(parent, a.tr.T("INSTALL_TITLE"), popupText(a.tr.T("INSTALL_HINT")), qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__Yes)
	if ans != qt6.QMessageBox__Yes {
		a.settings.InstallDeclined = svc.VERSION
		a.persist()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := svc.Install(ctx, configDir); err != nil {
		a.debug("install: %v", err)
		qt6.QMessageBox_Warning(parent, a.tr.T("INSTALL_TITLE"), err.Error())
		return
	}
	if b, err := os.ReadFile(svc.LogoPath(configDir)); err == nil {
		applyAppIcon(b)
		if a.win != nil {
			a.win.SetWindowIcon(qt6.QGuiApplication_WindowIcon())
		}
	}
	a.settings.InstallDeclined = ""
	a.persist()
}
