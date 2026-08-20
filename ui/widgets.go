package ui

import "github.com/mappu/miqt/qt6"

func hline() *qt6.QFrame {
	f := qt6.NewQFrame2()
	f.SetFrameShape(qt6.QFrame__HLine)
	f.SetFrameShadow(qt6.QFrame__Sunken)
	return f
}

func setBigFont(w *qt6.QWidget, delta int) {
	f := w.Font()
	if f.PointSize() > 0 {
		f.SetPointSize(f.PointSize() + delta)
	} else {
		f.SetPointSize(12 + delta)
	}
	w.SetFont(f)
}

func addRow(parent *qt6.QVBoxLayout, widgets ...*qt6.QWidget) *qt6.QHBoxLayout {
	row := qt6.NewQHBoxLayout2()
	row.SetContentsMargins(0, 0, 0, 0)
	for _, w := range widgets {
		row.AddWidget(w)
	}
	parent.AddLayout(row.QLayout)
	return row
}

func spacerWidget() *qt6.QWidget {
	s := qt6.NewQWidget2()
	s.SetSizePolicy2(qt6.QSizePolicy__Expanding, qt6.QSizePolicy__Preferred)
	return s
}

func linkLabel(html string) *qt6.QLabel {
	l := qt6.NewQLabel3(html)
	l.SetOpenExternalLinks(true)
	l.SetTextInteractionFlags(qt6.TextBrowserInteraction)
	return l
}
