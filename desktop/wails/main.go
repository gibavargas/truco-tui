//go:build wails

package main

import (
	"embed"
	"io/fs"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	app := NewApp()
	frontendAssets := mustSubFS(assets, "frontend/dist")

	err := wails.Run(&options.App{
		Title:            "Truco Paulista",
		Width:            1520,
		Height:           980,
		MinWidth:         1120,
		MinHeight:        760,
		BackgroundColour: options.NewRGBA(10, 16, 18, 255),
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		println("error:", err.Error())
	}
}

func mustSubFS(root fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(root, dir)
	if err != nil {
		panic("wails frontend assets unavailable: " + err.Error())
	}
	return sub
}
