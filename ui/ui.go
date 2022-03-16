package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
)

// UI ....
type UI interface {
	AddWindow(name string, w fyne.Window)
	AddWidget(name string, w fyne.CanvasObject)
	GetWindow(name string) (fyne.Window, error)
	GetWidget(name string) (fyne.CanvasObject, error)
	ShowNotification(title, contentStr string)
	Run()
	Stop()
	SetFocusOnWidget(w fyne.Window, wg fyne.CanvasObject)
	GetNoteService() service.NoteService
	GetObserver() observer.Observer
	SetWidgetVisibility(name string, visible bool) error
	SetWidgetEnabled(name string, enabled bool) error
	SetWindowVisibility(name string, visible bool) error
	ToggleFullScreen(w fyne.Window)
}

// UImpl Main ui configuration
type UImpl struct {
	app         fyne.App
	windows     map[string]fyne.Window
	winMux      *sync.Mutex
	widgets     map[string]fyne.CanvasObject
	widMux      *sync.Mutex
	confService service.ConfigService
	certService service.CertService
	noteService service.NoteService
	obs         observer.Observer
}

// NewUI UI constructor
func NewUI(
	app fyne.App,
	confService service.ConfigService,
	noteService service.NoteService,
	certService service.CertService,
	obs observer.Observer,
) *UImpl {
	return &UImpl{
		app:         app,
		windows:     make(map[string]fyne.Window),
		widgets:     make(map[string]fyne.CanvasObject),
		winMux:      &sync.Mutex{},
		widMux:      &sync.Mutex{},
		confService: confService,
		noteService: noteService,
		certService: certService,
		obs:         obs,
	}
}

// SetWidgetEnabled ....
func (ui *UImpl) SetWidgetEnabled(name string, enabled bool) error {
	ui.widMux.Lock()
	defer ui.widMux.Unlock()
	if w, ok := ui.widgets[name]; ok {
		wd, ok := w.(fyne.Disableable)
		if !ok {
			return fmt.Errorf("widget %s is not disableable", name)
		}
		if enabled {
			wd.Enable()
		} else {
			wd.Disable()
		}
		return nil
	}
	return fmt.Errorf("widget %s not found", name)
}

// ToggleFullScreen ....
func (ui *UImpl) ToggleFullScreen(w fyne.Window) {
	w.SetFullScreen(!w.FullScreen())
}

// SetWidgetVisibility ....
func (ui *UImpl) SetWidgetVisibility(name string, visible bool) error {
	ui.widMux.Lock()
	defer ui.widMux.Unlock()
	if w, ok := ui.widgets[name]; ok {
		if visible {
			w.Show()
		} else {
			w.Hide()
		}
		return nil
	}
	return fmt.Errorf("widget %s not found", name)
}

// SetWindowVisibility ....
func (ui *UImpl) SetWindowVisibility(name string, visible bool) error {
	ui.winMux.Lock()
	defer ui.winMux.Unlock()
	if w, ok := ui.windows[name]; ok {
		if visible {
			w.Show()
		} else {
			w.Hide()
		}
		return nil
	}
	return fmt.Errorf("window %s not found", name)
}

// GetNoteService ....
func (ui *UImpl) GetNoteService() service.NoteService {
	return ui.noteService
}

// GetObserver ....
func (ui *UImpl) GetObserver() observer.Observer {
	return ui.obs
}

// Run ....
func (ui *UImpl) Run() {
	ui.app.Run()
}

// Stop ....
func (ui *UImpl) Stop() {
	ui.app.Quit()
}

// SetFocusOnWidget ....
func (ui *UImpl) SetFocusOnWidget(w fyne.Window, wg fyne.Focusable) {
	w.Canvas().Focus(wg)
}

// AddWindow add window to map
func (ui *UImpl) AddWindow(name string, w fyne.Window) {
	ui.winMux.Lock()
	ui.windows[name] = w
	ui.winMux.Unlock()
}

// AddWidget add widget to map
func (ui *UImpl) AddWidget(name string, w fyne.CanvasObject) {
	ui.widMux.Lock()
	ui.widgets[name] = w
	ui.widMux.Unlock()
}

// GetWindow ....
func (ui *UImpl) GetWindow(name string) (fyne.Window, error) {
	if w, ok := ui.windows[name]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("window %s not found", name)
}

// GetWidget ....
func (ui *UImpl) GetWidget(name string) (fyne.CanvasObject, error) {
	if w, ok := ui.widgets[name]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("widget %s not found", name)
}

// ShowNotification ....
func (ui *UImpl) ShowNotification(title, contentStr string) {
	time.Sleep(time.Millisecond * 500)
	ui.app.SendNotification(fyne.NewNotification(title, contentStr))
}
