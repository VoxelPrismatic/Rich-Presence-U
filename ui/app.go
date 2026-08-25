package ui

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/discord"
	"github.com/voxelprismatic/richpresenceu/igdb"
	"github.com/voxelprismatic/richpresenceu/nso"
)

type App struct {
	tr       *i18n
	log      logger
	nso      *nso.Client
	igdbAPI  *igdb.Client
	rpc      *discord.Client
	settings Settings
	systems  map[string]*SystemState
	start    int64
	applied  string
	built    *discord.Activity
	busy     bool
	silent   bool
	warnHide bool
	inhibit  inhibitor

	win           *qt6.QMainWindow
	stack         *qt6.QStackedWidget
	cover         *qt6.QLabel
	system        *qt6.QComboBox
	platModel     *platformModel
	platPopup     *qt6.QWidget
	platList      *qt6.QListView
	platBack      *qt6.QToolButton
	platCrumb     *qt6.QLabel
	platSearch    *qt6.QLineEdit
	platHits      *qt6.QStandardItemModel
	platSearching bool
	platSavedRoot uintptr
	platHasSaved  bool
	game          *qt6.QComboBox
	region        *qt6.QComboBox
	infoBtn       *qt6.QPushButton
	desc          *qt6.QComboBox
	partyOn       *qt6.QCheckBox
	partyBox      *qt6.QWidget
	noParty       *qt6.QLabel
	partySize     *qt6.QSpinBox
	partyMax      *qt6.QSpinBox
	elapsed       *qt6.QLabel
	elapsedIcon   *qt6.QLabel
	timerEnabled  bool
	fcPrefix      *qt6.QLabel
	fcA, fcB, fcC *qt6.QLineEdit
	nnid          *qt6.QLineEdit
	fcRow         *qt6.QWidget
	nnidRow       *qt6.QWidget
	tagIcon       *qt6.QCheckBox
	preserve      *qt6.QCheckBox
	avatar        *qt6.QLabel
	userName      *qt6.QLabel
	userStatus    *qt6.QLabel
	applyBtn      *qt6.QPushButton
	timerBtn      *qt6.QPushButton
	visBtn        *qt6.QPushButton
	cfgBtn        *qt6.QPushButton

	langCombo  *qt6.QComboBox
	prefRegion *qt6.QComboBox
	autoConn   *qt6.QCheckBox
	keepOn     *qt6.QCheckBox
	debugOn    *qt6.QCheckBox
	igdbID     *qt6.QLineEdit
	igdbSecret *qt6.QLineEdit
	dataCombo  *qt6.QComboBox
	dataBtn    *qt6.QPushButton
	aboutTable *qt6.QTableWidget

	tick *qt6.QTimer
	hide *qt6.QTimer

	searchTimer   *qt6.QTimer
	searchHits    []nso.Game
	searchGen     int
	gameHighlight string
}

func (a *App) sysKey() string {
	if a.settings.Platform != "" {
		return a.settings.Platform
	}
	if slug := igdb.SlugForStoreCode(string(a.settings.System)); slug != "" {
		return slug
	}
	return "NS1"
}

func (a *App) sys() *SystemState {
	key := a.sysKey()
	st := a.systems[key]
	if st == nil {
		d := defaultSystem()
		st = &d
		a.systems[key] = st
	}
	return st
}

func (a *App) gameState() GameState {
	st := a.sys()
	g, ok := st.Library[st.Game]
	if !ok {
		g = defaultGame()
	}
	if g.Mode == "" {
		g.Mode = "empty"
	}
	if g.PartySize < 1 {
		g.PartySize = 1
	}
	if g.PartyMax < 1 {
		g.PartyMax = 1
	}
	return g
}

func (a *App) putGame(g GameState) {
	st := a.sys()
	if st.Library == nil {
		st.Library = map[string]GameState{}
	}
	if st.Game != "" {
		st.Library[st.Game] = g
	}
}

func (a *App) preferredRegion() nso.Region {
	g := a.gameState()
	if g.Region.Valid() {
		return g.Region
	}
	return a.settings.Region
}

func (a *App) currentGame() (nso.Game, bool) {
	st := a.sys()
	if st.Game == "" {
		return nso.Game{}, false
	}
	sys, ok := a.catalogSystem()
	if !ok {
		return nso.Game{}, false
	}
	return a.nso.Lookup(sys, st.Game)
}

func (a *App) title() string {
	g, ok := a.currentGame()
	if ok {
		return g.Title(a.preferredRegion())
	}
	return a.sys().Game
}

func (a *App) tag() string {
	st := a.sys()
	sys := a.settings.System
	if nsoSys, ok := a.nsoSystem(); ok {
		sys = nsoSys
	}
	log.Println("tag", sys, st.TagID, st.TagFC)
	return discord.FormatTag(string(sys), st.TagID, st.TagFC)
}

func (a *App) presence() discord.Presence {
	gstate := a.gameState()
	st := a.sys()
	p := discord.Presence{
		Title:     a.title(),
		Console:   a.platformDisplay(),
		ShowTag:   true,
		TagIcon:   st.TagIcon,
		Tag:       a.tag(),
		Party:     gstate.Party,
		PartySize: gstate.PartySize,
		PartyMax:  gstate.PartyMax,
		Start:     a.start,
	}
	switch gstate.Mode {
	case "custom":
		p.Description = gstate.Description
	case "friendcode":
		p.Description = ""
		p.TagIcon = false
		p.ShowTag = true
	case "empty":
		p.Description = ""
		p.ShowTag = st.TagIcon
	}
	if game, ok := a.currentGame(); ok {
		region := a.preferredRegion()
		p.CoverKey = a.nso.CoverKey(game, region)
	} else {
		p.CoverKey = "default"
	}
	if label, u := a.statusButton(); u != "" {
		p.Buttons = []discord.Button{{Label: label, URL: u}}
	}
	return p
}

func (a *App) statusButton() (label, u string) {
	if _, store := a.nsoSystem(); store {
		if game, ok := a.currentGame(); ok {
			if page := game.Store(a.preferredRegion()); page != "" {
				return a.tr.T("ESHOP_BUTTON"), page
			}
		}
		return "", ""
	}
	if strings.TrimSpace(a.title()) == "" {
		return "", ""
	}
	return a.tr.T("SEARCH_BUTTON_TEXT"), a.gameInfoURL()
}

func (a *App) presenceForPush() discord.Presence {
	p := a.presence()
	if rem := a.timerRemaining(); rem > 0 {
		p.End = time.Now().Unix() + int64(rem)
		p.Start = 0
	} else if a.timerEnabled && a.settings.Timer > 0 && a.settings.Activity {
		p.End = time.Now().Unix() + int64(a.settings.Timer)
		p.Start = 0
	}
	return p
}

func (a *App) fingerprint() string {
	p := a.presence()
	st := a.sys()
	raw, _ := json.Marshal([]any{a.settings.Platform, a.settings.System, st.Game, a.preferredRegion(), p, st.TagFC, st.TagID, st.TagIcon, a.gameState()})
	sum := sha1.Sum(raw)
	return hex.EncodeToString(sum[:])
}

func (a *App) persist() {
	_ = savePrefs(a.nso.ConfigDir, a.settings, a.systems)
}

func (a *App) rememberGame() {
	st := a.sys()
	if st.Game == "" {
		return
	}
	hist := make([]string, 0, len(st.History)+1)
	for _, id := range st.History {
		if id != st.Game {
			hist = append(hist, id)
		}
	}
	hist = append(hist, st.Game)
	if len(hist) > 8 {
		hist = hist[len(hist)-8:]
	}
	st.History = hist
}

func (a *App) bumpElapsed(gameChanged bool) {
	st := a.sys()
	if a.start == 0 || (gameChanged && !st.TimePreserve) {
		a.start = time.Now().Unix()
	}
}

func (a *App) debug(format string, args ...any) {
	a.log.Print(format, args...)
}

func (a *App) refreshScreensaver() {
	on := a.settings.KeepOn && a.settings.Activity && a.rpc.Connected() && a.built != nil
	a.inhibit.Set(on)
}

func (a *App) loadCover() {
	game, ok := a.currentGame()
	if !ok {
		a.cover.SetPixmap(qt6.NewQPixmap())
		return
	}
	id := game.ID
	region := a.preferredRegion()
	go func() {
		path, err := a.nso.CoverPath(context.Background(), game, region)
		mainthread.Start(func() {
			cur, _ := a.currentGame()
			if cur.ID != id {
				return
			}
			if err != nil {
				a.cover.SetPixmap(qt6.NewQPixmap())
				return
			}
			pix := qt6.NewQPixmap4(path)
			if pix.IsNull() {
				return
			}
			a.cover.SetPixmap(maskPixmap(pix, coverSize, 8))
		})
	}()
}
