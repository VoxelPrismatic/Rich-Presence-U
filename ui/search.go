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
	a.game.SetCompleter(nil)

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
		a.gameHighlight = text
	})
	a.game.OnWheelEvent(func(super func(e *qt6.QWheelEvent), e *qt6.QWheelEvent) {
		if v := a.game.View(); v == nil || !v.IsVisible() {
			e.Ignore()
			return
		}
		super(e)
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
	keep := a.gameHighlight
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
	if !show || !a.gameSearchOpen() || strings.TrimSpace(typed) == "" || a.game.Count() == 0 {
		return
	}
	a.game.ShowPopup()
	a.silent = true
	a.game.SetEditText(typed)
	if le := a.game.LineEdit(); le != nil {
		le.SetCursorPosition(cursor)
	}
	a.silent = prev
	a.restoreGameHighlight(keep)
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
	var game nso.Game
	found := false
	for _, g := range a.searchHits {
		if g.ID == id {
			game = g
			found = true
			break
		}
	}
	if found {
		_ = a.nso.Remember(a.settings.System, game)
	}
	a.silent = true
	a.setGameID(id)
	a.silent = false
	if !found {
		return
	}
	sys := a.settings.System
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		filled, err := a.nso.EnrichRegions(ctx, game, sys)
		if err != nil {
			a.debug("enrich regions: %v", err)
			return
		}
		_ = a.nso.Remember(sys, filled)
		mainthread.Start(func() {
			if a.sys().Game != id {
				return
			}
			a.refreshGameUI()
			a.updateApply()
		})
	}()
}

func (a *App) pickCompletion(text string) {
	if id := a.gameIDForTitle(text); id != "" {
		a.rememberAndSet(id)
		return
	}
	region := a.preferredRegion()
	for _, g := range a.searchHits {
		if g.Title(region) == text {
			a.rememberAndSet(g.ID)
			return
		}
	}
	hits := nso.Search(a.searchHits, text, region)
	for _, h := range hits {
		if h.Exact || h.DisplayTitle == text {
			a.rememberAndSet(h.Game.ID)
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
	sys := a.settings.System
	region := a.settings.Region
	display := a.preferredRegion()
	if len([]rune(text)) < 2 {
		local := nso.Search(a.nso.Games(sys), text, display)
		games := make([]nso.Game, 0, len(local))
		for _, h := range local {
			games = append(games, h.Game)
		}
		if text == "" {
			games = a.nso.Games(sys)
		}
		a.setSearchHits(games, text != "")
		return
	}
	gen := a.searchGen
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		extra, err := a.nso.SearchStore(ctx, text, sys, region)
		if err != nil {
			a.debug("eshop search: %v", err)
			extra = nil
		}
		mainthread.Start(func() {
			if gen != a.searchGen || strings.TrimSpace(a.game.CurrentText()) != text {
				return
			}
			local := nso.Search(a.nso.Games(sys), text, display)
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
	if v := a.game.View(); v != nil && v.IsVisible() {
		return true
	}
	return false
}

func (a *App) restoreGameHighlight(keep string) {
	if keep == "" || a.game == nil {
		return
	}
	idx := a.game.FindText(keep)
	if idx < 0 {
		return
	}
	view := a.game.View()
	model := a.game.Model()
	if view == nil || model == nil {
		return
	}
	mi := model.Index(idx, 0, qt6.NewQModelIndex())
	if mi != nil && mi.IsValid() {
		view.SetCurrentIndex(mi)
		a.gameHighlight = keep
	}
}
