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
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "NFA Tool Recode v2",
		Width:            980,
		Height:           640,
		MinWidth:         980,
		MinHeight:        640,
		MaxWidth:         980,
		MaxHeight:        640,
		Frameless:        true,
		BackgroundColour: application.NewRGB(15, 15, 26),
		URL:              "/",
		DisableResize:    true,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
