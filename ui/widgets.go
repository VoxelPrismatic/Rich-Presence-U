package ui

import (
	"strings"

	"github.com/mappu/miqt/qt6"
)

const coverSize = 160

func hline() *qt6.QFrame {
	f := qt6.NewQFrame2()
	f.SetFrameShape(qt6.QFrame__HLine)
	f.SetFrameShadow(qt6.QFrame__Sunken)
	return f
}

func vline() *qt6.QFrame {
	f := qt6.NewQFrame2()
	f.SetFrameShape(qt6.QFrame__VLine)
	f.SetFrameShadow(qt6.QFrame__Sunken)
	return f
}

func popupText(s string) string {
	for strings.Contains(s, "\n\n") {
		s = strings.ReplaceAll(s, "\n\n", "\n")
	}
	return s
}

func newFormGrid(parent *qt6.QWidget) *qt6.QGridLayout {
	var g *qt6.QGridLayout
	if parent != nil {
		g = qt6.NewQGridLayout(parent)
	} else {
		g = qt6.NewQGridLayout2()
	}
	g.SetContentsMargins(0, 0, 0, 0)
	g.SetHorizontalSpacing(12)
	g.SetColumnStretch(0, 1)
	g.SetColumnStretch(1, 1)
	return g
}

func addFormRow(g *qt6.QGridLayout, row int, title string, control *qt6.QWidget, help *qt6.QToolButton) {
	lab := qt6.NewQLabel3(title)
	lab.SetAlignment(qt6.AlignRight | qt6.AlignVCenter)
	g.AddWidget4(lab.QWidget, row, 0, qt6.AlignRight|qt6.AlignVCenter)
	if help == nil {
		g.AddWidget4(control, row, 1, qt6.AlignLeft|qt6.AlignVCenter)
		return
	}
	wrap := qt6.NewQWidget2()
	lay := qt6.NewQHBoxLayout(wrap)
	lay.SetContentsMargins(0, 0, 0, 0)
	lay.SetSpacing(4)
	lay.AddWidget(control)
	lay.AddWidget(help.QWidget)
	lay.AddStretch()
	g.AddWidget2(wrap, row, 1)
}

func maskPixmap(src *qt6.QPixmap, size int, radius float64) *qt6.QPixmap {
	if src == nil || src.IsNull() {
		return src
	}
	scaled := src.Scaled2(size, size, qt6.KeepAspectRatioByExpanding)
	x, y := 0, 0
	if scaled.Width() > size {
		x = (scaled.Width() - size) / 2
	}
	if scaled.Height() > size {
		y = (scaled.Height() - size) / 2
	}
	cropped := scaled.Copy(x, y, size, size)
	img := qt6.NewQImage3(size, size, qt6.QImage__Format_ARGB32_Premultiplied)
	img.Fill(0)
	p := qt6.NewQPainter2(img.QPaintDevice)
	p.SetRenderHint(qt6.QPainter__Antialiasing)
	path := qt6.NewQPainterPath()
	if radius <= 0 {
		path.AddEllipse2(0, 0, float64(size), float64(size))
	} else {
		path.AddRoundedRect2(0, 0, float64(size), float64(size), radius, radius)
	}
	p.SetClipPath(path)
	p.DrawPixmap9(0, 0, cropped)
	p.End()
	return qt6.QPixmap_FromImage(img)
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
