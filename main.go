package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ensureAdmin()

	if os.Getenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS") == "" {
		_ = os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--disable-gpu")
	}

	app := application.New(application.Options{
		Name:        fmt.Sprintf("%s v%s", AppName, AppVersion),
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
		Title:            fmt.Sprintf("%s v%s", AppName, AppVersion),
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
