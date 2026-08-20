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
	"github.com/voxelprismatic/richpresenceu/nso"
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
		a.userName.SetText("Discord")
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
	id := a.nso.Meta.ClientID(a.settings.System)
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
	act := discord.Build(a.presence())
	a.built = &act
	id := a.nso.Meta.ClientID(a.settings.System)
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

func (a *App) startHideTimer() {
	if a.hide == nil {
		return
	}
	a.hide.Stop()
	if a.settings.Timer > 0 && a.settings.Activity {
		a.hide.Start(a.settings.Timer * 1000)
	}
}

func (a *App) onTimer() {
	d := qt6.NewQDialog(a.win.QWidget)
	d.SetWindowTitle(a.tr.T("TIMER_TITLE"))
	lay := qt6.NewQVBoxLayout(d.QWidget)
	lay.AddWidget(qt6.NewQLabel3(popupText(a.tr.T("TIMER_HINT"))).QWidget)
	row := qt6.NewQHBoxLayout2()
	h := qt6.NewQSpinBox2()
	h.SetRange(0, 99)
	h.SetToolTip(a.tr.T("TIMER_HOURS"))
	m := qt6.NewQSpinBox2()
	m.SetRange(0, 59)
	m.SetToolTip(a.tr.T("TIMER_MINUTES"))
	s := qt6.NewQSpinBox2()
	s.SetRange(0, 59)
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
	lay.AddWidget(btns.QWidget)
	if d.Exec() == int(qt6.QDialog__Accepted) {
		a.settings.Timer = h.Value()*3600 + m.Value()*60 + s.Value()
		a.startHideTimer()
	}
}

func (a *App) onDataAction() {
	kind := a.dataCombo.CurrentData().ToString()
	switch kind {
	case "cache":
		if qt6.QMessageBox_Question6(a.win.QWidget, a.tr.T("RESET_CACHE_TITLE"), popupText(a.tr.T("RESET_CACHE_HINT")), qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__No) != qt6.QMessageBox__Yes {
			return
		}
		_ = a.nso.ClearCache()
		a.refreshTitles(true)
	case "all":
		if qt6.QMessageBox_Question6(a.win.QWidget, a.tr.T("RESET_ALL_TITLE"), popupText(a.tr.T("RESET_ALL_HINT")), qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__No) != qt6.QMessageBox__Yes {
			return
		}
		_ = a.nso.ResetAll()
		qt6.QCoreApplication_Quit()
	}
}

func (a *App) refreshTitles(force bool) {
	if !force && a.nso.CachePresent() && !a.nso.NeedsRefresh(a.settings.RefreshEvery(), time.Unix(a.settings.RefreshLast, 0)) {
		_ = a.nso.LoadCache()
		a.refillCompleter()
		a.refreshGameUI()
		a.fillAboutLinks()
		return
	}
	prog := qt6.NewQProgressDialog5(a.tr.T("REFRESHING_STEP_1"), "", 0, 1+len(nso.Systems), a.win.QWidget)
	prog.SetWindowTitle(a.tr.T("REFRESH_TITLE"))
	prog.SetMinimumDuration(0)
	prog.SetCancelButton(nil)
	prog.Show()
	go func() {
		err := a.nso.Refresh(context.Background(), func(p nso.Progress) {
			mainthread.Start(func() {
				prog.SetValue(p.Current)
				if p.Stage == "done" {
					prog.SetLabelText(a.tr.T("REFRESHING_DONE"))
				}
			})
		})
		mainthread.Start(func() {
			prog.Close()
			if err != nil {
				a.debug("refresh: %v", err)
				qt6.QMessageBox_Warning(a.win.QWidget, a.tr.T("REFRESH_TITLE"), err.Error())
				_ = a.nso.LoadCache()
			} else {
				a.settings.RefreshLast = time.Now().Unix()
			}
			a.refillCompleter()
			a.refreshGameUI()
			a.fillAboutLinks()
			a.updateApply()
		})
	}()
}

func (a *App) updateElapsed() {
	if a.elapsed == nil {
		return
	}
	if a.start == 0 {
		a.elapsed.SetText("0:00")
		return
	}
	d := time.Since(time.Unix(a.start, 0))
	if d < 0 {
		d = 0
	}
	sec := int(d.Seconds())
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		a.elapsed.SetText(fmt.Sprintf("%d:%02d:%02d", h, m, s))
	} else {
		a.elapsed.SetText(fmt.Sprintf("%d:%02d", m, s))
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
