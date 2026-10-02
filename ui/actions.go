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
	case a.statusApplied():
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
	a.refreshVisibility()
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

func visibilityIcon(pauseMode, counting bool) (name, tip string) {
	if pauseMode {
		if counting {
			return "media-playback-start", "STATUS_COUNTING"
		}
		return "media-playback-pause", "STATUS_PAUSED"
	}
	if counting {
		return "view-visible", "STATUS_ENABLED"
	}
	return "view-visible-off", "STATUS_DISABLED"
}

func (a *App) refreshVisibility() {
	if a.visBtn == nil {
		return
	}
	a.visBtn.SetChecked(a.settings.Activity)
	name, tip := visibilityIcon(a.settings.pausesTimer(), a.settings.Activity)
	a.visBtn.SetIcon(iconNamed(name, name))
	a.visBtn.SetToolTip(a.tr.T(tip))
}

func (a *App) onDisconnected() {
	a.busy = false
	a.applied = ""
	a.appliedBody = ""
	a.built = nil
	a.clockSync = false
	a.clockDirty = false
	a.clockFailed = false
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
	a.unhideForApply()
	if !a.settings.Activity && !a.settings.pausesTimer() && !a.warnHide {
		a.warnHide = true
		qt6.QMessageBox_Information(a.win.QWidget, a.tr.T("INVISIBLE_STATUS_TITLE"), popupText(a.tr.T("INVISIBLE_STATUS_HINT")))
	}
	a.pushStatus()
}

// unhideForApply turns the status on before a connected apply when that option is set.
func (a *App) unhideForApply() {
	if !a.settings.AutoUnhide || a.settings.Activity {
		return
	}
	a.settings.Activity = true
	a.clk.heldOK = false
	a.clockDirty = false
	a.refreshVisibility()
	a.refreshScreensaver()
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
	// The payload and the fingerprint have to include the clock that this
	// apply will start. presenceTimestamps stays empty until running is set,
	// so without this the first click reaches Discord and then the button
	// still says Apply, because starting the clock changes the fingerprint.
	wasRunning := a.clk.running
	a.clk.running = true
	act := discord.Build(a.presenceForPush())
	fp := a.fingerprint()
	body := a.fingerprintBody()
	if !wasRunning {
		a.clk.running = false
	}
	sysKey := a.sysKey()
	gameID := a.sys().Game
	a.built = &act
	send := a.settings.Activity || a.settings.pausesTimer()
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
		if send {
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
				a.appliedBody = body
				a.clockDirty = false
				a.clockFailed = false
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
	if a.settings.pausesTimer() {
		if a.settings.Activity {
			a.clk.heldOK = false
			a.clockDirty = true
			a.clockFailed = false
		}
		a.updateApply()
		a.refreshScreensaver()
		return
	}
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

// syncAppliedClock pushes the play clock when the visible status is the one
// already applied. Unchanged timestamps are left alone.
func (a *App) syncAppliedClock() {
	if a.clockSync || a.built == nil || !a.rpc.Connected() {
		return
	}
	start, end := a.presenceTimestamps()
	same := a.built.StartTimestamp == start && a.built.EndTimestamp == end
	if a.clockFailed && a.clockFailStart == start && a.clockFailEnd == end {
		return
	}
	if same && !a.clockDirty {
		return
	}
	if !a.clockStatusApplied() {
		a.clockDirty = false
		a.clockFailed = true
		a.clockFailStart = start
		a.clockFailEnd = end
		return
	}
	act := *a.built
	act.StartTimestamp = start
	act.EndTimestamp = end
	a.clockSync = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := a.rpc.SetActivity(ctx, &act)
		mainthread.Start(func() {
			a.clockSync = false
			if err != nil {
				a.clockFailed = true
				a.clockFailStart = start
				a.clockFailEnd = end
				a.debug("sync clock: %v", err)
				return
			}
			if a.built == nil || !a.rpc.Connected() || !a.settings.pausesTimer() || !a.settings.Activity {
				return
			}
			nowStart, nowEnd := a.presenceTimestamps()
			if nowStart != start || nowEnd != end || !a.clockStatusApplied() {
				return
			}
			next := *a.built
			next.StartTimestamp = start
			next.EndTimestamp = end
			a.built = &next
			a.applied = a.fingerprint()
			a.clockDirty = false
			a.clockFailed = false
			a.updateApply()
		})
	}()
}

func (a *App) setHideBehavior(mode string) {
	if a.silent {
		return
	}
	if mode != hidePause {
		mode = hideDiscord
	}
	if a.settings.HideBehavior == mode {
		return
	}
	a.settings.HideBehavior = mode
	if mode == hidePause {
		a.clockDirty = true
		a.clockFailed = false
	} else {
		a.clk.heldOK = false
	}
	a.updateApply()
	a.refreshScreensaver()
	if mode == hidePause || a.settings.Activity || !a.rpc.Connected() || a.built == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := a.rpc.Clear(ctx)
		mainthread.Start(func() {
			if err != nil {
				a.debug("hide activity: %v", err)
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
