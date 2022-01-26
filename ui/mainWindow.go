package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

// UI Main ui configuration
type UI struct {
	app fyne.App
}

// NewUI UI constructor
func NewUI(
	app fyne.App,
) *UI {
	return &UI{
		app: app,
	}
}

// CreateMainWindow ....
func (ui *UI) CreateMainWindow() {
	w := ui.app.NewWindow("EcNotes")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))
	w.Resize(fyne.NewSize(200, 200))

	_ = runPopUp(w)

	w.ShowAndRun()
}

func runPopUp(w fyne.Window) (modal *widget.PopUp) {
	modal = widget.NewModalPopUp(
		widget.NewVBox(
			widget.NewLabel("bar"),
			widget.NewButton("Close", func() { modal.Hide() }),
		),
		w.Canvas(),
	)
	modal.Show()
	return modal
}
