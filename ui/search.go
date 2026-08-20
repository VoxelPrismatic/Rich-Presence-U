package ui

import (
	"context"
	"strings"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/nso"
)

func (a *App) setupGameSearch() {
	a.game.SetMaxVisibleItems(8)
	a.game.SetSizeAdjustPolicy(qt6.QComboBox__AdjustToMinimumContentsLengthWithIcon)
	a.game.SetMinimumContentsLength(12)

	a.completer = a.game.Completer()
	if a.completer != nil {
		a.completer.SetCompletionMode(qt6.QCompleter__PopupCompletion)
		a.completer.SetFilterMode(qt6.MatchContains)
		a.completer.SetCaseSensitivity(qt6.CaseInsensitive)
		a.completer.SetMaxVisibleItems(8)
		a.completer.SetWrapAround(true)
		a.completer.OnActivated(func(text string) {
			a.pickCompletion(text)
		})
		a.completer.OnHighlighted(func(text string) {
			a.completerHighlight = text
		})
	}
	if le := a.game.LineEdit(); le != nil {
		le.OnTextEdited(func(text string) {
			if a.silent {
				return
			}
			a.onGameTyped(text)
		})
	}
	a.game.OnActivated(func(index int) {
		if a.silent {
			return
		}
		a.pickGameIndex(index)
	})
	a.game.OnTextHighlighted(func(text string) {
		a.completerHighlight = text
	})

	a.searchTimer = qt6.NewQTimer2(a.game.QObject)
	a.searchTimer.SetSingleShot(true)
	a.searchTimer.OnTimeout(func() { a.runStoreSearch() })

	a.refillGameCombo()
}

func (a *App) refillGameCombo() {
	a.setSearchHits(a.nso.Games(a.settings.System), false)
}

func (a *App) setSearchHits(games []nso.Game, show bool) {
	if a.game == nil {
		return
	}
	a.searchHits = games
	typed := a.game.CurrentText()
	cursor := 0
	if le := a.game.LineEdit(); le != nil {
		cursor = le.CursorPosition()
	}
	keep := a.completerHighlight
	prev := a.silent
	a.silent = true
	a.game.Clear()
	region := a.preferredRegion()
	for _, g := range games {
		a.game.AddItem3(g.Title(region), qt6.NewQVariant11(g.ID))
	}
	a.game.SetEditText(typed)
	if le := a.game.LineEdit(); le != nil {
		le.SetCursorPosition(cursor)
	}
	a.silent = prev
	if !show || a.completer == nil || !a.gameSearchOpen() || strings.TrimSpace(typed) == "" || a.game.Count() == 0 {
		return
	}
	a.completer.Complete()
	a.restoreCompletion(keep)
}

func (a *App) pickGameIndex(index int) {
	if a.game == nil || index < 0 || index >= a.game.Count() {
		return
	}
	if id := a.game.ItemData(index).ToString(); id != "" {
		a.rememberAndSet(id)
		return
	}
	a.pickCompletion(a.game.ItemText(index))
}

func (a *App) rememberAndSet(id string) {
	for _, g := range a.searchHits {
		if g.ID == id {
			_ = a.nso.Remember(a.settings.System, g)
			break
		}
	}
	a.silent = true
	a.setGameID(id)
	a.silent = false
}

func (a *App) pickCompletion(text string) {
	if id := a.gameIDForTitle(text); id != "" {
		a.rememberAndSet(id)
		return
	}
	region := a.preferredRegion()
	for _, g := range a.searchHits {
		if g.Title(region) == text {
			_ = a.nso.Remember(a.settings.System, g)
			a.silent = true
			a.setGameID(g.ID)
			a.silent = false
			return
		}
	}
	hits := nso.Search(a.searchHits, text, region)
	for _, h := range hits {
		if h.Exact || h.DisplayTitle == text {
			_ = a.nso.Remember(a.settings.System, h.Game)
			a.silent = true
			a.setGameID(h.Game.ID)
			a.silent = false
			return
		}
	}
}

func (a *App) gameIDForTitle(text string) string {
	if a.game == nil || text == "" {
		return ""
	}
	idx := a.game.FindText(text)
	if idx < 0 {
		return ""
	}
	return a.game.ItemData(idx).ToString()
}

func (a *App) scheduleGameSearch() {
	a.searchGen++
	if a.searchTimer != nil {
		a.searchTimer.Start(250)
	}
}

func (a *App) runStoreSearch() {
	text := strings.TrimSpace(a.game.CurrentText())
	if len(text) < 2 {
		a.setSearchHits(a.nso.Games(a.settings.System), false)
		return
	}
	gen := a.searchGen
	sys := a.settings.System
	region := a.preferredRegion()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		extra, err := a.nso.SearchStore(ctx, text, sys)
		if err != nil {
			a.debug("eshop search: %v", err)
			extra = nil
		}
		mainthread.Start(func() {
			if gen != a.searchGen || strings.TrimSpace(a.game.CurrentText()) != text {
				return
			}
			local := nso.Search(a.nso.Games(sys), text, region)
			a.setSearchHits(nso.MergeGames(local, extra, 0), true)
		})
	}()
}

func (a *App) gameSearchOpen() bool {
	if a.game == nil {
		return false
	}
	if a.game.HasFocus() {
		return true
	}
	if le := a.game.LineEdit(); le != nil && le.HasFocus() {
		return true
	}
	if a.completer != nil {
		if pop := a.completer.Popup(); pop != nil && pop.IsVisible() {
			return true
		}
	}
	if v := a.game.View(); v != nil && v.IsVisible() {
		return true
	}
	return false
}

func (a *App) restoreCompletion(keep string) {
	if a.completer == nil || keep == "" {
		return
	}
	n := a.completer.CompletionCount()
	for i := 0; i < n; i++ {
		if !a.completer.SetCurrentRow(i) {
			continue
		}
		if a.completer.CurrentCompletion() == keep {
			if pop := a.completer.Popup(); pop != nil {
				pop.SetCurrentIndex(a.completer.CurrentIndex())
			}
			a.completerHighlight = keep
			return
		}
	}
}
