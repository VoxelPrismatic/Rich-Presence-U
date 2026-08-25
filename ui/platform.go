package ui

import (
	"strings"

	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/igdb"
	"github.com/voxelprismatic/richpresenceu/nso"
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
	pop := qt6.NewQFrame3(a.win.QWidget, qt6.Popup)
	pop.SetFrameShape(qt6.QFrame__StyledPanel)
	pop.SetFrameShadow(qt6.QFrame__Raised)
	box := qt6.NewQVBoxLayout(pop.QWidget)
	box.SetContentsMargins(4, 4, 4, 4)
	box.SetSpacing(2)

	head := qt6.NewQHBoxLayout2()
	head.SetContentsMargins(0, 0, 0, 0)
	back := qt6.NewQToolButton2()
	back.SetIcon(iconNamed("go-previous", "go-previous"))
	back.SetAutoRaise(true)
	back.SetToolTip(a.tr.T("SETTINGS_BACK"))
	back.OnClicked(func() { a.platGoBack() })
	crumb := qt6.NewQLabel2()
	head.AddWidget(back.QWidget)
	head.AddWidget2(crumb.QWidget, 1)
	box.AddLayout(head.QLayout)

	search := qt6.NewQLineEdit(pop.QWidget)
	search.SetPlaceholderText(a.tr.T("PLATFORM_SEARCH"))
	search.SetClearButtonEnabled(true)
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
	box.AddWidget(search.QWidget)

	list := qt6.NewQListView(pop.QWidget)
	list.SetModel(a.platModel.QAbstractItemModel)
	list.SetEditTriggers(qt6.QAbstractItemView__NoEditTriggers)
	list.SetSelectionMode(qt6.QAbstractItemView__SingleSelection)
	list.SetUniformItemSizes(true)
	list.SetHorizontalScrollBarPolicy(qt6.ScrollBarAlwaysOff)
	list.SetFocusPolicy(qt6.StrongFocus)
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
	g := a.system.MapToGlobalWithQPoint(qt6.NewQPoint2(0, a.system.Height()))
	a.platPopup.MoveWithQPoint(g)
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
		q := ""
		if a.platSearch != nil {
			q = strings.TrimSpace(a.platSearch.Text())
		}
		if q == "" {
			q = a.tr.T("PLATFORM_SEARCH")
		}
		a.platCrumb.SetText(q)
		return
	}
	root := a.platList.RootIndex()
	a.platBack.SetEnabled(root != nil && root.IsValid())
	parts := a.platModel.pathLabels(root)
	if len(parts) == 0 {
		a.platCrumb.SetText(a.tr.T("GAME_SYSTEM"))
		return
	}
	a.platCrumb.SetText(strings.Join(parts, "  ›  "))
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
	if rowH <= 0 {
		rowH = 24
	}
	head := 8
	if a.platBack != nil {
		head += a.platBack.SizeHint().Height()
	}
	if a.platSearch != nil {
		head += a.platSearch.SizeHint().Height() + 6
	}
	w := a.system.Width()
	if w < 240 {
		w = 240
	}
	a.platList.SetMinimumHeight(rowH * rows)
	a.platPopup.SetFixedWidth(w)
	a.platPopup.AdjustSize()
	a.platPopup.Resize(w, head+rowH*rows+12)
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
