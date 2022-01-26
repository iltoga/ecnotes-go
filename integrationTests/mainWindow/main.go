package main

import (
	"fyne.io/fyne/app"
	"github.com/iltoga/ecnotes-go/ui"
)

// main function to be called from the test
func main() {
	// create a new ui
	testUI := ui.NewUI(app.New())
	testUI.CreateMainWindow()
}
