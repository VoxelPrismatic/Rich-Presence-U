package ui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/igdb"
)

// platformModel is a QAbstractItemModel over igdb.PickerTree.
// Internal ids are 1-based indexes into nodes (0 is the invisible root).
type platformModel struct {
	*qt6.QAbstractItemModel
	nodes []platNode
	roots []uintptr
}

type platNode struct {
	parent   uintptr
	row      int
	label    string
	children []uintptr
	platform igdb.Platform
}

func newPlatformModel(parent *qt6.QObject) *platformModel {
	m := &platformModel{
		QAbstractItemModel: qt6.NewQAbstractItemModel2(parent),
		nodes:              []platNode{{}}, // id 0 unused
	}
	for _, n := range igdb.PickerTree() {
		m.roots = append(m.roots, m.addNode(0, len(m.roots), n))
	}
	m.OnIndex(m.index)
	m.OnParent(m.parent)
	m.OnRowCount(m.rowCount)
	m.OnColumnCount(func(*qt6.QModelIndex) int { return 1 })
	m.OnHasChildren(func(_ func(*qt6.QModelIndex) bool, parent *qt6.QModelIndex) bool {
		return m.rowCount(parent) > 0
	})
	m.OnData(m.data)
	m.OnFlags(m.flags)
	return m
}

func (m *platformModel) addNode(parent uintptr, row int, src igdb.Node) uintptr {
	id := uintptr(len(m.nodes))
	n := platNode{
		parent:   parent,
		row:      row,
		label:    src.Label,
		platform: src.Platform,
	}
	m.nodes = append(m.nodes, n)
	for i, child := range src.Children {
		m.nodes[id].children = append(m.nodes[id].children, m.addNode(id, i, child))
	}
	return id
}

func (m *platformModel) nodeID(index *qt6.QModelIndex) uintptr {
	if index == nil || !index.IsValid() {
		return 0
	}
	id := index.InternalId()
	if id == 0 || int(id) >= len(m.nodes) {
		return 0
	}
	return id
}

func (m *platformModel) node(index *qt6.QModelIndex) *platNode {
	id := m.nodeID(index)
	if id == 0 {
		return nil
	}
	return &m.nodes[id]
}

func (m *platformModel) makeIndex(row int, id uintptr) *qt6.QModelIndex {
	idx := m.CreateIndex2(row, 0, id)
	return &idx
}

func (m *platformModel) index(row, column int, parent *qt6.QModelIndex) *qt6.QModelIndex {
	if column != 0 || row < 0 {
		return qt6.NewQModelIndex()
	}
	pid := m.nodeID(parent)
	var kids []uintptr
	if pid == 0 {
		kids = m.roots
	} else {
		kids = m.nodes[pid].children
	}
	if row >= len(kids) {
		return qt6.NewQModelIndex()
	}
	return m.makeIndex(row, kids[row])
}

func (m *platformModel) parent(child *qt6.QModelIndex) *qt6.QModelIndex {
	n := m.node(child)
	if n == nil || n.parent == 0 {
		return qt6.NewQModelIndex()
	}
	p := m.nodes[n.parent]
	return m.makeIndex(p.row, n.parent)
}

func (m *platformModel) rowCount(parent *qt6.QModelIndex) int {
	pid := m.nodeID(parent)
	if pid == 0 {
		return len(m.roots)
	}
	return len(m.nodes[pid].children)
}

func (m *platformModel) data(index *qt6.QModelIndex, role int) *qt6.QVariant {
	n := m.node(index)
	if n == nil {
		return qt6.NewQVariant()
	}
	if role != int(qt6.DisplayRole) {
		return qt6.NewQVariant()
	}
	label := n.label
	if len(n.children) > 0 {
		label += "  ›"
	}
	return qt6.NewQVariant11(label)
}

func (m *platformModel) flags(_ func(*qt6.QModelIndex) qt6.ItemFlag, index *qt6.QModelIndex) qt6.ItemFlag {
	n := m.node(index)
	if n == nil {
		return qt6.NoItemFlags
	}
	f := qt6.ItemIsEnabled | qt6.ItemIsSelectable
	if len(n.children) == 0 {
		f |= qt6.ItemNeverHasChildren
	}
	return f
}

func (m *platformModel) indexForSlug(slug string) *qt6.QModelIndex {
	if slug == "" {
		return qt6.NewQModelIndex()
	}
	for id := uintptr(1); id < uintptr(len(m.nodes)); id++ {
		if m.nodes[id].platform.Slug == slug {
			return m.makeIndex(m.nodes[id].row, id)
		}
	}
	return qt6.NewQModelIndex()
}

func (m *platformModel) pathLabels(index *qt6.QModelIndex) []string {
	var parts []string
	for id := m.nodeID(index); id != 0; id = m.nodes[id].parent {
		parts = append([]string{m.nodes[id].label}, parts...)
	}
	return parts
}

// soleLeaf returns n if it is a console, or the only nested console when n
// has exactly one child path (e.g. GameCube). Multiple children means drill.
func (m *platformModel) soleLeaf(n *platNode) *platNode {
	if n == nil {
		return nil
	}
	if len(n.children) == 0 {
		if n.platform.Slug != "" || n.platform.ID != 0 {
			return n
		}
		return nil
	}
	if len(n.children) != 1 {
		return nil
	}
	return m.soleLeaf(&m.nodes[n.children[0]])
}
