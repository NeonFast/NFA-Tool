package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ensureAdmin()

	app := application.New(application.Options{
		Name:        "NFA Tool Recode v2",
		Description: "Steam ConnectCache token login",
		Services: []application.Service{
			application.NewService(NewAppService()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// Keep running if guide window is closed while main is open (and vice versa on mac).
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "NFA Tool Recode v2",
		Width:            1120,
		Height:           780,
		MinWidth:         1120,
		MinHeight:        780,
		MaxWidth:         1120,
		MaxHeight:        780,
		Frameless:        true,
		BackgroundColour: application.NewRGB(15, 15, 26),
		URL:              "/",
		DisableResize:    true,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
