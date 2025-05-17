package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

//go:embed content/nerd-fonts/icons.json
var embeddedIcons []byte

func main() {
	// Create an instance of the app structure
	app := NewApp() // This NewApp() comes from app.go

	// Create application with options
	err := wails.Run(&options.App{
		Title:                    "uniGO",
		Width:                    768,
		Height:                   1024,
		MinWidth:                 600,
		MinHeight:                800,
		Fullscreen:               false,
		CSSDragProperty:          "drag",
		EnableDefaultContextMenu: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
			// Handler: iconLoader(),
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup, // Use the startup method from app.go
		Bind: []interface{}{
			app, // Bind the entire app struct
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:  true,
			CSSDropProperty: "dropper",
			CSSDropValue:    "drop",
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "uniGO",
				Message: "MIT",
				Icon:    appIcon,
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			// BackdropType:                      windows.Acrylic,
			DisablePinchZoom:                  true,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
