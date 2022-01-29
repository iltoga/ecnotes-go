package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/service"
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
	confSrv service.ConfigService
}

// NewUI UI constructor
func NewUI(
	app fyne.App,
	confSrv service.ConfigService,
) *UImpl {
	return &UImpl{
		app:     app,
		windows: make(map[string]fyne.Window),
		mux:     &sync.Mutex{},
		confSrv: confSrv,
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

	// if we have encryption key, show password entry to decrypt and save to global map
	if _, err := ui.confSrv.GetConfig("common.CONFIG_ENCRYPTION_KEY"); err == nil {
		w.SetContent(container.NewVBox(
			widget.NewLabel("Decrypting..."),
			ui.runPasswordPopUp(w, common.EncryptionKeyAction_Decrypt),
		))
	} else {
		w.SetContent(container.NewVBox(
			widget.NewLabel("Generating encryption key..."),
			ui.runPasswordPopUp(w, common.EncryptionKeyAction_Generate),
		))
	}

	// if we don't have encryption key, show password entry to generate it and save to config file
	// _ = ui.runPasswordPopUp(w)
	w.ShowAndRun()
}

func (ui *UImpl) runPasswordPopUp(w fyne.Window, keyAction common.EncryptionKeyAction) (modal *widget.PopUp) {
	var (
		encKey, decKey string
		err            error
		popUpText      = widget.NewLabel("Enter password")
		pwdWg          = widget.NewPasswordEntry()
		btnWg          = widget.NewButton("OK", func() {
			// generate encryption key, encrypt with password and save to file
			switch keyAction {
			case common.EncryptionKeyAction_Generate:
				// generate encryption key
				decKey, err = cryptoUtil.SecureRandomStr(common.EncryptionKeyLength)
				if err != nil {
					ui.showNotification("Error generating encryption key", err.Error())
					return
				}
				ui.confSrv.SetGlobal("common.CONFIG_ENCRYPTION_KEY", decKey)
				// encrypt the key with password input in the password entry
				if encKey, err = cryptoUtil.EncryptMessage(decKey, pwdWg.Text); err != nil {
					ui.showNotification("Error encrypting encryption key", err.Error())
					return
				}
				// save encrypted encryption key to config file
				ui.confSrv.SetConfig("common.CONFIG_ENCRYPTION_KEY", encKey)
				if err := ui.confSrv.SaveConfig(); err != nil {
					ui.showNotification("Error saving configuration", err.Error())
					return
				}
				ui.showNotification("Encryption key generated", "")
			case common.EncryptionKeyAction_Decrypt:
				// decrypt the key with password input in the password entry
				if encKey, err = ui.confSrv.GetConfig("common.CONFIG_ENCRYPTION_KEY"); err != nil {
					ui.showNotification("Error loading encryption key from app configuration", err.Error())
					return
				}
				if decKey, err = cryptoUtil.DecryptMessage(encKey, pwdWg.Text); err != nil {
					ui.showNotification("Error decrypting encryption key", err.Error())
					return
				}
				ui.confSrv.SetGlobal("common.CONFIG_ENCRYPTION_KEY", decKey)
				ui.showNotification("Encryption key decrypted and stored in memory till app is closed", "")
			default:
				ui.showNotification("Error", "Unknown key action")
			}

			modal.Hide()
			// reset password entry for security
			pwdWg.SetText("")
		})
	)

	if keyAction == common.EncryptionKeyAction_Decrypt {
		btnWg.SetText("Enter password to Decrypt Key")
	} else {
		btnWg.SetText("Enter password to Encrypt generated Key")
	}
	modal = widget.NewModalPopUp(
		widget.NewVBox(
			popUpText,
			pwdWg,
			btnWg,
		),
		w.Canvas(),
	)
	modal.Show()
	return modal
}

func (ui *UImpl) showNotification(title, contentStr string) {
	time.Sleep(time.Millisecond * 500)
	ui.app.SendNotification(fyne.NewNotification(title, contentStr))
}
