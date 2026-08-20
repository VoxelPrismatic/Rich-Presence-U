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
	a.completerModel = qt6.NewQStringListModel3(a.game.QObject)
	a.completer = qt6.NewQCompleter5(a.completerModel.QAbstractItemModel, a.game.QObject)
	a.completer.SetCompletionMode(qt6.QCompleter__PopupCompletion)
	a.completer.SetFilterMode(qt6.MatchContains)
	a.completer.SetCaseSensitivity(qt6.CaseInsensitive)
	a.completer.SetMaxVisibleItems(8)
	a.completer.SetWrapAround(true)
	a.game.SetCompleter(a.completer)
	a.completer.OnActivated(func(text string) {
		a.pickCompletion(text)
	})
	a.completer.OnHighlighted(func(text string) {
		a.completerHighlight = text
	})

	a.searchTimer = qt6.NewQTimer2(a.game.QObject)
	a.searchTimer.SetSingleShot(true)
	a.searchTimer.OnTimeout(func() { a.runStoreSearch() })

	a.refillCompleter()
}

func (a *App) refillCompleter() {
	if a.completerModel == nil {
		return
	}
	games := a.nso.Games(a.settings.System)
	a.searchHits = games
	titles := make([]string, 0, len(games))
	region := a.preferredRegion()
	for _, g := range games {
		titles = append(titles, g.Title(region))
	}
	a.completerModel.SetStringList(titles)
}

func (a *App) pickCompletion(text string) {
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
	hits := nso.Search(a.nso.Games(a.settings.System), text, region)
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

func (a *App) scheduleGameSearch() {
	a.searchGen++
	if a.searchTimer != nil {
		a.searchTimer.Start(250)
	}
}

func (a *App) runStoreSearch() {
	text := strings.TrimSpace(a.game.Text())
	if len(text) < 2 {
		return
	}
	gen := a.searchGen
	sys := a.settings.System
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		extra, err := a.nso.SearchStore(ctx, text, sys)
		if err != nil {
			a.debug("eshop search: %v", err)
			return
		}
		if len(extra) == 0 {
			return
		}
		mainthread.Start(func() {
			if gen != a.searchGen || strings.TrimSpace(a.game.Text()) != text {
				return
			}
			a.mergeStoreGames(extra)
		})
	}()
}

func (a *App) mergeStoreGames(extra []nso.Game) {
	if a.completerModel == nil {
		return
	}
	region := a.preferredRegion()
	seen := map[string]bool{}
	titles := make([]string, 0, len(a.searchHits)+len(extra))
	for _, g := range a.searchHits {
		t := g.Title(region)
		seen[strings.ToLower(t)] = true
		titles = append(titles, t)
	}
	changed := false
	for _, g := range extra {
		t := g.Title(region)
		key := strings.ToLower(t)
		if seen[key] {
			continue
		}
		seen[key] = true
		titles = append(titles, t)
		a.searchHits = append(a.searchHits, g)
		changed = true
	}
	if !changed {
		return
	}
	keep := a.completerHighlight
	if keep == "" && a.completer != nil {
		keep = a.completer.CurrentCompletion()
	}
	a.completerModel.SetStringList(titles)
	if a.game.HasFocus() && strings.TrimSpace(a.game.Text()) != "" {
		a.completer.Complete()
		a.restoreCompletion(keep)
	}
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
