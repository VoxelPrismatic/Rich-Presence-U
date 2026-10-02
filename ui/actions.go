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

// visibilityPressed is the button's checked state. Pause and hidden are
// pressed, so the usual counting or visible state keeps a normal background.
func visibilityPressed(activity bool) bool {
	return !activity
}

func (a *App) refreshVisibility() {
	if a.visBtn == nil {
		return
	}
	a.visBtn.SetChecked(visibilityPressed(a.settings.Activity))
	name, tip := visibilityIcon(!a.settings.HideDiscord, a.settings.Activity)
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
	a.clockSent = time.Time{}
	a.clockSentStart = 0
	a.clockSentEnd = 0
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
	if !a.settings.Activity && a.settings.HideDiscord && !a.warnHide {
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

const connectPollEvery = 5 * time.Second

func (a *App) initConnectPoll() {
	if a.win == nil {
		return
	}
	a.pollTimer = qt6.NewQTimer2(a.win.QObject)
	a.pollTimer.SetInterval(int(connectPollEvery / time.Millisecond))
	a.pollTimer.OnTimeout(func() { a.pollDiscord() })
}

func (a *App) syncConnectPoll() {
	if a.pollTimer == nil {
		return
	}
	if a.settings.AutoConnect != connectPoll {
		a.pollTimer.Stop()
		return
	}
	if !a.pollTimer.IsActive() {
		a.pollTimer.Start2()
	}
	a.pollDiscord()
}

func (a *App) pollDiscord() {
	if a.settings.AutoConnect != connectPoll || a.busy || a.polling || a.rpc.Connected() {
		return
	}
	a.connectQuiet()
}

func (a *App) connectQuiet() {
	a.connectGen++
	gen := a.connectGen
	a.polling = true
	id := a.nso.Meta.ClientID(a.discordSystem())
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		user, err := a.rpc.Connect(ctx, id)
		mainthread.Start(func() {
			if gen != a.connectGen {
				return
			}
			a.polling = false
			if err != nil || a.busy {
				return
			}
			a.debug("discord connected as %s", user.DisplayName())
			a.loadAvatar(user)
			a.updateApply()
			a.refreshScreensaver()
		})
	}()
}

func (a *App) connect(andPush bool) {
	a.connectGen++
	gen := a.connectGen
	a.polling = false
	a.busy = true
	a.updateApply()
	id := a.nso.Meta.ClientID(a.discordSystem())
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		user, err := a.rpc.Connect(ctx, id)
		mainthread.Start(func() {
			if gen != a.connectGen {
				return
			}
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
	send := a.settings.Activity || !a.settings.HideDiscord
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
				a.clockSent = time.Time{}
				a.clockSentStart = 0
				a.clockSentEnd = 0
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
	a.settings.Activity = !a.visBtn.IsChecked()
	if a.settings.Activity {
		a.clk.heldOK = false
		if a.settings.PauseTimer {
			a.clockDirty = true
			a.clockFailed = false
		}
	}
	a.updateApply()
	a.refreshScreensaver()
	if !a.settings.HideDiscord || !a.rpc.Connected() || a.built == nil {
		return
	}
	show := a.settings.Activity
	act := *a.built
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		if show {
			err = a.rpc.SetActivity(ctx, &act)
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

// syncAppliedClock pushes the play clock when that status is still on Discord.
// A running clock keeps one start time, and Discord counts from it. A frozen
// clock rebases that start as wall time moves, so the new pair is written
// about every five seconds. Hide from Discord clears the activity instead.
func (a *App) syncAppliedClock() {
	if a.clockSync || a.built == nil || !a.rpc.Connected() {
		return
	}
	if !a.settings.Activity && a.settings.HideDiscord {
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
	now := time.Now()
	if !a.clockDirty && !clockPushDue(a.clockSent, now) && !clockDisplayChanged(a.clk.down, a.clockSent, a.clockSentStart, a.clockSentEnd, start, end, now) {
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
	a.clockSent = now
	a.clockSentStart = start
	a.clockSentEnd = end
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
			if a.built == nil || !a.rpc.Connected() || !a.settings.PauseTimer {
				return
			}
			if !a.settings.Activity && a.settings.HideDiscord {
				return
			}
			if !a.clockStatusApplied() {
				return
			}
			// The frozen clock rebases again while this write is in flight.
			// Drop the dirty flag anyway so that rebase waits for the next
			// five-second push instead of sending on every tick.
			a.clockDirty = false
			a.clockFailed = false
			nowStart, nowEnd := a.presenceTimestamps()
			if nowStart != start || nowEnd != end {
				return
			}
			next := *a.built
			next.StartTimestamp = start
			next.EndTimestamp = end
			a.built = &next
			a.applied = a.fingerprint()
			a.updateApply()
		})
	}()
}

func (a *App) setPauseTimer(on bool) {
	if a.silent {
		return
	}
	a.settings.PauseTimer = on
	if on {
		a.clockDirty = true
		a.clockFailed = false
	} else {
		a.clk.heldOK = false
	}
	a.updateApply()
	a.refreshScreensaver()
}

func (a *App) setHideDiscord(on bool) {
	if a.silent {
		return
	}
	a.settings.HideDiscord = on
	a.updateApply()
	a.refreshScreensaver()
	if a.settings.Activity || !a.rpc.Connected() || a.built == nil {
		return
	}
	act := *a.built
	if !on && a.settings.PauseTimer {
		act.StartTimestamp, act.EndTimestamp = a.presenceTimestamps()
		a.clockSent = time.Now()
		a.clockSentStart = act.StartTimestamp
		a.clockSentEnd = act.EndTimestamp
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		if on {
			err = a.rpc.Clear(ctx)
		} else {
			err = a.rpc.SetActivity(ctx, &act)
		}
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
