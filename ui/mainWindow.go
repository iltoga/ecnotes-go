package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

// UI ....
type UI interface {
	CreateMainWindow()
	AddWindow(name string, w fyne.Window)
	GetWindow(name string) (fyne.Window, error)
}

// UImpl Main ui configuration
type UImpl struct {
	app     fyne.App
	windows map[string]fyne.Window
	mux     *sync.Mutex
}

// NewUI UI constructor
func NewUI(
	app fyne.App,
) *UImpl {
	return &UImpl{
		app:     app,
		windows: make(map[string]fyne.Window),
		mux:     &sync.Mutex{},
	}
}

// AddWindow add window to map
func (ui *UImpl) AddWindow(name string, w fyne.Window) {
	ui.mux.Lock()
	ui.windows[name] = w
	ui.mux.Unlock()
}

// GetWindow ....
func GetWindow(ui *UImpl, name string) (fyne.Window, error) {
	if w, ok := ui.windows[name]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("window %s not found", name)
}

// CreateMainWindow ....
func (ui *UImpl) CreateMainWindow() {
	w := ui.app.NewWindow("EcNotes")
	ui.AddWindow("main", w)

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))
	w.Resize(fyne.NewSize(400, 600))

	_ = ui.runPasswordPopUp(w)

	w.ShowAndRun()
}

func (ui *UImpl) runPasswordPopUp(w fyne.Window) (modal *widget.PopUp) {
	var (
		pwdWg = widget.NewPasswordEntry()
		btnWg = widget.NewButton("OK", func() {
			// generate encryption key, encrypt with password and save to file

			modal.Hide()
			// reset password entry for security
			// STEF totest only!
			pwdWg.SetText("")
		})
	)
	modal = widget.NewModalPopUp(
		widget.NewVBox(
			widget.NewLabel("Generate Encryption Key"),
			pwdWg,
			btnWg,
		),
		w.Canvas(),
	)
	modal.Show()
	return modal
}

func showNotification(a fyne.App, title, contentStr string) {
	time.Sleep(time.Second * 2)
	a.SendNotification(fyne.NewNotification(title, contentStr))
}
