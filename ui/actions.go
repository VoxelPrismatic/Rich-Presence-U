package ui

import (
	"context"
	"fmt"
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
	if a.timerBtn != nil {
		a.timerBtn.SetChecked(a.timerRemaining() > 0)
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
	a.bumpElapsed(false)
	act := discord.Build(a.presenceForPush())
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
				a.applied = a.fingerprint()
				a.startHideTimer()
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

func (a *App) timerRemaining() int {
	if a.hide == nil || !a.hide.IsActive() {
		return 0
	}
	ms := a.hide.RemainingTime()
	if ms <= 0 {
		return 0
	}
	return (ms + 999) / 1000
}

func formatClock(sec int) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func paddedSpin(max, digits, width int) *qt6.QSpinBox {
	sp := qt6.NewQSpinBox2()
	sp.SetRange(0, max)
	sp.SetAlignment(qt6.AlignRight | qt6.AlignVCenter)
	sp.SetSizePolicy2(qt6.QSizePolicy__Fixed, qt6.QSizePolicy__Fixed)
	sp.SetFixedWidth(width)
	pad := func() {
		if le := sp.LineEdit(); le != nil {
			txt := fmt.Sprintf("%0*d", digits, sp.Value())
			if le.Text() != txt {
				cur := le.CursorPosition()
				le.SetText(txt)
				if cur > len(txt) {
					cur = len(txt)
				}
				le.SetCursorPosition(cur)
			}
		}
	}
	sp.OnValueChanged(func(int) { pad() })
	if sp.Value() == 0 {
		sp.SetValue(1)
		sp.SetValue(0)
	}
	return sp
}

func (a *App) startHideTimer() {
	if a.hide == nil {
		return
	}
	a.hide.Stop()
	if a.timerEnabled && a.settings.Timer > 0 && a.settings.Activity {
		a.hide.Start(a.settings.Timer * 1000)
		if a.rpc.Connected() && a.built != nil {
			act := discord.Build(a.presenceForPush())
			a.built = &act
			go func() {
				_ = a.rpc.SetActivity(context.Background(), &act)
			}()
		}
	}
	a.updateApply()
	a.updateElapsed()
}

func (a *App) onTimer() {
	d := qt6.NewQDialog(a.win.QWidget)
	d.SetWindowTitle(a.tr.T("TIMER_TITLE"))
	lay := qt6.NewQVBoxLayout(d.QWidget)
	lay.SetSizeConstraint(qt6.QLayout__SetFixedSize)
	hint := qt6.NewQLabel3(popupText(a.tr.T("TIMER_HINT")))
	hint.SetWordWrap(true)
	lay.AddWidget(hint.QWidget)
	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(2)
	row.SetContentsMargins(0, 0, 0, 0)
	h := paddedSpin(99, 2, 64)
	h.SetToolTip(a.tr.T("TIMER_HOURS"))
	m := paddedSpin(59, 2, 56)
	m.SetToolTip(a.tr.T("TIMER_MINUTES"))
	s := paddedSpin(59, 2, 56)
	s.SetToolTip(a.tr.T("TIMER_SECONDS"))
	sec := a.settings.Timer
	h.SetValue(sec / 3600)
	m.SetValue((sec % 3600) / 60)
	s.SetValue(sec % 60)
	row.AddWidget(h.QWidget)
	row.AddWidget(qt6.NewQLabel3(":").QWidget)
	row.AddWidget(m.QWidget)
	row.AddWidget(qt6.NewQLabel3(":").QWidget)
	row.AddWidget(s.QWidget)
	lay.AddLayout(row.QLayout)
	btns := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel)
	btns.OnAccepted(func() { d.Accept() })
	btns.OnRejected(func() { d.Reject() })
	remove := btns.AddButton2(a.tr.T("TIMER_REMOVE"), qt6.QDialogButtonBox__DestructiveRole)
	remove.OnClicked(func() { d.Done(2) })
	lay.AddWidget(btns.QWidget)
	switch d.Exec() {
	case int(qt6.QDialog__Accepted):
		a.settings.Timer = h.Value()*3600 + m.Value()*60 + s.Value()
		a.timerEnabled = a.settings.Timer > 0
		a.persist()
		a.startHideTimer()
	case 2:
		a.clearTimer()
	}
	if a.timerBtn != nil {
		a.timerBtn.SetChecked(a.timerRemaining() > 0)
	}
}

func (a *App) stopTimer() {
	a.timerEnabled = false
	if a.hide != nil {
		a.hide.Stop()
	}
	if a.rpc.Connected() && a.settings.Activity && a.built != nil {
		act := discord.Build(a.presenceForPush())
		a.built = &act
		go func() {
			_ = a.rpc.SetActivity(context.Background(), &act)
		}()
	}
	a.updateApply()
	a.updateElapsed()
}

func (a *App) clearTimer() {
	a.settings.Timer = 0
	a.persist()
	a.stopTimer()
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

func (a *App) updateElapsed() {
	if a.elapsed == nil {
		return
	}
	if rem := a.timerRemaining(); rem > 0 {
		a.elapsed.SetText(formatClock(rem))
		if a.elapsedIcon != nil {
			a.elapsedIcon.SetPixmap(iconNamed("chronometer", "chronometer").Pixmap2(16, 16))
		}
		if a.timerBtn != nil {
			a.timerBtn.SetChecked(true)
		}
		return
	}
	if a.elapsedIcon != nil {
		a.elapsedIcon.SetPixmap(iconGames().Pixmap2(16, 16))
	}
	if a.timerBtn != nil {
		a.timerBtn.SetChecked(false)
	}
	if a.start == 0 {
		a.elapsed.SetText("0:00")
		return
	}
	d := time.Since(time.Unix(a.start, 0))
	if d < 0 {
		d = 0
	}
	a.elapsed.SetText(formatClock(int(d.Seconds())))
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
