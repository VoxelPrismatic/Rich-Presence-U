package ui

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/svc"
)

// updateWanted reports whether latest should be offered. Test mode fetches
// the latest release even when it is not newer than the running build.
func updateWanted(latest, current string, testUpdate bool) bool {
	if latest == "" {
		return false
	}
	if testUpdate {
		return true
	}
	return svc.Newer(latest, current)
}

// downloadBar maps a byte count onto a progress-bar range.
// known is false when the server did not send a length.
func downloadBar(written, total int64) (value, maximum int, known bool) {
	if total <= 0 {
		return 0, 0, false
	}
	if written < 0 {
		written = 0
	}
	if written > total {
		written = total
	}
	const maximumBar = 1000
	return int(written * maximumBar / total), maximumBar, true
}

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
	if !updateWanted(latest, svc.VERSION, a.settings.TestUpdate) {
		return
	}
	if !a.settings.TestUpdate && a.settings.UpdateDeclined == latest {
		return
	}
	parent := a.dialogParent()
	if runtime.GOOS != "linux" {
		qt6.QMessageBox_Information(parent, a.tr.T("UPDATE_TITLE"), popupText(fmt.Sprintf(a.tr.T("UPDATE_AVAILABLE"), latest)))
		a.settings.UpdateDeclined = latest
		a.persist()
		return
	}
	ans := qt6.QMessageBox_Question6(
		parent,
		a.tr.T("UPDATE_TITLE"),
		popupText(fmt.Sprintf(a.tr.T("UPDATE_AUTO_HINT"), latest)),
		qt6.QMessageBox__Yes|qt6.QMessageBox__No, qt6.QMessageBox__Yes,
	)
	if ans != qt6.QMessageBox__Yes {
		a.settings.UpdateDeclined = latest
		a.persist()
		return
	}
	bin, err := a.downloadUpdate(parent, latest)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
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

// downloadUpdate shows a progress dialog while the Linux release is fetched.
func (a *App) downloadUpdate(parent *qt6.QWidget, latest string) (string, error) {
	label := fmt.Sprintf(a.tr.T("UPDATE_DOWNLOADING"), latest)
	cancelText := a.tr.T("UPDATE_CANCEL_DOWNLOAD")
	var dlg *qt6.QProgressDialog
	if parent == nil {
		dlg = qt6.NewQProgressDialog3(label, cancelText, 0, 1000)
		dlg.GoGC()
	} else {
		dlg = qt6.NewQProgressDialog5(label, cancelText, 0, 1000, parent)
	}
	dlg.SetWindowTitle(a.tr.T("UPDATE_TITLE"))
	dlg.SetWindowModality(qt6.ApplicationModal)
	dlg.SetMinimumDuration(0)
	dlg.SetAutoClose(false)
	dlg.SetAutoReset(false)
	dlg.SetValue(0)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	var shown atomic.Bool
	shown.Store(true)
	dlg.OnCanceled(func() { cancel() })

	type result struct {
		bin string
		err error
	}
	done := make(chan result, 1)
	go func() {
		bin, err := svc.ApplyUpdateTo(ctx, a.nso.ConfigDir, svc.ApplicationsDir(), latest, func(written, total int64) {
			mainthread.Start(func() {
				if !shown.Load() {
					return
				}
				paintDownload(dlg, written, total)
			})
		})
		done <- result{bin, err}
		mainthread.Start(func() {
			if shown.Load() {
				dlg.Accept()
			}
		})
	}()
	dlg.Exec()
	canceled := dlg.WasCanceled()
	shown.Store(false)
	cancel()
	r := <-done
	if canceled || errors.Is(r.err, context.Canceled) {
		return "", context.Canceled
	}
	return r.bin, r.err
}

func paintDownload(dlg *qt6.QProgressDialog, written, total int64) {
	value, maximum, known := downloadBar(written, total)
	if !known {
		dlg.SetRange(0, 0)
		return
	}
	dlg.SetRange(0, maximum)
	if maximum > 0 && value >= maximum {
		value = maximum - 1
	}
	dlg.SetValue(value)
}
