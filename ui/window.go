package ui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/voxelprismatic/richpresenceu/nso"
)

func (a *App) buildWindow() {
	a.win = qt6.NewQMainWindow2()
	a.win.SetWindowTitle("Rich Presence U")
	a.win.SetMinimumSize2(480, 520)
	if a.settings.WindowW > 0 && a.settings.WindowH > 0 {
		a.win.Resize(a.settings.WindowW, a.settings.WindowH)
	}
	if a.settings.WindowX != 0 || a.settings.WindowY != 0 {
		a.win.Move(a.settings.WindowX, a.settings.WindowY)
	}

	a.stack = qt6.NewQStackedWidget2()
	a.stack.AddWidget(a.buildMain())
	a.stack.AddWidget(a.buildSettings())
	a.win.SetCentralWidget(a.stack.QWidget)

	a.win.OnCloseEvent(func(super func(event *qt6.QCloseEvent), event *qt6.QCloseEvent) {
		a.settings.WindowW = a.win.Width()
		a.settings.WindowH = a.win.Height()
		a.settings.WindowX = a.win.X()
		a.settings.WindowY = a.win.Y()
		a.persist()
		a.inhibit.Set(false)
		a.rpc.Close()
		super(event)
	})
}

func (a *App) buildMain() *qt6.QWidget {
	page := qt6.NewQWidget2()
	root := qt6.NewQVBoxLayout(page)
	root.SetContentsMargins(12, 12, 12, 12)

	scroll := qt6.NewQScrollArea2()
	scroll.SetWidgetResizable(true)
	scroll.SetFrameShape(qt6.QFrame__NoFrame)
	inner := qt6.NewQWidget2()
	box := qt6.NewQVBoxLayout(inner)
	box.SetContentsMargins(0, 0, 0, 0)
	box.AddWidget(a.buildPresence())
	box.AddSpacing(8)
	a.buildFriendCode(box)
	a.buildToggles(box)
	box.AddStretch()
	scroll.SetWidget(inner)
	root.AddWidget2(scroll.QWidget, 1)
	root.AddWidget(hline().QWidget)
	root.AddWidget(a.buildBar())
	return page
}

func (a *App) buildPresence() *qt6.QWidget {
	card := qt6.NewQFrame2()
	card.SetFrameShape(qt6.QFrame__StyledPanel)
	lay := qt6.NewQHBoxLayout(card.QWidget)
	lay.SetContentsMargins(12, 12, 12, 12)

	a.cover = qt6.NewQLabel2()
	a.cover.SetFixedSize2(96, 96)
	a.cover.SetScaledContents(true)
	a.cover.SetStyleSheet("background: rgba(0,0,0,40); border-radius: 8px;")
	lay.AddWidget(a.cover.QWidget)

	col := qt6.NewQVBoxLayout2()
	col.SetSpacing(8)

	sysRow := qt6.NewQHBoxLayout2()
	a.system = qt6.NewQComboBox2()
	a.system.SetToolTip(a.tr.T("GAME_SYSTEM"))
	setBigFont(a.system.QWidget, 2)
	for _, sys := range nso.Systems {
		a.system.AddItem3(sys.DisplayName(), qt6.NewQVariant11(string(sys)))
	}
	a.system.OnCurrentIndexChanged(func(index int) {
		if a.silent {
			return
		}
		id := a.system.ItemData(index).ToString()
		sys, ok := nso.ParseSystem(id)
		if !ok || sys == a.settings.System {
			return
		}
		a.settings.System = sys
		a.bumpElapsed(true)
		a.reloadSystem()
	})
	sysRow.AddWidget(a.system.QWidget)
	col.AddLayout(sysRow.QLayout)

	gameRow := qt6.NewQHBoxLayout2()
	a.game = qt6.NewQLineEdit2()
	a.game.SetPlaceholderText(a.tr.T("GAME_HINT_TYPE"))
	a.game.SetToolTip(a.tr.T("GAME_TITLE"))
	a.game.OnTextChanged(func(text string) {
		if a.silent {
			return
		}
		a.onGameTyped(text)
	})
	a.region = qt6.NewQComboBox2()
	a.region.AddItem3(a.tr.T("RENAME_DEFAULT"), qt6.NewQVariant11(""))
	a.region.AddItem3(a.tr.T("REGION_US"), qt6.NewQVariant11("US"))
	a.region.AddItem3(a.tr.T("REGION_EU"), qt6.NewQVariant11("EU"))
	a.region.AddItem3(a.tr.T("REGION_JP"), qt6.NewQVariant11("JP"))
	a.region.OnCurrentIndexChanged(func(index int) {
		if a.silent {
			return
		}
		g := a.gameState()
		g.Region, _ = nso.ParseRegion(a.region.ItemData(index).ToString())
		a.putGame(g)
		a.refreshGameUI()
		a.updateApply()
	})
	gameRow.AddWidget2(a.game.QWidget, 1)
	gameRow.AddWidget(a.region.QWidget)
	col.AddLayout(gameRow.QLayout)

	descRow := qt6.NewQHBoxLayout2()
	pub := qt6.NewQLabel2()
	pub.SetPixmap(iconPublic().Pixmap2(16, 16))
	a.desc = qt6.NewQComboBox2()
	a.desc.SetToolTip(a.tr.T("SHORT_DESC"))
	a.custom = qt6.NewQLineEdit2()
	a.custom.SetPlaceholderText(a.tr.T("DESCRIPTION_TITLE"))
	a.custom.SetVisible(false)
	a.desc.OnCurrentIndexChanged(func(index int) {
		if a.silent {
			return
		}
		g := a.gameState()
		g.Mode = a.desc.ItemData(index).ToString()
		a.putGame(g)
		a.custom.SetVisible(g.Mode == "custom")
		a.updateApply()
	})
	a.custom.OnTextChanged(func(text string) {
		if a.silent {
			return
		}
		g := a.gameState()
		g.Description = text
		a.putGame(g)
		a.updateApply()
	})
	descRow.AddWidget(pub.QWidget)
	descRow.AddWidget2(a.desc.QWidget, 1)
	descRow.AddWidget2(a.custom.QWidget, 1)
	col.AddLayout(descRow.QLayout)

	partyRow := qt6.NewQHBoxLayout2()
	a.partyOn = qt6.NewQCheckBox2()
	a.noParty = qt6.NewQLabel3(a.tr.T("NO_PARTY"))
	a.partyBox = qt6.NewQWidget2()
	pl := qt6.NewQHBoxLayout(a.partyBox)
	pl.SetContentsMargins(0, 0, 0, 0)
	a.partySize = qt6.NewQSpinBox2()
	a.partySize.SetRange(1, 99)
	of := qt6.NewQLabel3(a.tr.T("PARTY_OF"))
	a.partyMax = qt6.NewQSpinBox2()
	a.partyMax.SetRange(1, 99)
	pl.AddWidget(a.partySize.QWidget)
	pl.AddWidget(of.QWidget)
	pl.AddWidget(a.partyMax.QWidget)
	a.partyOn.OnToggled(func(on bool) {
		if a.silent {
			return
		}
		g := a.gameState()
		g.Party = on
		a.putGame(g)
		a.noParty.SetVisible(!on)
		a.partyBox.SetVisible(on)
		a.updateApply()
	})
	a.partySize.OnValueChanged(func(v int) {
		if a.silent {
			return
		}
		g := a.gameState()
		g.PartySize = v
		if g.PartyMax < v {
			g.PartyMax = v
			a.silent = true
			a.partyMax.SetValue(v)
			a.silent = false
		}
		a.putGame(g)
		a.updateApply()
	})
	a.partyMax.OnValueChanged(func(v int) {
		if a.silent {
			return
		}
		g := a.gameState()
		g.PartyMax = v
		if g.PartySize > v {
			g.PartySize = v
			a.silent = true
			a.partySize.SetValue(v)
			a.silent = false
		}
		a.putGame(g)
		a.updateApply()
	})
	partyRow.AddWidget(a.partyOn.QWidget)
	partyRow.AddWidget(a.noParty.QWidget)
	partyRow.AddWidget(a.partyBox)
	partyRow.AddStretch()
	col.AddLayout(partyRow.QLayout)

	timeRow := qt6.NewQHBoxLayout2()
	gicon := qt6.NewQLabel2()
	gicon.SetPixmap(iconGames().Pixmap2(16, 16))
	a.elapsed = qt6.NewQLabel3("0:00")
	timeRow.AddWidget(gicon.QWidget)
	timeRow.AddWidget(a.elapsed.QWidget)
	timeRow.AddStretch()
	col.AddLayout(timeRow.QLayout)

	lay.AddLayout2(col.QLayout, 1)
	return card.QWidget
}

func (a *App) buildFriendCode(box *qt6.QVBoxLayout) {
	a.fcRow = qt6.NewQWidget2()
	row := qt6.NewQHBoxLayout(a.fcRow)
	row.SetContentsMargins(0, 0, 0, 0)
	row.AddWidget(qt6.NewQLabel3(a.tr.T("TAG_TITLE_FCID")).QWidget)
	row.AddStretch()
	a.fcPrefix = qt6.NewQLabel3("SW-")
	row.AddWidget(a.fcPrefix.QWidget)
	a.fcA = digitBox()
	a.fcB = digitBox()
	a.fcC = digitBox()
	row.AddWidget(a.fcA.QWidget)
	row.AddWidget(qt6.NewQLabel3("-").QWidget)
	row.AddWidget(a.fcB.QWidget)
	row.AddWidget(qt6.NewQLabel3("-").QWidget)
	row.AddWidget(a.fcC.QWidget)
	a.fcA.OnTextChanged(func(text string) { a.onFCChanged(a.fcA, a.fcB, text) })
	a.fcB.OnTextChanged(func(text string) { a.onFCChanged(a.fcB, a.fcC, text) })
	a.fcC.OnTextChanged(func(text string) { a.onFCChanged(a.fcC, nil, text) })
	box.AddWidget(a.fcRow)

	a.nnidRow = qt6.NewQWidget2()
	nr := qt6.NewQHBoxLayout(a.nnidRow)
	nr.SetContentsMargins(0, 0, 0, 0)
	nr.AddWidget(qt6.NewQLabel3(a.tr.T("TAG_TITLE_NNID")).QWidget)
	nr.AddStretch()
	a.nnid = qt6.NewQLineEdit2()
	a.nnid.SetMaximumWidth(180)
	a.nnid.OnTextChanged(func(text string) {
		if a.silent {
			return
		}
		a.sys().TagID = text
		a.fillDescOptions()
		a.updateApply()
	})
	nr.AddWidget(a.nnid.QWidget)
	box.AddWidget(a.nnidRow)
}

func (a *App) buildToggles(box *qt6.QVBoxLayout) {
	iconRow := qt6.NewQHBoxLayout2()
	iconRow.AddWidget(qt6.NewQLabel3(a.tr.T("TAG_ICON_TITLE")).QWidget)
	iconRow.AddStretch()
	iconRow.AddWidget(helpButton(a.tr.T("TAG_ICON_HINT")).QWidget)
	a.tagIcon = qt6.NewQCheckBox2()
	a.tagIcon.OnToggled(func(on bool) {
		if a.silent {
			return
		}
		a.sys().TagIcon = on
		a.updateApply()
	})
	iconRow.AddWidget(a.tagIcon.QWidget)
	box.AddLayout(iconRow.QLayout)

	presRow := qt6.NewQHBoxLayout2()
	presRow.AddWidget(qt6.NewQLabel3(a.tr.T("PRESERVE_TIME_TITLE")).QWidget)
	presRow.AddStretch()
	presRow.AddWidget(helpButton(a.tr.T("PRESERVE_TIME_HINT")).QWidget)
	a.preserve = qt6.NewQCheckBox2()
	a.preserve.OnToggled(func(on bool) {
		if a.silent {
			return
		}
		a.sys().TimePreserve = on
	})
	presRow.AddWidget(a.preserve.QWidget)
	box.AddLayout(presRow.QLayout)
}

func (a *App) buildBar() *qt6.QWidget {
	bar := qt6.NewQWidget2()
	row := qt6.NewQHBoxLayout(bar)
	row.SetContentsMargins(0, 8, 0, 0)

	a.avatar = qt6.NewQLabel2()
	a.avatar.SetFixedSize2(32, 32)
	a.avatar.SetScaledContents(true)
	userCol := qt6.NewQVBoxLayout2()
	userCol.SetSpacing(0)
	a.userName = qt6.NewQLabel3("Discord")
	a.userStatus = qt6.NewQLabel3(a.tr.T("USER_DISCONNECTED"))
	userCol.AddWidget(a.userName.QWidget)
	userCol.AddWidget(a.userStatus.QWidget)
	row.AddWidget(a.avatar.QWidget)
	row.AddLayout(userCol.QLayout)
	row.AddStretch()

	a.applyBtn = qt6.NewQPushButton2()
	a.applyBtn.OnClicked(func() { a.onApply() })
	a.timerBtn = qt6.NewQPushButton2()
	a.timerBtn.SetIcon(iconNamed("chronometer", "chronometer"))
	a.timerBtn.SetToolTip(a.tr.T("TIMER_TITLE"))
	a.timerBtn.OnClicked(func() { a.onTimer() })
	a.visBtn = qt6.NewQPushButton2()
	a.visBtn.SetCheckable(true)
	a.visBtn.OnClicked(func() { a.onVisibility() })
	a.cfgBtn = qt6.NewQPushButton2()
	a.cfgBtn.SetIcon(iconNamed("configure", "configure"))
	a.cfgBtn.SetToolTip(a.tr.T("CONFIGURE"))
	a.cfgBtn.OnClicked(func() { a.stack.SetCurrentIndex(1) })
	row.AddWidget(a.applyBtn.QWidget)
	row.AddWidget(a.timerBtn.QWidget)
	row.AddWidget(a.visBtn.QWidget)
	row.AddWidget(a.cfgBtn.QWidget)
	return bar
}

func digitBox() *qt6.QLineEdit {
	e := qt6.NewQLineEdit2()
	e.SetMaxLength(4)
	e.SetMaximumWidth(56)
	e.SetAlignment(qt6.AlignHCenter)
	e.SetPlaceholderText("0000")
	return e
}

func digitsOnly(s string, n int) string {
	out := make([]rune, 0, n)
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out = append(out, r)
			if len(out) >= n {
				break
			}
		}
	}
	return string(out)
}

func (a *App) onFCChanged(cur, next *qt6.QLineEdit, text string) {
	cleaned := digitsOnly(text, 4)
	if cleaned != text {
		pos := cur.CursorPosition()
		cur.SetText(cleaned)
		if pos > len(cleaned) {
			pos = len(cleaned)
		}
		cur.SetCursorPosition(pos)
		return
	}
	if next != nil && len(text) >= 4 {
		next.SetFocus()
		next.SelectAll()
	}
	if a.silent {
		return
	}
	a.sys().TagFC = [3]string{a.fcA.Text(), a.fcB.Text(), a.fcC.Text()}
	a.fillDescOptions()
	a.updateApply()
}

func (a *App) selectComboData(c *qt6.QComboBox, data string) {
	for i := 0; i < c.Count(); i++ {
		if c.ItemData(i).ToString() == data {
			c.SetCurrentIndex(i)
			return
		}
	}
}

func (a *App) fillDescOptions() {
	a.silent = true
	defer func() { a.silent = false }()
	a.desc.Clear()
	tag := a.tag()
	fcLabel := a.tr.T("TAG_TITLE_FCID")
	if tag != "" {
		fcLabel = tag
	}
	a.desc.AddItem3(fcLabel, qt6.NewQVariant11("friendcode"))
	a.desc.AddItem3(a.tr.T("SHORT_DESC_CUSTOM"), qt6.NewQVariant11("custom"))
	a.desc.AddItem3(a.tr.T("SHORT_DESC_EMPTY"), qt6.NewQVariant11("empty"))
	a.selectComboData(a.desc, a.gameState().Mode)
}

func (a *App) refillCompleter() {
	titles := []string{}
	for _, g := range a.nso.Games(a.settings.System) {
		titles = append(titles, g.Title(a.settings.Region))
	}
	comp := qt6.NewQCompleter6(titles, a.game.QObject)
	comp.SetCaseSensitivity(qt6.CaseInsensitive)
	comp.SetFilterMode(qt6.MatchContains)
	comp.SetMaxVisibleItems(8)
	comp.OnActivated(func(text string) {
		hits := nso.Search(a.nso.Games(a.settings.System), text, a.settings.Region)
		if len(hits) == 0 {
			return
		}
		for _, h := range hits {
			if h.DisplayTitle == text || h.Exact {
				a.setGameID(h.Game.ID)
				return
			}
		}
		a.setGameID(hits[0].Game.ID)
	})
	a.game.SetCompleter(comp)
}

func (a *App) onGameTyped(text string) {
	prev := a.sys().Game
	g, ok := nso.Resolve(a.nso.Games(a.settings.System), text, a.settings.Region)
	id := text
	if ok {
		id = g.ID
	}
	a.sys().Game = id
	if prev != id {
		a.bumpElapsed(true)
	}
	a.loadCover()
	a.updateApply()
}

func (a *App) setGameID(id string) {
	prev := a.sys().Game
	a.sys().Game = id
	if prev != id {
		a.bumpElapsed(true)
	}
	a.refreshGameUI()
	a.updateApply()
}

func (a *App) reloadSystem() {
	a.silent = true
	a.selectComboData(a.system, string(a.settings.System))
	st := a.sys()
	wiiu := a.settings.System == nso.WUP
	a.nnidRow.SetVisible(wiiu)
	a.fcRow.SetVisible(!wiiu)
	a.fcPrefix.SetVisible(a.settings.System == nso.HAC || a.settings.System == nso.BEE)
	a.nnid.SetText(st.TagID)
	a.fcA.SetText(st.TagFC[0])
	a.fcB.SetText(st.TagFC[1])
	a.fcC.SetText(st.TagFC[2])
	a.tagIcon.SetChecked(st.TagIcon)
	a.preserve.SetChecked(st.TimePreserve)
	a.silent = false
	a.refillCompleter()
	a.refreshGameUI()
	a.updateApply()
	a.updateElapsed()
}

func (a *App) refreshGameUI() {
	a.silent = true
	st := a.sys()
	g := a.gameState()
	if game, ok := a.currentGame(); ok {
		a.game.SetText(game.Title(a.preferredRegion()))
		a.region.SetEnabled(true)
	} else {
		a.game.SetText(st.Game)
		a.region.SetEnabled(false)
	}
	a.selectComboData(a.region, string(g.Region))
	a.fillDescOptions()
	a.custom.SetText(g.Description)
	a.custom.SetVisible(g.Mode == "custom")
	a.partyOn.SetChecked(g.Party)
	a.noParty.SetVisible(!g.Party)
	a.partyBox.SetVisible(g.Party)
	a.partySize.SetValue(g.PartySize)
	a.partyMax.SetValue(g.PartyMax)
	a.silent = false
	a.loadCover()
}
