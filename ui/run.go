package ui

import (
	"context"
	"os"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/voxelprismatic/richpresenceu/discord"
	"github.com/voxelprismatic/richpresenceu/nso"
)

func Main() {
	qt6.NewQApplication(os.Args)
	qt6.QCoreApplication_SetApplicationName("Rich Presence U")
	qt6.QCoreApplication_SetApplicationVersion(Version)
	qt6.QCoreApplication_SetOrganizationName("VoxelPrismatic")

	client, err := nso.New("", "")
	if err != nil {
		panic(err)
	}
	_ = client.LoadCache()

	settings, systems := loadPrefs(client.ConfigDir)
	a := &App{
		tr:       newI18n(),
		nso:      client,
		rpc:      discord.New(),
		settings: settings,
		systems:  systems,
		log:      logger{dir: client.ConfigDir},
	}
	a.log.SetEnabled(a.settings.DebugLog)
	a.tr.Set(a.settings.Language)

	a.rpc.OnClose(func(err error) {
		mainthread.Start(func() {
			if err != nil {
				a.debug("discord closed: %v", err)
			}
			a.onDisconnected()
		})
	})

	a.buildWindow()
	a.loadSettingsIntoUI()
	a.reloadSystem()
	a.updateApply()
	a.bumpElapsed(false)

	a.tick = qt6.NewQTimer2(a.win.QObject)
	a.tick.SetInterval(1000)
	a.tick.OnTimeout(func() { a.updateElapsed() })
	a.tick.Start2()

	a.hide = qt6.NewQTimer2(a.win.QObject)
	a.hide.SetSingleShot(true)
	a.hide.OnTimeout(func() {
		a.settings.Activity = false
		if a.rpc.Connected() {
			go func() { _ = a.rpc.Clear(context.Background()) }()
		}
		a.updateApply()
		a.updateElapsed()
		a.refreshScreensaver()
	})

	a.win.Show()
	a.refreshTitles(false)
	if a.settings.AutoConnect {
		a.connect(false)
	}
	os.Exit(qt6.QApplication_Exec())
}
