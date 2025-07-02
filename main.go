package main

import (
	"embed"
	// "io/ioutil" // deprecated, replaced by os
	"log"
	"os"                // added for os.readfile
	goruntime "runtime" // alias for standard library runtime

	// required for runtime.opendirectorydialog wrapper if used
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu" // import menu package
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	// alias for wails runtime package
)

//go:embed all:frontend/dist
var assets embed.FS

// add this selectdirectory method to the app struct in app.go
// or, if you prefer to keep app.go clean of wails runtime imports for some reason,
// you can define it here and pass `app.ctx` if needed, though it's better in app.go.
// for this example, let's assume it's added to app.go and `app.ctx` is used.
// example of how it would look in app.go:
/*
func (a *app) selectdirectory() (string, error) {
	selecteddirectory, err := wailsruntime.opendirectorydialog(a.ctx, wailsruntime.opendialogoptions{
		title: "select project directory",
	})
	if err != nil {
		return "", err
	}
	return selecteddirectory, nil
}
*/

func main() {
	app := NewApp() // creates an instance of app from app.go
	// load icons

	iconPNG, errPNG := os.ReadFile("appicon.png") // changed from ioutil.readfile
	if errPNG != nil {
		log.Println("warning: could not load appicon.png:", errPNG)
	}

	appMenu := menu.NewMenu() // create an empty menu

	if goruntime.GOOS == "darwin" { // check if os is macos
		appMenu.Append(menu.AppMenu())  // add standard appmenu (quit, about, hide, etc.)
		appMenu.Append(menu.EditMenu()) // add standard editmenu (copy, paste, cut, select all)
	}

	err := wails.Run(&options.App{
		Title:  "shotgun",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, // this binds all public methods of app
		},
		Menu: appMenu, // set the application menu

		Linux: &linux.Options{
			Icon: iconPNG,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
