package ui

import (
	"time"

	"github.com/mappu/miqt/qt6"
)

const clockMaxSeconds = 23*3600 + 59*60 + 59

// playClock is the QTimeEdit field, its count mode, and the single-shot
// Qt timer that refreshes it from the start and end timestamps.
type playClock struct {
	edit  *qt6.QTimeEdit
	mode  *qt6.QPushButton
	reset *qt6.QPushButton
	icon  *qt6.QLabel
	lock  bool

	down    bool
	origin  int
	start   time.Time
	end     time.Time
	running bool
	alerted bool
	drafts  map[string]int

	timer *qt6.QTimer
}

func (a *App) initClockTimer() {
	c := &a.clk
	c.timer = qt6.NewQTimer2(a.win.QObject)
	c.timer.SetInterval(500)
	c.timer.OnTimeout(func() { a.updateElapsed() })
}

func (a *App) buildClockRow() *qt6.QHBoxLayout {
	c := &a.clk
	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(2)
	c.icon = qt6.NewQLabel2()
	c.icon.SetPixmap(iconGames().Pixmap2(16, 16))
	c.edit = newClockEdit()
	c.edit.OnTimeChanged(func(qt6.QTime) { a.onClockEdited() })
	c.edit.OnStepBy(func(super func(steps int), steps int) { a.stepClock(steps) })
	c.edit.OnWheelEvent(func(super func(event *qt6.QWheelEvent), event *qt6.QWheelEvent) {
		c.edit.SetFocus()
		super(event)
	})
	row.AddWidget(c.icon.QWidget)
	row.AddWidget(c.edit.QWidget)

	c.mode = qt6.NewQPushButton2()
	c.mode.OnClicked(func() { a.toggleClockMode() })
	c.reset = qt6.NewQPushButton2()
	c.reset.SetIcon(iconNamed("view-refresh", "view-refresh"))
	c.reset.SetToolTip(a.tr.T("CLOCK_RESET"))
	c.reset.OnClicked(func() { a.resetClock() })
	row.AddWidget(c.mode.QWidget)
	row.AddWidget(c.reset.QWidget)
	a.refreshClockMode()
	a.showSeconds(0)
	return row
}

func newClockEdit() *qt6.QTimeEdit {
	edit := qt6.NewQTimeEdit2()
	edit.SetDisplayFormat("HH:mm:ss")
	edit.SetTimeRange(*qt6.NewQTime4(0, 0, 0), *qt6.NewQTime4(23, 59, 59))
	edit.SetWrapping(false)
	edit.SetAccelerated(true)
	edit.SetAlignment(qt6.AlignHCenter)
	edit.SetFocusPolicy(qt6.WheelFocus)
	return edit
}

// stepClockSeconds moves one field by steps and carries into the next field.
// unit is 1, 60, or 3600. The result stays inside a QTime.
func stepClockSeconds(current, unit, steps int) int {
	next := current + steps*unit
	if next < 0 {
		return 0
	}
	if next > clockMaxSeconds {
		return clockMaxSeconds
	}
	return next
}

func displayedSeconds(countDown bool, start, end, now time.Time) int {
	if countDown {
		if end.IsZero() {
			return 0
		}
		left := end.Sub(now)
		if left <= 0 {
			return 0
		}
		return int((left + time.Second - time.Nanosecond) / time.Second)
	}
	if start.IsZero() {
		return 0
	}
	elapsed := now.Sub(start)
	if elapsed < 0 {
		return 0
	}
	return int(elapsed / time.Second)
}

func countdownPercent(start, end, now time.Time) int {
	total := end.Sub(start)
	if total <= 0 {
		return 100
	}
	done := now.Sub(start)
	if done <= 0 {
		return 0
	}
	if done >= total {
		return 100
	}
	return int(done * 100 / total)
}

func (a *App) clockSeconds() int {
	c := &a.clk
	if c.edit == nil {
		return 0
	}
	t := c.edit.Time()
	if t == nil {
		return 0
	}
	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}

func (a *App) showSeconds(sec int) {
	c := &a.clk
	if c.edit == nil {
		return
	}
	if sec < 0 {
		sec = 0
	}
	if sec > clockMaxSeconds {
		sec = clockMaxSeconds
	}
	c.lock = true
	c.edit.SetTime(*qt6.NewQTime4(sec/3600, (sec%3600)/60, sec%60))
	c.lock = false
}

func (a *App) stepClock(steps int) {
	c := &a.clk
	if c.edit == nil || steps == 0 {
		return
	}
	unit := 1
	switch c.edit.CurrentSection() {
	case qt6.QDateTimeEdit__HourSection:
		unit = 3600
	case qt6.QDateTimeEdit__MinuteSection:
		unit = 60
	}
	a.showSeconds(stepClockSeconds(a.clockSeconds(), unit, steps))
	a.onClockEdited()
}

func (a *App) onClockEdited() {
	c := &a.clk
	if a.silent || c.lock || c.edit == nil {
		return
	}
	secs := a.clockSeconds()
	if c.down && (!c.running || secs > c.origin) {
		c.origin = secs
		c.alerted = false
	}
	if a.onAppliedGame() {
		a.rebaseClock(secs)
	} else {
		if c.drafts == nil {
			c.drafts = map[string]int{}
		}
		c.drafts[a.clockKey(a.sysKey(), a.sys().Game)] = secs
	}
	a.updateApply()
}

func (a *App) clockEditing() bool {
	edit := a.clk.edit
	if edit == nil {
		return false
	}
	if edit.HasFocus() {
		return true
	}
	if le := edit.LineEdit(); le != nil && le.HasFocus() {
		return true
	}
	return false
}

func (a *App) onAppliedGame() bool {
	c := &a.clk
	return c.running && a.appliedSys == a.sysKey() && a.appliedGame == a.sys().Game
}

func (a *App) clockKey(sys, game string) string {
	return sys + "\x00" + game
}

func (a *App) noteClockDraft(sys, game string) {
	c := &a.clk
	if a.onAppliedGame() && sys == a.appliedSys && game == a.appliedGame {
		return
	}
	if c.drafts == nil {
		c.drafts = map[string]int{}
	}
	c.drafts[a.clockKey(sys, game)] = a.clockSeconds()
}

func (a *App) liveSeconds() int {
	c := &a.clk
	return displayedSeconds(c.down, c.start, c.end, time.Now())
}

func (a *App) rebaseClock(shown int) {
	c := &a.clk
	if shown < 0 {
		shown = 0
	}
	now := time.Now()
	if c.down {
		if shown > c.origin {
			c.origin = shown
		}
		c.end = now.Add(time.Duration(shown) * time.Second)
		c.start = c.end.Add(-time.Duration(c.origin) * time.Second)
		return
	}
	c.start = now.Add(-time.Duration(shown) * time.Second)
	c.end = time.Time{}
}

func (a *App) presenceTimestamps() (start, end int64) {
	c := &a.clk
	if !(c.running || c.alerted) {
		return 0, 0
	}
	if !c.start.IsZero() {
		start = c.start.Unix()
	}
	if c.down && !c.end.IsZero() {
		end = c.end.Unix()
	}
	return start, end
}

func (a *App) showClockForCurrent() {
	c := &a.clk
	viewing := a.appliedSys == a.sysKey() && a.appliedGame == a.sys().Game
	if viewing && (c.running || c.alerted) && (!c.start.IsZero() || !c.end.IsZero()) {
		a.showSeconds(a.liveSeconds())
		return
	}
	a.showSeconds(c.drafts[a.clockKey(a.sysKey(), a.sys().Game)])
}

func (a *App) syncStartFromClock() {
	a.clk.start = time.Now().Add(-time.Duration(a.clockSeconds()) * time.Second)
	a.clk.end = time.Time{}
}

func (a *App) commitClockForApply() {
	c := &a.clk
	same := a.onAppliedGame()
	if c.down {
		shown := a.clockSeconds()
		if !same || shown > c.origin {
			c.origin = shown
		}
		c.alerted = false
		a.rebaseClock(shown)
		return
	}
	if !same && a.sys().TimePreserve && c.running && !c.start.IsZero() {
		return
	}
	a.syncStartFromClock()
}

func (a *App) toggleClockMode() {
	c := &a.clk
	c.down = !c.down
	c.alerted = false
	a.refreshClockMode()
	shown := a.clockSeconds()
	if c.down {
		c.origin = shown
		if c.running {
			now := time.Now()
			c.end = now.Add(time.Duration(shown) * time.Second)
			c.start = now
		}
		return
	}
	if c.running {
		a.syncStartFromClock()
	}
}

func (a *App) resetClock() {
	c := &a.clk
	c.alerted = false
	if c.down {
		a.showSeconds(c.origin)
		if c.running {
			now := time.Now()
			c.end = now.Add(time.Duration(c.origin) * time.Second)
			c.start = now
			a.armClockTick()
		}
		return
	}
	a.showSeconds(0)
	if c.running {
		c.start = time.Now()
		c.end = time.Time{}
		a.armClockTick()
	}
}

func (a *App) refreshClockMode() {
	c := &a.clk
	if c.mode == nil {
		return
	}
	if c.down {
		c.mode.SetIcon(iconNamed("chronometer", "chronometer"))
		c.mode.SetToolTip(a.tr.T("CLOCK_COUNT_DOWN"))
		return
	}
	c.mode.SetIcon(iconNamed("player-time", "player-time"))
	c.mode.SetToolTip(a.tr.T("CLOCK_COUNT_UP"))
}

func (a *App) updateElapsed() {
	c := &a.clk
	if !c.running {
		return
	}
	now := time.Now()
	if c.down && !c.end.IsZero() && !now.Before(c.end) {
		if a.onAppliedGame() && !a.clockEditing() {
			a.showSeconds(0)
		}
		a.finishCountdown()
		return
	}
	if a.onAppliedGame() && !a.clockEditing() {
		a.showSeconds(a.liveSeconds())
	}
}

func (a *App) finishCountdown() {
	c := &a.clk
	if c.alerted {
		return
	}
	c.alerted = true
	if a.onAppliedGame() {
		a.showSeconds(0)
	}
	c.running = false
	if c.timer != nil {
		c.timer.Stop()
	}
	if a.win != nil {
		qt6.QApplication_Alert(a.win.QWidget)
	}
}

func (a *App) ensureElapsedTick() {
	a.clk.running = true
	a.armClockTick()
}

func (a *App) armClockTick() {
	c := &a.clk
	if c.timer == nil || !c.running || c.timer.IsActive() {
		return
	}
	c.timer.Start2()
}
