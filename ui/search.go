package ui

import (
	"context"
	"strings"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/igdb"
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
	sys, ok := a.catalogSystem()
	if !ok {
		a.setSearchHits(nil, false)
		return
	}
	a.setSearchHits(a.nso.Games(sys), false)
}

func (a *App) comboCaret() (text string, cursor, selStart, selLen int, focused, popup bool) {
	if a.game == nil {
		return
	}
	text = a.game.CurrentText()
	if v := a.game.View(); v != nil {
		popup = v.IsVisible()
	}
	le := a.game.LineEdit()
	if le == nil {
		return
	}
	focused = le.HasFocus()
	cursor = le.CursorPosition()
	if le.HasSelectedText() {
		selStart = le.SelectionStart()
		selLen = le.SelectionLength()
	}
	return
}

func (a *App) restoreComboCaret(text string, cursor, selStart, selLen int, focused bool) {
	if a.game == nil {
		return
	}
	a.game.SetEditText(text)
	le := a.game.LineEdit()
	if le == nil {
		return
	}
	if focused {
		le.SetFocus()
	}
	if selLen > 0 {
		le.SetSelection(selStart, selLen)
		return
	}
	if cursor < 0 {
		cursor = 0
	}
	le.SetCursorPosition(cursor)
}

func (a *App) setSearchHits(games []nso.Game, show bool) {
	if a.game == nil {
		return
	}
	a.searchHits = games
	text, cursor, selStart, selLen, focused, popup := a.comboCaret()
	keep := a.gameHighlight
	prev := a.silent
	a.silent = true
	a.game.Clear()
	region := a.preferredRegion()
	for _, g := range games {
		a.game.AddItem3(g.Title(region), qt6.NewQVariant11(g.ID))
	}
	a.restoreComboCaret(text, cursor, selStart, selLen, focused)
	if show && a.game.Count() > 0 && strings.TrimSpace(text) != "" && (focused || popup) {
		if !popup {
			a.game.ShowPopup()
		}
		a.restoreComboCaret(text, cursor, selStart, selLen, true)
		a.restoreGameHighlight(keep)
	}
	a.silent = prev
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
		if sys, ok := a.catalogSystem(); ok {
			_ = a.nso.Remember(sys, game)
		}
	}
	a.silent = true
	a.setGameID(id)
	a.silent = false
	if !found {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		filled := game
		if sys, ok := a.nsoSystem(); ok {
			next, err := a.nso.EnrichRegions(ctx, game, sys)
			if err != nil {
				a.debug("enrich regions: %v", err)
			} else {
				filled = next
				_ = a.nso.Remember(sys, filled)
			}
		}
		a.nso.CacheCovers(ctx, filled)
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
	sys, ok := a.catalogSystem()
	_, store := a.nsoSystem()
	igdbOn := a.igdbAPI != nil && a.igdbAPI.Configured() && !store
	region := a.settings.Region
	display := a.preferredRegion()
	var catalog []nso.Game
	if ok {
		catalog = a.nso.Games(sys)
	}
	remote := store || igdbOn
	if len([]rune(text)) < 2 || !remote {
		local := nso.Search(catalog, text, display)
		games := make([]nso.Game, 0, len(local))
		for _, h := range local {
			games = append(games, h.Game)
		}
		if text == "" {
			games = catalog
		}
		a.setSearchHits(games, text != "" && remote)
		return
	}
	gen := a.searchGen
	plat, _ := a.platform()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		var extra []nso.Game
		var err error
		if store {
			extra, err = a.nso.SearchStore(ctx, text, sys, region)
			if err != nil {
				a.debug("eshop search: %v", err)
				extra = nil
			}
		} else {
			hits, serr := a.igdbAPI.SearchGames(ctx, text, plat)
			if serr != nil {
				a.debug("igdb search: %v", serr)
			} else {
				extra = igdbHitsToGames(hits)
			}
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

func igdbHitsToGames(hits []igdb.GameHit) []nso.Game {
	out := make([]nso.Game, 0, len(hits))
	for _, h := range hits {
		if h.CatalogID() == "" || h.Name == "" {
			continue
		}
		g := nso.Game{
			ID:       h.CatalogID(),
			Titles:   map[nso.Region]string{nso.US: h.Name},
			Icons:    map[nso.Region]bool{},
			Covers:   map[nso.Region]string{},
			Stores:   map[nso.Region]string{},
			CoverArt: h.CoverURL,
		}
		if h.CoverURL != "" {
			g.Covers[nso.US] = h.CoverURL
			g.Icons[nso.US] = true
		}
		if h.URL != "" {
			g.Stores[nso.US] = h.URL
		}
		out = append(out, g)
	}
	return out
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
