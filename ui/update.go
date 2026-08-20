package ui

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/svc"
)

func (a *App) maybeUpdate() {
	if a == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	latest, err := svc.LatestVersion(ctx)
	if err != nil {
		a.debug("update check: %v", err)
		return
	}
	if !svc.Newer(latest, svc.VERSION) {
		return
	}
	if a.settings.UpdateDeclined == latest {
		return
	}
	parent := a.dialogParent()
	if runtime.GOOS != "linux" {
		qt6.QMessageBox_Information(parent, a.tr.T("UPDATE_TITLE"), popupText(fmt.Sprintf(a.tr.T("UPDATE_AVAILABLE"), latest)))
		a.settings.UpdateDeclined = latest
		a.persist()
		return
	}
	ans := qt6.QMessageBox_Question6(parent, a.tr.T("UPDATE_TITLE"), popupText(fmt.Sprintf(a.tr.T("UPDATE_HINT"), latest)), qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__Yes)
	if ans != qt6.QMessageBox__Yes {
		a.settings.UpdateDeclined = latest
		a.persist()
		return
	}
	dl, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	bin, err := svc.ApplyUpdate(dl, a.nso.ConfigDir, latest)
	if err != nil {
		a.debug("update: %v", err)
		qt6.QMessageBox_Warning(parent, a.tr.T("UPDATE_TITLE"), err.Error())
		return
	}
	a.settings.UpdateDeclined = ""
	a.persist()
	if err := svc.Restart(bin); err != nil {
		qt6.QMessageBox_Warning(parent, a.tr.T("UPDATE_TITLE"), err.Error())
	}
}
