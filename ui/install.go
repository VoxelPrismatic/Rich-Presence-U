package ui

import (
	"github.com/mappu/miqt/qt6"
)

func applyAppIcon(logo []byte) {
	if len(logo) == 0 {
		return
	}
	pix := qt6.NewQPixmap()
	if !pix.LoadFromDataWithData(logo) || pix.IsNull() {
		return
	}
	ic := qt6.NewQIcon2(pix)
	qt6.QGuiApplication_SetWindowIcon(ic)
}

func (a *App) dialogParent() *qt6.QWidget {
	if a != nil && a.win != nil {
		return a.win.QWidget
	}
	return nil
}
