package ui

import (
	"context"
	"fmt"
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
	// The popup takes keyboard focus when it opens. The line edit stays
	// drawn as focused, but keystrokes land on the list unless forwarded.
	a.installGameTypeForwarding()
	a.installGameThumbDelegate()

	a.game.OnKeyPressEvent(func(super func(e *qt6.QKeyEvent), e *qt6.QKeyEvent) {
		if a.gameSearchNav(e) {
			e.Accept()
			return
		}
		a.restoreTypedForEdit()
		super(e)
	})
	if le := a.game.LineEdit(); le != nil {
		le.OnTextEdited(func(text string) {
			a.gameTyped = text
			a.gameNav = false
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

func (a *App) installGameTypeForwarding() {
	view := a.game.View()
	if view == nil {
		return
	}
	a.game.OnEventFilter(func(super func(watched *qt6.QObject, event *qt6.QEvent) bool, watched *qt6.QObject, event *qt6.QEvent) bool {
		if a.forwardGameSearchKey(event) {
			return true
		}
		return super(watched, event)
	})
	watch := []*qt6.QObject{view.QObject}
	if vp := view.Viewport(); vp != nil {
		watch = append(watch, vp.QObject)
	}
	if win := view.Window(); win != nil && win.UnsafePointer() != a.game.Window().UnsafePointer() {
		watch = append(watch, win.QObject)
	}
	for _, obj := range watch {
		if obj != nil {
			obj.InstallEventFilter(a.game.QObject)
		}
	}
}

func (a *App) gamePopupVisible() bool {
	view := a.game.View()
	return view != nil && view.IsVisible()
}

// gameSearchNavKey is true for keys the open list owns. ShortcutOverride
// for these is swallowed so the popup does not also accept the row.
func (a *App) gameSearchNavKey(key *qt6.QKeyEvent) bool {
	if key == nil || a.game == nil {
		return false
	}
	if key.Modifiers()&(qt6.AltModifier|qt6.ControlModifier|qt6.MetaModifier) != 0 {
		return false
	}
	switch qt6.Key(key.Key()) {
	case qt6.Key_Up, qt6.Key_Down, qt6.Key_PageUp, qt6.Key_PageDown:
		return true
	case qt6.Key_Escape, qt6.Key_Return, qt6.Key_Enter:
		return a.gamePopupVisible()
	default:
		return false
	}
}

func (a *App) gameSearchNav(key *qt6.QKeyEvent) bool {
	if !a.gameSearchNavKey(key) {
		return false
	}
	switch qt6.Key(key.Key()) {
	case qt6.Key_Up:
		a.moveGameHighlight(-1)
	case qt6.Key_Down:
		a.moveGameHighlight(1)
	case qt6.Key_PageUp:
		a.moveGameHighlight(-a.gamePageStep())
	case qt6.Key_PageDown:
		a.moveGameHighlight(a.gamePageStep())
	case qt6.Key_Escape:
		a.cancelGamePopup()
	case qt6.Key_Return, qt6.Key_Enter:
		a.acceptGameHighlight()
	}
	return true
}

func (a *App) gamePageStep() int {
	n := a.game.MaxVisibleItems()
	if n < 1 {
		return 8
	}
	return n
}

func (a *App) moveGameHighlight(delta int) {
	if a.game == nil || a.game.Count() == 0 || delta == 0 {
		return
	}
	if !a.gameNav {
		a.gameTyped = a.game.CurrentText()
		a.gameNav = true
	}
	row := a.highlightedGameRow()
	if row < 0 {
		if delta > 0 {
			row = 0
		} else {
			row = a.game.Count() - 1
		}
	} else {
		row += delta
		if row < 0 {
			row = 0
		}
		if row >= a.game.Count() {
			row = a.game.Count() - 1
		}
	}
	if !a.gamePopupVisible() {
		a.game.ShowPopup()
	}
	a.highlightGameRow(row)
}

func (a *App) highlightedGameRow() int {
	if a.game == nil {
		return -1
	}
	view := a.game.View()
	if view == nil {
		return -1
	}
	idx := view.CurrentIndex()
	if idx == nil || !idx.IsValid() {
		return -1
	}
	return idx.Row()
}

func (a *App) highlightGameRow(row int) {
	if a.game == nil || row < 0 || row >= a.game.Count() {
		return
	}
	view := a.game.View()
	model := a.game.Model()
	if view == nil || model == nil {
		return
	}
	mi := model.Index(row, 0, qt6.NewQModelIndex())
	if mi == nil || !mi.IsValid() {
		return
	}
	name := a.game.ItemText(row)
	view.SetCurrentIndex(mi)
	view.ScrollTo(mi, qt6.QAbstractItemView__EnsureVisible)
	a.gameHighlight = name
	// Show the highlighted title without treating it as a new search.
	prev := a.silent
	a.silent = true
	a.game.SetEditText(name)
	if le := a.game.LineEdit(); le != nil {
		le.End(false)
	}
	a.silent = prev
}

func (a *App) restoreTypedForEdit() {
	if !a.gameNav || a.game == nil {
		return
	}
	a.gameNav = false
	prev := a.silent
	a.silent = true
	a.game.SetEditText(a.gameTyped)
	if le := a.game.LineEdit(); le != nil {
		le.End(false)
	}
	a.silent = prev
}

func (a *App) cancelGamePopup() {
	restore := a.game.CurrentText()
	if a.gameNav {
		restore = a.gameTyped
	}
	a.gameNav = false
	prev := a.silent
	a.silent = true
	a.game.HidePopup()
	a.game.SetEditText(restore)
	if le := a.game.LineEdit(); le != nil {
		le.End(false)
	}
	a.silent = prev
}

func (a *App) acceptGameHighlight() {
	row := a.highlightedGameRow()
	if row < 0 || a.game == nil || row >= a.game.Count() {
		return
	}
	a.gameNav = false
	a.game.HidePopup()
	a.pickGameIndex(row)
}

func (a *App) forwardGameSearchKey(event *qt6.QEvent) bool {
	if a.game == nil || a.relayingGameKey || event == nil {
		return false
	}
	view := a.game.View()
	if view == nil || !view.IsVisible() {
		return false
	}
	switch event.Type() {
	case qt6.QEvent__ShortcutOverride:
		key := qt6.UnsafeNewQKeyEvent(event.UnsafePointer())
		return a.gameSearchNavKey(key)
	case qt6.QEvent__KeyPress:
		key := qt6.UnsafeNewQKeyEvent(event.UnsafePointer())
		if a.gameSearchNav(key) {
			return true
		}
		return a.relayGameSearchKey(event)
	case qt6.QEvent__KeyRelease:
		key := qt6.UnsafeNewQKeyEvent(event.UnsafePointer())
		if a.gameSearchNavKey(key) {
			return true
		}
		return a.relayGameSearchKey(event)
	case qt6.QEvent__InputMethod:
		return a.relayGameSearchKey(event)
	default:
		return false
	}
}

func (a *App) relayGameSearchKey(event *qt6.QEvent) bool {
	le := a.game.LineEdit()
	if le == nil {
		return false
	}
	a.restoreTypedForEdit()
	le = a.game.LineEdit()
	if le == nil {
		return false
	}
	a.relayingGameKey = true
	le.Event(event)
	a.relayingGameKey = false
	return true
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
	px := a.gameIconPx()
	a.gameThumbs = map[string]*qt6.QPixmap{}
	for _, g := range games {
		if pix := a.cachedGameThumb(g, px); pix != nil {
			a.gameThumbs[g.ID] = pix
		}
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
	if show {
		a.fetchGameThumbs(games, a.searchGen, px)
	}
}

func (a *App) pickGameIndex(index int) {
	if a.game == nil || index < 0 || index >= a.game.Count() {
		return
	}
	a.gameNav = false
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

func (a *App) searchQuery() string {
	if a.gameNav {
		return strings.TrimSpace(a.gameTyped)
	}
	if a.game == nil {
		return ""
	}
	return strings.TrimSpace(a.game.CurrentText())
}

func (a *App) runStoreSearch() {
	text := a.searchQuery()
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
			if gen != a.searchGen || a.searchQuery() != text {
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
		view.ScrollTo(mi, qt6.QAbstractItemView__EnsureVisible)
		a.gameHighlight = keep
	}
}

func (a *App) gameFontPx() int {
	if a.game == nil {
		return 16
	}
	h := a.game.FontMetrics().Height()
	if h < 1 {
		return 16
	}
	return h
}

func (a *App) gameIconPx() int {
	return a.gameFontPx() + 8
}

func (a *App) installGameThumbDelegate() {
	view := a.game.View()
	if view == nil {
		return
	}
	// The combo list ignores a delegate size hint. Item padding is what
	// actually makes each row taller than the font.
	px := a.gameIconPx()
	view.SetIconSize(qt6.NewQSize2(px, px))
	view.SetStyleSheet(fmt.Sprintf("::item { padding-top: 4px; padding-bottom: 4px; min-height: %dpx; }", px))
	d := qt6.NewQStyledItemDelegate2(view.QObject)
	d.OnPaint(func(super func(painter *qt6.QPainter, option *qt6.QStyleOptionViewItem, index *qt6.QModelIndex), painter *qt6.QPainter, option *qt6.QStyleOptionViewItem, index *qt6.QModelIndex) {
		pix := a.gameThumbAt(index)
		orig := option.Rect()
		pad := 0
		if pix != nil && !pix.IsNull() {
			pad = pix.Width() + 6
		}
		if pad > 0 && orig != nil {
			option.SetRect(*orig.Adjusted(pad, 0, 0, 0))
		}
		super(painter, option, index)
		if pad > 0 && orig != nil {
			y := orig.Y() + (orig.Height()-pix.Height())/2
			painter.DrawPixmap9(orig.X()+3, y, pix)
		}
	})
	view.SetItemDelegate(d.QAbstractItemDelegate)
}

func (a *App) gameThumbAt(index *qt6.QModelIndex) *qt6.QPixmap {
	if index == nil || !index.IsValid() || a.gameThumbs == nil {
		return nil
	}
	id := ""
	if v := index.DataWithRole(int(qt6.UserRole)); v != nil {
		id = v.ToString()
	}
	return a.gameThumbs[id]
}

func (a *App) cachedGameThumb(g nso.Game, px int) *qt6.QPixmap {
	if a.nso == nil {
		return nil
	}
	path, ok := a.nso.CachedCover(g, a.preferredRegion())
	if !ok {
		return nil
	}
	return gameThumbPixmap(path, px)
}

func gameThumbPixmap(path string, px int) *qt6.QPixmap {
	pix := qt6.NewQPixmap4(path)
	if pix == nil || pix.IsNull() || px < 1 {
		return nil
	}
	return maskPixmap(pix, px, 4)
}

func (a *App) fetchGameThumbs(games []nso.Game, gen, px int) {
	if a.nso == nil || len(games) == 0 {
		return
	}
	if a.thumbCancel != nil {
		a.thumbCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.thumbCancel = cancel
	region := a.preferredRegion()
	list := append([]nso.Game{}, games...)
	if len(list) > 30 {
		list = list[:30]
	}
	go func() {
		for _, g := range list {
			if ctx.Err() != nil {
				return
			}
			if _, ok := a.nso.CachedCover(g, region); ok {
				continue
			}
			if u, _ := g.Cover(region); u == "" {
				continue
			}
			cctx, stop := context.WithTimeout(ctx, 8*time.Second)
			path, err := a.nso.CoverPath(cctx, g, region)
			stop()
			if err != nil || path == "" || ctx.Err() != nil {
				continue
			}
			id := g.ID
			mainthread.Start(func() {
				a.setGameThumb(gen, id, path, px)
			})
		}
	}()
}

func (a *App) setGameThumb(gen int, id, path string, px int) {
	if a.game == nil || gen != a.searchGen || id == "" {
		return
	}
	pix := gameThumbPixmap(path, px)
	if pix == nil {
		return
	}
	for row := 0; row < a.game.Count(); row++ {
		if a.game.ItemData(row).ToString() != id {
			continue
		}
		if a.gameThumbs == nil {
			a.gameThumbs = map[string]*qt6.QPixmap{}
		}
		a.gameThumbs[id] = pix
		if v := a.game.View(); v != nil && v.Viewport() != nil {
			v.Viewport().Update()
		}
		return
	}
}
