package ui

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/discord"
)

func (a *App) updateApply() {
	if a.applyBtn == nil {
		return
	}
	switch {
	case a.busy:
		a.applyBtn.SetEnabled(false)
		a.applyBtn.SetIcon(iconNamed("network-connect", "network-connect"))
		a.applyBtn.SetText(a.tr.T("STATUS_CONNECTING"))
		a.applyBtn.SetToolTip(a.tr.T("STATUS_CONNECTING"))
	case !a.rpc.Connected():
		a.applyBtn.SetEnabled(true)
		a.applyBtn.SetIcon(iconNamed("network-disconnect", "network-disconnect"))
		a.applyBtn.SetText(a.tr.T("STATUS_CONNECT"))
		a.applyBtn.SetToolTip(a.tr.T("STATUS_CONNECT"))
	case a.fingerprint() == a.applied:
		a.applyBtn.SetEnabled(false)
		a.applyBtn.SetIcon(iconNamed("checkmark", "checkmark"))
		a.applyBtn.SetText(a.tr.T("STATUS_APPLIED"))
		a.applyBtn.SetToolTip(a.tr.T("STATUS_APPLIED"))
	default:
		a.applyBtn.SetEnabled(true)
		a.applyBtn.SetIcon(iconNamed("document-save", "document-save"))
		a.applyBtn.SetText(a.tr.T("STATUS_APPLY"))
		a.applyBtn.SetToolTip(a.tr.T("STATUS_APPLY"))
	}
	a.visBtn.SetChecked(a.settings.Activity)
	if a.settings.Activity {
		a.visBtn.SetIcon(iconNamed("view-visible", "view-visible"))
		a.visBtn.SetToolTip(a.tr.T("STATUS_ENABLED"))
	} else {
		a.visBtn.SetIcon(iconNamed("view-visible-off", "view-visible-off"))
		a.visBtn.SetToolTip(a.tr.T("STATUS_DISABLED"))
	}
	if a.rpc.Connected() {
		u := a.rpc.User()
		a.userName.SetText(u.DisplayName())
		a.userStatus.SetText(a.tr.T("USER_CONNECTED"))
	} else if a.busy {
		a.userStatus.SetText(a.tr.T("USER_CONNECTING"))
	} else {
		a.userStatus.SetText(a.tr.T("USER_DISCONNECTED"))
	}
}

func (a *App) onDisconnected() {
	a.busy = false
	a.applied = ""
	a.built = nil
	if a.avatar != nil {
		a.avatar.SetPixmap(qt6.NewQPixmap())
	}
	if a.userName != nil {
		a.userName.SetText(a.tr.T("USER_DISCORD"))
	}
	a.updateApply()
	a.refreshScreensaver()
}

func (a *App) onApply() {
	if a.busy {
		return
	}
	if !a.rpc.Connected() {
		a.connect(false)
		return
	}
	if !a.settings.Activity && !a.warnHide {
		a.warnHide = true
		qt6.QMessageBox_Information(a.win.QWidget, a.tr.T("INVISIBLE_STATUS_TITLE"), popupText(a.tr.T("INVISIBLE_STATUS_HINT")))
	}
	a.pushStatus()
}

func (a *App) connect(andPush bool) {
	a.busy = true
	a.updateApply()
	id := a.nso.Meta.ClientID(a.discordSystem())
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		user, err := a.rpc.Connect(ctx, id)
		mainthread.Start(func() {
			a.busy = false
			if err != nil {
				a.debug("discord connect: %v", err)
				a.userStatus.SetText(a.tr.T("USER_DISCONNECTED"))
				qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("CONNECTION_ERROR_TITLE"), popupText(a.tr.T("CONNECTION_ERROR_HINT")))
				a.onDisconnected()
				return
			}
			a.debug("discord connected as %s", user.DisplayName())
			a.loadAvatar(user)
			if andPush {
				a.pushStatus()
			} else {
				a.updateApply()
			}
			a.refreshScreensaver()
		})
	}()
}

func (a *App) pushStatus() {
	a.rememberGame()
	a.commitClockForApply()
	act := discord.Build(a.presenceForPush())
	fp := a.fingerprint()
	sysKey := a.sysKey()
	gameID := a.sys().Game
	a.built = &act
	id := a.nso.Meta.ClientID(a.discordSystem())
	needSwitch := a.rpc.Connected() && a.rpc.ClientID() != id
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if needSwitch {
			if _, err := a.rpc.Connect(ctx, id); err != nil {
				mainthread.Start(func() {
					qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("CONNECTION_ERROR_TITLE"), popupText(a.tr.T("CONNECTION_ERROR_HINT")))
					a.onDisconnected()
				})
				return
			}
		}
		var err error
		if a.settings.Activity {
			err = a.rpc.SetActivity(ctx, &act)
		} else {
			err = a.rpc.Clear(ctx)
		}
		mainthread.Start(func() {
			if err != nil {
				a.debug("set activity: %v", err)
				qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("CONNECTION_ERROR_TITLE"), err.Error())
			} else {
				a.applied = fp
				a.appliedSys = sysKey
				a.appliedGame = gameID
				a.ensureElapsedTick()
				a.showClockForCurrent()
				a.updateElapsed()
			}
			a.updateApply()
			a.refreshScreensaver()
			a.persist()
		})
	}()
}

func (a *App) onVisibility() {
	a.settings.Activity = a.visBtn.IsChecked()
	if !a.rpc.Connected() || a.built == nil {
		a.updateApply()
		a.refreshScreensaver()
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		if a.settings.Activity {
			err = a.rpc.SetActivity(ctx, a.built)
		} else {
			err = a.rpc.Clear(ctx)
		}
		mainthread.Start(func() {
			if err != nil {
				a.debug("toggle activity: %v", err)
			}
			a.updateApply()
			a.refreshScreensaver()
		})
	}()
}

func (a *App) onDataAction() {
	kind := a.dataCombo.CurrentData().ToString()
	switch kind {
	case "cache":
		if qt6.QMessageBox_Question6(a.win.QWidget, a.tr.T("RESET_CACHE_TITLE"), popupText(a.tr.T("RESET_CACHE_HINT")), qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__No) != qt6.QMessageBox__Yes {
			return
		}
		_ = a.nso.ClearCache()
		a.refillGameCombo()
		a.refreshGameUI()
		a.updateApply()
	case "all":
		if qt6.QMessageBox_Question6(a.win.QWidget, a.tr.T("RESET_ALL_TITLE"), popupText(a.tr.T("RESET_ALL_HINT")), qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__No) != qt6.QMessageBox__Yes {
			return
		}
		_ = a.nso.ResetAll()
		qt6.QCoreApplication_Quit()
	}
}

func (a *App) loadAvatar(user discord.User) {
	url := user.AvatarURL()
	if url == "" {
		return
	}
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		mainthread.Start(func() {
			pix := qt6.NewQPixmap()
			if pix.LoadFromDataWithData(b) {
				a.avatar.SetPixmap(maskPixmap(pix, 32, 0))
			}
		})
	}()
}
