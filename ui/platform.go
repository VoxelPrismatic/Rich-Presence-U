package ui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/igdb"
	"github.com/voxelprismatic/richpresenceu/nso"
)

const (
	chevron          = "  ›  "
	platShadowPad    = 16
	platRowPad       = 5
	platArrowSlot    = 22
	platCornerRadius = 4
)

func (a *App) setupPlatformSelector() {
	a.platModel = newPlatformModel(a.system.QObject)
	a.system.SetToolTip(a.tr.T("GAME_SYSTEM"))
	a.system.SetSizeAdjustPolicy(qt6.QComboBox__AdjustToMinimumContentsLengthWithIcon)
	a.system.SetMinimumContentsLength(12)
	a.system.SetMaxVisibleItems(1)
	a.system.OnShowPopup(func(func()) { a.showPlatPopup() })
	a.system.OnHidePopup(func(func()) {
		if a.platPopup != nil && a.platPopup.IsVisible() {
			a.platPopup.Hide()
		}
	})
	a.system.OnWheelEvent(func(_ func(e *qt6.QWheelEvent), e *qt6.QWheelEvent) {
		e.Ignore()
	})
	a.buildPlatPopup()
	a.refreshPlatformButton()
}

func (a *App) buildPlatPopup() {
	pop := qt6.NewQWidget3(a.win.QWidget, qt6.Popup|qt6.FramelessWindowHint)
	pop.SetAttribute(qt6.WA_TranslucentBackground)
	outer := qt6.NewQVBoxLayout(pop)
	outer.SetContentsMargins(platShadowPad, platShadowPad, platShadowPad, platShadowPad)
	outer.SetSpacing(0)

	wrap := qt6.NewQWidget2()
	shadow := qt6.NewQGraphicsDropShadowEffect2(wrap.QObject)
	shadow.SetBlurRadius(24)
	shadow.SetOffset2(0, 4)
	shadow.SetColor(qt6.NewQColor11(0, 0, 0, 100))
	wrap.SetGraphicsEffect(shadow.QGraphicsEffect)
	outer.AddWidget(wrap)

	inner := qt6.NewQFrame2()
	inner.SetObjectName(*qt6.NewQAnyStringView3("platPanel"))
	inner.SetFrameShape(qt6.QFrame__NoFrame)
	inner.SetAutoFillBackground(true)
	inner.SetAttribute(qt6.WA_StyledBackground)
	inner.SetStyleSheet(fmt.Sprintf(
		"#platPanel { background-color: palette(window); border: 1px solid palette(mid); border-radius: %dpx; }",
		platCornerRadius,
	))
	wl := qt6.NewQVBoxLayout(wrap)
	wl.SetContentsMargins(0, 0, 0, 0)
	wl.AddWidget(inner.QWidget)

	box := qt6.NewQVBoxLayout(inner.QWidget)
	box.SetContentsMargins(4, 4, 4, 4)
	box.SetSpacing(0)

	header := qt6.NewQWidget2()
	header.SetObjectName(*qt6.NewQAnyStringView3("platHeader"))
	header.SetAutoFillBackground(true)
	header.SetStyleSheet("#platHeader { background-color: palette(window); border: none; }")
	hBox := qt6.NewQVBoxLayout(header)
	hBox.SetContentsMargins(0, 0, 0, 4)
	hBox.SetSpacing(4)

	head := qt6.NewQHBoxLayout2()
	head.SetContentsMargins(0, 0, 0, 0)
	back := qt6.NewQToolButton2()
	back.SetIcon(iconNamed("go-previous", "go-previous"))
	back.SetAutoRaise(true)
	back.SetToolButtonStyle(qt6.ToolButtonIconOnly)
	back.SetFocusPolicy(qt6.StrongFocus)
	back.SetStyleSheet("QToolButton { border: 1px solid transparent; background: transparent; border-radius: 4px; } QToolButton:hover:enabled { border: 1px solid palette(highlight); } QToolButton:pressed:enabled { border: 1px solid palette(highlight); background: palette(highlight); }")
	back.SetToolTip(a.tr.T("SETTINGS_BACK"))
	back.OnClicked(func() { a.platGoBack() })
	crumb := qt6.NewQLabel2()
	crumb.SetFrameShape(qt6.QFrame__NoFrame)
	head.AddWidget(back.QWidget)
	head.AddWidget2(crumb.QWidget, 1)
	hBox.AddLayout(head.QLayout)

	search := qt6.NewQLineEdit(header)
	search.SetPlaceholderText(a.tr.T("PLATFORM_SEARCH"))
	search.SetClearButtonEnabled(true)
	search.AddAction2(iconFind(), qt6.QLineEdit__LeadingPosition)
	sp := search.Palette()
	sp.SetColor2(qt6.QPalette__Base, sp.ColorWithCr(qt6.QPalette__Base))
	search.SetPalette(sp)
	search.OnTextChanged(func(text string) { a.filterPlatSearch(text) })
	search.OnReturnPressed(func() { a.activatePlatCurrent() })
	search.OnKeyPressEvent(func(super func(param1 *qt6.QKeyEvent), param1 *qt6.QKeyEvent) {
		switch qt6.Key(param1.Key()) {
		case qt6.Key_Down:
			a.focusPlatList()
			param1.Accept()
			return
		case qt6.Key_Escape:
			if strings.TrimSpace(search.Text()) != "" {
				search.Clear()
				param1.Accept()
				return
			}
			if !a.platAtRoot() {
				a.platGoBack()
				param1.Accept()
				return
			}
			pop.Hide()
			param1.Accept()
			return
		}
		super(param1)
	})
	hBox.AddWidget(search.QWidget)
	box.AddWidget(header)

	list := qt6.NewQListView(inner.QWidget)
	list.SetModel(a.platModel.QAbstractItemModel)
	list.SetEditTriggers(qt6.QAbstractItemView__NoEditTriggers)
	list.SetSelectionMode(qt6.QAbstractItemView__SingleSelection)
	list.SetUniformItemSizes(true)
	list.SetHorizontalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	list.SetFocusPolicy(qt6.StrongFocus)
	list.SetFrameShape(qt6.QFrame__NoFrame)
	list.SetIconSize(qt6.NewQSize2(16, 16))
	list.SetAutoFillBackground(true)
	list.SetStyleSheet("QListView, QListView::viewport { background-color: palette(base); border: none; } QListView::item { padding: 4px 5px; }")
	list.SetItemDelegate(newPlatRowDelegate(list.QObject).QAbstractItemDelegate)
	list.OnClicked(func(index *qt6.QModelIndex) { a.onPlatIndex(index) })
	list.OnActivated(func(index *qt6.QModelIndex) { a.onPlatIndex(index) })
	list.OnKeyPressEvent(func(super func(event *qt6.QKeyEvent), event *qt6.QKeyEvent) {
		switch qt6.Key(event.Key()) {
		case qt6.Key_Escape:
			if a.platSearching {
				search.Clear()
				search.SetFocus()
				event.Accept()
				return
			}
			if a.platAtRoot() {
				pop.Hide()
				return
			}
			a.platGoBack()
			event.Accept()
			return
		case qt6.Key_Backspace:
			if a.platSearching || a.platAtRoot() {
				super(event)
				return
			}
			a.platGoBack()
			event.Accept()
			return
		}
		super(event)
	})
	box.AddWidget(list.QWidget)

	a.platPopup = pop
	a.platList = list
	a.platBack = back
	a.platCrumb = crumb
	a.platSearch = search
	a.platHits = qt6.NewQStandardItemModel3(pop.QObject)
}

func newPlatRowDelegate(parent *qt6.QObject) *qt6.QStyledItemDelegate {
	arrow := iconGoNext()
	d := qt6.NewQStyledItemDelegate2(parent)
	d.OnPaint(func(super func(painter *qt6.QPainter, option *qt6.QStyleOptionViewItem, index *qt6.QModelIndex), painter *qt6.QPainter, option *qt6.QStyleOptionViewItem, index *qt6.QModelIndex) {
		orig := option.Rect()
		if platRowReserveArrow(index) && orig != nil {
			clipped := orig.Adjusted(0, 0, -platArrowSlot, 0)
			option.SetRect(*clipped)
			super(painter, option, index)
			if platRowHasArrow(index) {
				pix := arrow.Pixmap2(16, 16)
				x := orig.Right() - 18
				y := orig.Center().Y() - 8
				painter.DrawPixmap9(x, y, pix)
			}
			return
		}
		super(painter, option, index)
	})
	d.OnSizeHint(func(super func(option *qt6.QStyleOptionViewItem, index *qt6.QModelIndex) *qt6.QSize, option *qt6.QStyleOptionViewItem, index *qt6.QModelIndex) *qt6.QSize {
		sz := super(option, index)
		if sz == nil {
			return qt6.NewQSize2(0, 24+platRowPad)
		}
		sz.SetHeight(sz.Height() + platRowPad)
		if platRowReserveArrow(index) {
			sz.SetWidth(sz.Width() + platArrowSlot)
		}
		return sz
	})
	return d
}

func platRowHasArrow(index *qt6.QModelIndex) bool {
	if index == nil || !index.IsValid() {
		return false
	}
	m := index.Model()
	return m != nil && m.HasChildren(index)
}

// platRowReserveArrow keeps a trailing slot on mixed pages (folders + leaves)
// so the highlight bar does not change width between rows.
func platRowReserveArrow(index *qt6.QModelIndex) bool {
	if platRowHasArrow(index) {
		return true
	}
	if index == nil || !index.IsValid() {
		return false
	}
	m := index.Model()
	if m == nil {
		return false
	}
	parent := m.Parent(index)
	n := m.RowCount(parent)
	for i := 0; i < n; i++ {
		if m.HasChildren(m.Index(i, 0, parent)) {
			return true
		}
	}
	return false
}

func (a *App) showPlatPopup() {
	if a.platPopup == nil || a.system == nil {
		return
	}
	a.platHasSaved = false
	if a.platSearch != nil {
		a.platSearch.SetText("")
	}
	a.showPlatTree()
	a.sizePlatPopup()
	a.platPopup.Show()
	if a.platSearch != nil {
		a.platSearch.SetFocus()
	}
}

func (a *App) showPlatTree() {
	a.platSearching = false
	a.platList.SetModel(a.platModel.QAbstractItemModel)
	if a.platHasSaved {
		a.platHasSaved = false
		if a.platSavedRoot == 0 {
			a.platList.SetRootIndex(qt6.NewQModelIndex())
		} else {
			n := a.platModel.nodes[a.platSavedRoot]
			idx := a.platModel.makeIndex(n.row, a.platSavedRoot)
			a.platList.SetRootIndex(idx)
		}
		a.updatePlatHeader()
		a.sizePlatPopup()
		return
	}
	sel := a.platModel.indexForSlug(a.settings.Platform)
	if sel != nil && sel.IsValid() {
		a.platList.SetRootIndex(a.platModel.Parent(sel))
		a.platList.SetCurrentIndex(sel)
	} else {
		a.platList.SetRootIndex(qt6.NewQModelIndex())
	}
	a.updatePlatHeader()
	a.sizePlatPopup()
}

func (a *App) platAtRoot() bool {
	if a.platList == nil {
		return true
	}
	root := a.platList.RootIndex()
	return root == nil || !root.IsValid()
}

func (a *App) platGoBack() {
	if a.platSearching {
		if a.platSearch != nil {
			a.platSearch.SetText("")
		}
		return
	}
	if a.platAtRoot() {
		return
	}
	parent := a.platModel.Parent(a.platList.RootIndex())
	a.platList.SetRootIndex(parent)
	a.updatePlatHeader()
	a.sizePlatPopup()
	if a.platModel.rowCount(parent) > 0 {
		a.platList.SetCurrentIndex(a.platModel.index(0, 0, parent))
	}
}

func (a *App) updatePlatHeader() {
	if a.platBack == nil || a.platCrumb == nil {
		return
	}
	if a.platSearching {
		a.platBack.SetEnabled(true)
		a.platCrumb.SetText(a.tr.T("PLATFORM_RESULTS"))
		return
	}
	root := a.platList.RootIndex()
	a.platBack.SetEnabled(root != nil && root.IsValid())
	parts := a.platModel.pathLabels(root)
	if len(parts) == 0 {
		a.platCrumb.SetText(a.tr.T("GAME_SYSTEM"))
		return
	}
	a.platCrumb.SetText(strings.Join(parts, chevron))
}

func (a *App) sizePlatPopup() {
	if a.platPopup == nil || a.platList == nil {
		return
	}
	rows := 1
	if a.platSearching && a.platHits != nil {
		rows = a.platHits.RowCount(qt6.NewQModelIndex())
	} else {
		rows = a.platModel.rowCount(a.platList.RootIndex())
	}
	if rows < 1 {
		rows = 1
	}
	if rows > 12 {
		rows = 12
	}
	rowH := a.platList.SizeHintForRow(0)
	if rowH < 22 {
		rowH = 24 + platRowPad
	}
	head := 16
	if a.platBack != nil {
		head += a.platBack.SizeHint().Height()
	}
	if a.platSearch != nil {
		head += a.platSearch.SizeHint().Height()
	}
	w := max(240, a.system.Width())
	listH := rowH * rows
	a.platList.SetFixedHeight(listH)
	innerH := head + listH + 4
	a.platPopup.SetFixedSize2(w+platShadowPad*2, innerH+platShadowPad*2)
	g := a.system.MapToGlobalWithQPoint(qt6.NewQPoint2(0, a.system.Height()))
	a.platPopup.Move(g.X()-platShadowPad, g.Y()-platShadowPad)
}

func (a *App) onPlatIndex(index *qt6.QModelIndex) {
	if index == nil || !index.IsValid() {
		return
	}
	if a.platSearching {
		slug := index.DataWithRole(int(qt6.UserRole)).ToString()
		p, ok := igdb.BySlug(slug)
		if !ok {
			return
		}
		a.pickPlatform(p)
		if a.platPopup != nil {
			a.platPopup.Hide()
		}
		return
	}
	n := a.platModel.node(index)
	if n == nil {
		return
	}
	if leaf := a.platModel.soleLeaf(n); leaf != nil {
		a.pickPlatform(leaf.platform)
		if a.platPopup != nil {
			a.platPopup.Hide()
		}
		return
	}
	if len(n.children) == 0 {
		return
	}
	a.platList.SetRootIndex(index)
	a.updatePlatHeader()
	a.sizePlatPopup()
	a.platList.SetCurrentIndex(a.platModel.index(0, 0, index))
}

func (a *App) filterPlatSearch(text string) {
	if a.platList == nil || a.platModel == nil {
		return
	}
	q := strings.TrimSpace(text)
	if q == "" {
		a.showPlatTree()
		return
	}
	if !a.platSearching {
		a.platSavedRoot = a.platModel.nodeID(a.platList.RootIndex())
		a.platHasSaved = true
	}
	hits := igdb.Search(q)
	a.platSearching = true
	a.platHits.Clear()
	for _, h := range hits {
		item := qt6.NewQStandardItem2(h.Label())
		item.SetEditable(false)
		item.SetData(qt6.NewQVariant11(h.Platform.Slug), int(qt6.UserRole))
		a.platHits.AppendRowWithItem(item)
	}
	a.platList.SetModel(a.platHits.QAbstractItemModel)
	if a.platHits.RowCount(qt6.NewQModelIndex()) > 0 {
		a.platList.SetCurrentIndex(a.platHits.Index(0, 0, qt6.NewQModelIndex()))
	}
	a.updatePlatHeader()
	a.sizePlatPopup()
}

func (a *App) focusPlatList() {
	if a.platList == nil {
		return
	}
	a.platList.SetFocus()
	cur := a.platList.CurrentIndex()
	if cur != nil && cur.IsValid() {
		return
	}
	model := a.platList.Model()
	if model == nil {
		return
	}
	first := model.Index(0, 0, a.platList.RootIndex())
	if first != nil && first.IsValid() {
		a.platList.SetCurrentIndex(first)
	}
}

func (a *App) activatePlatCurrent() {
	if a.platList == nil {
		return
	}
	a.onPlatIndex(a.platList.CurrentIndex())
}

func (a *App) pickPlatform(p igdb.Platform) {
	if p.Slug == "" && p.ID == 0 {
		return
	}
	if p.Slug == a.settings.Platform {
		a.refreshPlatformButton()
		return
	}
	a.settings.Platform = p.Slug
	if sys, ok := nso.ParseSystem(p.StoreCode()); ok {
		a.settings.System = sys
	} else {
		a.settings.System = ""
	}
	a.bumpElapsed(true)
	a.reloadSystem()
}

func (a *App) refreshPlatformButton() {
	if a.system == nil {
		return
	}
	name := a.platformDisplay()
	prev := a.silent
	a.silent = true
	if a.system.Count() == 0 {
		a.system.AddItem(name)
	} else {
		a.system.SetItemText(0, name)
	}
	a.system.SetCurrentIndex(0)
	a.silent = prev
}

func (a *App) platform() (igdb.Platform, bool) {
	if a.settings.Platform != "" {
		return igdb.BySlug(a.settings.Platform)
	}
	if slug := igdb.SlugForStoreCode(string(a.settings.System)); slug != "" {
		return igdb.BySlug(slug)
	}
	return igdb.Platform{}, false
}

func (a *App) platformDisplay() string {
	if p, ok := a.platform(); ok {
		return p.DisplayName()
	}
	if a.settings.System.Valid() {
		return a.settings.System.DisplayName()
	}
	return a.tr.T("GAME_SYSTEM")
}

// nsoSystem is the Nintendo title-database console used for eShop search.
// Only Switch consoles have an override; other picker entries return false.
func (a *App) nsoSystem() (nso.System, bool) {
	if p, ok := a.platform(); ok {
		sys, ok := nso.ParseSystem(p.StoreCode())
		return sys, ok
	}
	if a.settings.System.Valid() && (a.settings.System == nso.HAC || a.settings.System == nso.BEE) {
		return a.settings.System, true
	}
	return "", false
}

func (a *App) discordSystem() nso.System {
	if sys, ok := a.nsoSystem(); ok {
		return sys
	}
	return a.settings.System
}

func (a *App) updateInfoButton() {
	if a.infoBtn == nil {
		return
	}
	_, store := a.nsoSystem()
	if a.region != nil {
		a.region.SetVisible(store)
	}
	if store {
		a.infoBtn.SetIcon(iconShop())
		shop := a.tr.T("SHOP_ESHOP")
		a.infoBtn.SetToolTip(fmt.Sprintf(a.tr.T("OPEN_SHOP_PAGE"), shop))
		u := ""
		if game, ok := a.currentGame(); ok {
			u = game.Store(a.preferredRegion())
		}
		a.infoBtn.SetEnabled(u != "")
		return
	}
	a.infoBtn.SetIcon(iconIGDB())
	a.infoBtn.SetToolTip(a.tr.T("OPEN_IGDB_PAGE"))
	_, ok := a.platform()
	a.infoBtn.SetEnabled(ok)
}

func (a *App) openGameInfo() {
	if _, store := a.nsoSystem(); store {
		game, ok := a.currentGame()
		if !ok {
			return
		}
		u := game.Store(a.preferredRegion())
		if u == "" {
			return
		}
		openURL(u)
		return
	}
	p, ok := a.platform()
	if !ok {
		return
	}
	title := strings.TrimSpace(a.title())
	if title != "" && title != p.DisplayName() {
		openURL("https://www.igdb.com/search?q=" + url.QueryEscape(title))
		return
	}
	openURL(p.PageURL())
}

func openURL(u string) {
	if strings.TrimSpace(u) == "" {
		return
	}
	qt6.QDesktopServices_OpenUrl(qt6.NewQUrl3(u))
}
