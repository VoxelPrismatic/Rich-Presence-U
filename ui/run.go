package ui

import (
	"os"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/discord"
	"github.com/voxelprismatic/richpresenceu/igdb"
	"github.com/voxelprismatic/richpresenceu/nso"
	"github.com/voxelprismatic/richpresenceu/svc"
)

func Main() {
	qt6.NewQApplication(os.Args)
	qt6.QCoreApplication_SetApplicationVersion(svc.VERSION)
	qt6.QCoreApplication_SetOrganizationName("VoxelPrismatic")
	nso.UserAgent = svc.UserAgent()

	client, err := nso.New("", "")
	if err != nil {
		panic(err)
	}
	_ = client.LoadCache()
	if b, err := os.ReadFile(svc.LogoPath(client.ConfigDir)); err == nil {
		applyAppIcon(b)
	}

	settings, systems := loadPrefs(client.ConfigDir)
	igdbc := igdb.NewClient()
	igdbc.UserAgent = svc.UserAgent()
	igdbc.SetCredentials(settings.IGDBClientID, settings.IGDBClientSecret)
	a := &App{
		tr:       newI18n(),
		nso:      client,
		igdbAPI:  igdbc,
		rpc:      discord.New(),
		settings: settings,
		systems:  systems,
		log:      logger{dir: client.ConfigDir},
	}
	a.log.SetEnabled(a.settings.DebugLog)
	a.tr.Set(a.settings.Language)
	qt6.QCoreApplication_SetApplicationName(a.tr.T("APP_TITLE"))
	// Krohnkite floats windows whose class is wl.float. On Wayland this is the app id.
	qt6.QGuiApplication_SetDesktopFileName("wl.float")

	a.rpc.OnClose(func(err error) {
		mainthread.Start(func() {
			if err != nil {
				a.debug("discord closed: %v", err)
			}
			a.onDisconnected()
		})
	})

	a.buildWindow()
	if a.win != nil {
		a.win.SetWindowIcon(qt6.QGuiApplication_WindowIcon())
	}
	a.maybeUpdate()
	a.maybeInstall()
	a.loadSettingsIntoUI()
	a.reloadSystem()
	a.updateApply()

	a.initClockTimer()

	a.win.Show()
	if a.settings.AutoConnect {
		a.connect(false)
	}
	os.Exit(qt6.QApplication_Exec())
}
