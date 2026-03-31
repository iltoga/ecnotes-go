// Package ui contains only widget construction, layout, and event wiring.
//
// # Architecture contract — READ BEFORE EDITING
//
// This file is intentionally THIN. The rules are:
//
//  1. No business logic here. Every non-trivial operation (crypto, config
//     mutations, file I/O) must live in a service (KeyService, NoteService, …)
//     or a lib/util. Call those from callbacks; never inline the logic.
//
//  2. Widget callbacks are allowed to:
//     - Read widget state (text, selected value, checkbox, …)
//     - Call a service method
//     - Call ShowNotification / Hide / Show / ch<-true
//     Nothing else.
//
//  3. If you feel the urge to import "encoding/hex", "crypto/…", or call
//     certService / confService directly from this file: stop. Add or extend a
//     KeyService method instead, then call that.
//
//  4. keyService (ui.keyService) owns ALL encryption-key lifecycle operations.
//     Never duplicate that logic here.
package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
)

type MainWindow interface {
	WindowInterface
	UpdateNoteListWidget() observer.Listener
}

type MainWindowImpl struct {
	UImpl
	WindowDefaultOptions
	titlesDataBinding binding.ExternalStringList
	selectedNote      *model.Note
	selectedNoteID    int
	w                 fyne.Window
	cryptoService     service.CryptoServiceFactory
}

func NewMainWindow(
	ui *UImpl,
	cryptoService service.CryptoServiceFactory,
) MainWindow {
	return &MainWindowImpl{
		UImpl:         *ui,
		cryptoService: cryptoService,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// noButtonDialog — a dialog.Dialog adapter backed by widget.ModalPopUp.
//
// Fyne v2.3.1's dialog.NewCustom always allocates a dismiss button, even when
// the dismiss text is "". That renders as a small phantom rectangle (visual
// artifact). This adapter uses widget.NewModalPopUp directly, which gives us
// full control over the content with zero built-in buttons.
// ──────────────────────────────────────────────────────────────────────────────

type noButtonDialog struct {
	popup    *widget.PopUp
	onClosed func()
}

func (d *noButtonDialog) Show()                   { d.popup.Show() }
func (d *noButtonDialog) Hide()                   { d.popup.Hide(); if d.onClosed != nil { d.onClosed() } }
func (d *noButtonDialog) Refresh()                { d.popup.Refresh() }
func (d *noButtonDialog) Resize(s fyne.Size)      { d.popup.Resize(s) }
func (d *noButtonDialog) MinSize() fyne.Size      { return d.popup.MinSize() }
func (d *noButtonDialog) SetDismissText(_ string) {} // no button — intentionally a no-op
func (d *noButtonDialog) SetOnClosed(fn func())   { d.onClosed = fn }

// newModalNoCancel builds a modal popup with a title label and arbitrary content
// but without any dismiss button. Returns the raw *widget.PopUp (for resize) and
// a hide closure that can be stored and called from callbacks.
func (ui *MainWindowImpl) newModalNoCancel(title string, content fyne.CanvasObject) *widget.PopUp {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	wrapped := container.NewVBox(
		container.NewPadded(titleLabel),
		widget.NewSeparator(),
		container.NewPadded(content),
	)
	// NewModalPopUp blocks interaction without any built-in button bar — no phantom button.
	popup := widget.NewModalPopUp(wrapped, ui.w.Canvas())
	return popup
}

// GetWindow returns the window object
func (ui *MainWindowImpl) GetWindow() fyne.Window {
	return ui.w
}

// getCurEncryptionKeyName returns the name of the configured default key.
// This is the only config read allowed in the UI layer; it is a pure lookup
// used only for display / routing purposes, not for crypto operations.
func (ui *MainWindowImpl) getCurEncryptionKeyName() (string, error) {
	return ui.confService.GetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME)
}

// createWindowContainer builds the main layout skeleton (search, buttons, separator).
func (ui *MainWindowImpl) createWindowContainer() *fyne.Container {
	var winLoaderText string
	if _, err := ui.getCurEncryptionKeyName(); err == nil {
		winLoaderText = "Decrypting encryption key..."
	} else {
		winLoaderText = "Generating encryption key..."
	}

	searchBox := widget.NewEntry()
	searchBox.SetPlaceHolder("Search note titles")
	ui.AddWidget(common.WDG_SEARCH_BOX, searchBox)
	searchBox.OnChanged = func(text string) {
		if _, err := ui.noteService.SearchNotes(text, true); err != nil {
			ui.ShowNotification("Error searching notes", err.Error())
		}
	}

	newNoteBtn := widget.NewButton("New", func() {
		ui.GetObserver().
			Notify(observer.EVENT_CREATE_NOTE_WINDOW, new(model.Note), common.WindowMode_Edit, common.WindowAction_New)
		if err := ui.SetWindowVisibility(common.WIN_NOTE_DETAILS, true); err != nil {
			ui.ShowNotification("Error", err.Error())
		}
	})
	hideBtn := widget.NewButton("Hide", func() {
		if ui.selectedNote != nil {
			ui.selectedNote.Hidden = true
			unEncContent := ui.selectedNote.Content
			if err := ui.noteService.UpdateNoteContent(ui.selectedNote); err != nil {
				ui.ShowNotification("Error updating note content", err.Error())
			}
			ui.selectedNote.Content = unEncContent
		}
	})
	deleteNoteBtn := widget.NewButton("Delete", func() {
		if ui.selectedNoteID != 0 {
			if err := ui.noteService.DeleteNote(ui.selectedNoteID); err != nil {
				ui.ShowNotification("Error deleting note", err.Error())
				return
			}
			ui.ShowNotification("Note Deleted", "")
		}
	})

	btnBar := container.New(layout.NewHBoxLayout(), newNoteBtn, hideBtn, deleteNoteBtn)
	btnContainer := container.New(
		layout.NewBorderLayout(nil, nil, nil, btnBar),
		btnBar,
	)

	mainWinLoaderLabel := widget.NewLabel("Loading main window...")
	mainWinLoaderLabel.Hidden = true
	mainWinLoader := func(msg string) *widget.Label {
		mainWinLoaderLabel.SetText(msg)
		mainWinLoaderLabel.Alignment = fyne.TextAlignCenter
		return mainWinLoaderLabel
	}

	return container.NewVBox(
		searchBox,
		btnContainer,
		widget.NewSeparator(),
		mainWinLoader(winLoaderText),
	)
}

// createPasswordPopUp decides whether to auto-load a passwordless key or show a
// dialog, then wires up the channel that triggers note-list rendering after auth.
func (ui *MainWindowImpl) createPasswordPopUp(w fyne.Window, c *fyne.Container) error {
	ch := make(chan bool, 1)

	// Determine whether we need to generate a new key or just decrypt one.
	keyAction := common.EncryptionKeyAction_Decrypt
	if nCerts, err := ui.certService.CountCerts(); err != nil || nCerts == 0 {
		keyAction = common.EncryptionKeyAction_Generate
	}

	// Try silent / passwordless auto-load before showing any dialog.
	if keyAction == common.EncryptionKeyAction_Decrypt {
		if ok, err := ui.keyService.TryAutoLoad(); err != nil {
			ui.ShowNotification("Error", "Auto-load failed: "+err.Error())
		} else if ok {
			go ui.addNoteList(w, c)
			return nil
		}
	}

	ui.createPasswordDialog(keyAction, ch)

	go func() {
		if !<-ch {
			return
		}
		ui.addNoteList(w, c)
	}()
	return nil
}

// addNoteList is a helper that adds the scrollable note list to the main container.
func (ui *MainWindowImpl) addNoteList(w fyne.Window, c *fyne.Container) {
	noteContainer := container.NewScroll(ui.runNoteList())
	noteContainer.SetMinSize(w.Canvas().Size().Subtract(fyne.NewSize(100, 200)))
	c.Add(noteContainer)
}

// ──────────────────────────────────────────────────────────────────────────────
// Menu
// ──────────────────────────────────────────────────────────────────────────────

func (ui *MainWindowImpl) createMainWindowMenu() *fyne.MainMenu {
	menuItemCopyEncKey := &fyne.MenuItem{
		Label: "Copy encryption key to clipboard",
		Action: func() {
			// Ask for the export password via a small dialog, then delegate to KeyService.
			pwdWdg := widget.NewPasswordEntry()
			pwdWdg.SetPlaceHolder("Password used to protect this key (or leave blank)")
			var exportDg dialog.Dialog
			exportDg = dialog.NewCustom("Export Encryption Key", "Cancel",
				container.NewVBox(
					widget.NewLabel("Enter the key password to encrypt the export:"),
					pwdWdg,
					widget.NewButton("Copy to Clipboard", func() {
						result, err := ui.keyService.ExportKeyForClipboard(pwdWdg.Text)
						if err != nil {
							ui.ShowNotification("Error", err.Error())
							return
						}
						ui.w.Clipboard().SetContent(result)
						exportDg.Hide()
						ui.ShowNotification("Copied", "Encryption key copied to clipboard")
					}),
				), ui.w)
			exportDg.Resize(fyne.NewSize(460, 160))
			exportDg.Show()
		},
	}

	menuItemImportEncKey := &fyne.MenuItem{
		Label: "Import encryption key",
		Action: func() {
			ui.showImportKeyDialog()
		},
	}

	menuItemGenerateEncKey := &fyne.MenuItem{
		Label: "Generate New Encryption Key",
		Action: func() {
			ui.showGenerateKeyDialog(false)
		},
	}

	return fyne.NewMainMenu(&fyne.Menu{
		Label: "File",
		Items: []*fyne.MenuItem{menuItemCopyEncKey, menuItemImportEncKey, menuItemGenerateEncKey},
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Dialogs — Import key
// ──────────────────────────────────────────────────────────────────────────────

// showImportKeyDialog presents the import-key UI and delegates all import logic
// to KeyService.ImportKey.
func (ui *MainWindowImpl) showImportKeyDialog() {
	encKeyWdg := widget.NewEntry()
	encKeyWdg.SetPlaceHolder("Paste ALGO:HEX or raw encrypted hex")
	encAlgoWdg := widget.NewSelect(common.SUPPORTED_ENCRYPTION_ALGORITHMS, func(s string) {
		encKeyWdg.SetPlaceHolder(fmt.Sprintf("Enter %s key", s))
	})
	keyPasswordWdg := widget.NewPasswordEntry()

	wdg := container.NewVBox(
		widget.NewLabel("Paste the exported key string (ALGO:HEX), or a raw encrypted key"),
		encAlgoWdg,
		encKeyWdg,
		widget.NewLabel("Enter the password used to encrypt the key"),
		keyPasswordWdg,
		widget.NewLabel(
			"Attention! By confirming, the key will be saved in the configuration file.\n"+
				"If there already is one, it will be overwritten.\n"+
				"The key must have been generated by EcNotes.",
		),
		widget.NewButton("Confirm", func() {
			if _, err := ui.keyService.ImportKey(encKeyWdg.Text, encAlgoWdg.Selected, keyPasswordWdg.Text); err != nil {
				ui.ShowNotification("Error", err.Error())
				return
			}
			ui.ShowNotification("", "Encryption key imported. All notes have been re-encrypted.")
		}),
	)
	dg := dialog.NewCustom("Import Encryption Key", "Cancel", wdg, ui.w)
	dg.Resize(fyne.NewSize(600, 200))
	dg.Show()
}

// ──────────────────────────────────────────────────────────────────────────────
// Dialogs — Generate / load key
// ──────────────────────────────────────────────────────────────────────────────

// showGenerateKeyDialog presents the key-generation UI.
// isStartup=false means the dialog is opened from the File menu; the app does
// not quit if the user cancels.
func (ui *MainWindowImpl) showGenerateKeyDialog(isStartup bool) {
	ch := make(chan bool, 1)
	_, dg, _ := ui.newCertDialog("Generate New Encryption Key", true, isStartup, ch)
	dg.Resize(fyne.NewSize(600, 650))
	dg.Show()

	go func() {
		if !<-ch {
			return
		}
		keyName, err := ui.getCurEncryptionKeyName()
		if err != nil {
			return
		}
		cert, err := ui.certService.GetCert(keyName)
		if err != nil {
			return
		}
		notes, err := ui.noteService.GetNotes()
		if err != nil {
			ui.ShowNotification("Error", "Error loading notes for re-encryption: "+err.Error())
			return
		}
		if err := ui.keyService.RotateKey(notes, *cert); err != nil {
			ui.ShowNotification("Error", err.Error())
			return
		}
		ui.ShowNotification("", "All notes have been migrated to the new encryption key!")
	}()
}

// loadCertDialog builds the "Decrypt Encryption Key" dialog.
// Auth logic is delegated entirely to KeyService.
func (ui *MainWindowImpl) loadCertDialog(
	keyName string,
	dgTitle string,
	ch chan bool,
) (fyne.CanvasObject, dialog.Dialog, error) {
	notifyResult := func(v bool) {
		select {
		case ch <- v:
		default:
		}
	}

	var (
		wdg        fyne.CanvasObject
		dg         dialog.Dialog
		recoveryDg dialog.Dialog
	)
	mainCompleted := false
	recoveryCompleted := false

	// onConfirm: called when user clicks Confirm in the decrypt dialog.
	onConfirm := func(pwd string) {
		if err := ui.keyService.LoadKey(keyName, pwd); err != nil {
			ui.ShowNotification("Error", err.Error())
			return
		}
		ui.ShowNotification("Success", "Key decrypted successfully")
		mainCompleted = true
		notifyResult(true)
		dg.Hide()
	}

	// onForgotPwd: called when user clicks "Forgot Password?"
	onForgotPwd := func() {
		storedQuestion, err := ui.confService.GetConfig(keyName + "_recovery_question")
		if err != nil || storedQuestion == "" {
			ui.ShowNotification("Error",
				"No recovery question was set up for this key. Use File > Generate New Encryption Key instead.")
			return
		}

		mainCompleted = true
		dg.Hide()

		answerWdg := widget.NewPasswordEntry()
		answerWdg.SetPlaceHolder("Your answer")
		newPwdWdg := widget.NewPasswordEntry()
		newPwdWdg.SetPlaceHolder("New password (optional)")

		recoveryContent := container.NewVBox(
			widget.NewLabel(storedQuestion),
			answerWdg,
			widget.NewLabel("Enter a NEW password (optional):"),
			newPwdWdg,
			widget.NewButton("Recover & Reset Password", func() {
				if answerWdg.Text == "" {
					ui.ShowNotification("Error", "Answer is required")
					return
				}
				// All recovery logic (brute-force delay included) lives in KeyService.
				if err := ui.keyService.VerifyAndRecoverKey(keyName, answerWdg.Text, newPwdWdg.Text); err != nil {
					ui.ShowNotification("Error", err.Error())
					return
				}
				ui.ShowNotification("Success", "Key recovered & password reset successfully")
				recoveryCompleted = true
				notifyResult(true)
				if recoveryDg != nil {
					recoveryDg.Hide()
				}
			}),
		)

		recoveryDg = dialog.NewCustom("Password Recovery", "Cancel", recoveryContent, ui.w)
		recoveryDg.SetOnClosed(func() {
			if !recoveryCompleted {
				notifyResult(false)
			}
		})
		recoveryDg.Resize(fyne.NewSize(500, 280))
		recoveryDg.Show()
	}

	keyPasswordWdg := widget.NewPasswordEntry()
	
	dialogItems := []fyne.CanvasObject{
		widget.NewLabel("Enter the password to decrypt the key (if any)"),
		keyPasswordWdg,
		widget.NewButton("Confirm", func() {
			onConfirm(keyPasswordWdg.Text)
		}),
	}
	
	if ui.keyService.HasRecovery(keyName) {
		dialogItems = append(dialogItems, widget.NewButton("Forgot Password?", onForgotPwd))
	}
	
	wdg = container.NewVBox(dialogItems...)

	// Use a ghost-button-free modal popup for the startup dialog.
	// dialog.NewCustom with an empty dismiss string still renders
	// an invisible button frame (visible artifact). widget.NewModalPopUp
	// gives us full content control with no built-in button bar.
	popup := ui.newModalNoCancel(dgTitle, wdg)
	popup.Show()

	// Wrap in a thin dialog.Dialog adapter so callers can resize/show/hide uniformly.
	dg = &noButtonDialog{popup: popup}
	dg.SetOnClosed(func() {
		if !mainCompleted {
			notifyResult(false)
		}
	})
	return wdg, dg, nil
}

// newCertDialog builds the "Generate Encryption Key" dialog.
// Key creation is delegated entirely to KeyService.GenerateKey.
func (ui *MainWindowImpl) newCertDialog(
	dgTitle string,
	setDefaultKey bool,
	isStartup bool, // false when opened from File menu; true when no keys exist
	ch chan bool,
) (fyne.CanvasObject, dialog.Dialog, error) {
	notifyResult := func(v bool) {
		select {
		case ch <- v:
		default:
		}
	}

	var (
		wdg fyne.CanvasObject
		dg  dialog.Dialog
	)
	completed := false

	keyNameWdg := widget.NewEntry()
	if isStartup {
		keyNameWdg.SetText("ecNotes")
	}
	encAlgoWdg := widget.NewSelect(common.SUPPORTED_ENCRYPTION_ALGORITHMS, func(s string) {})
	keyPasswordWdg := widget.NewPasswordEntry()
	defaultKeyWdg := widget.NewCheck("Set as default key", func(b bool) {})
	if setDefaultKey {
		defaultKeyWdg.SetChecked(true)
		defaultKeyWdg.Disable()
	}

	securityQuestions := []string{
		"What is your childhood hero's name?",
		"What is your first pet's name?",
		"In what city did you meet your spouse?",
		"What was the name of your first school?",
		"What is your mother's maiden name?",
		"What was your childhood nickname?",
	}
	securityQuestionWdg := widget.NewSelect(securityQuestions, func(s string) {})
	securityAnswerWdg := widget.NewPasswordEntry()
	securityAnswerWdg.SetPlaceHolder("Your answer")

	scrollContent := container.NewVBox(
		widget.NewLabel("Enter a name for the key"),
		keyNameWdg,
		widget.NewLabel("Select encryption algorithm"),
		encAlgoWdg,
		widget.NewLabel("Enter a password (optional — leave blank for auto-load at startup)"),
		keyPasswordWdg,
		widget.NewLabel("Password Recovery (Optional)"),
		widget.NewLabel("Choose a security question:"),
		securityQuestionWdg,
		securityAnswerWdg,
		widget.NewLabel(
			"Note: Password is optional.\n"+
				"Without a password, the key loads automatically at startup.\n"+
				"With a password, you can recover using the security question above.",
		),
		defaultKeyWdg,
		widget.NewButton("Confirm", func() {
			if keyNameWdg.Text == "" {
				ui.ShowNotification("Error", "Key name is required")
				return
			}
			if encAlgoWdg.Selected == "" {
				ui.ShowNotification("Error", "Please select an encryption algorithm")
				return
			}

			// Delegate entirely to KeyService — no crypto logic here.
			if _, err := ui.keyService.GenerateKey(
				keyNameWdg.Text,
				encAlgoWdg.Selected,
				keyPasswordWdg.Text,
				defaultKeyWdg.Checked,
				securityQuestionWdg.Selected,
				securityAnswerWdg.Text,
			); err != nil {
				ui.ShowNotification("Error", err.Error())
				return
			}
			ui.ShowNotification("Success", "Key generated successfully")
			completed = true
			notifyResult(true)
			dg.Hide()
		}),
	)
	wdg = container.NewScroll(scrollContent)
	dg = dialog.NewCustom(dgTitle, "Cancel", wdg, ui.w)
	dg.SetOnClosed(func() {
		if !completed {
			notifyResult(false)
		}
	})
	return wdg, dg, nil
}

// createPasswordDialog dispatches to the correct dialog based on keyAction.
func (ui *MainWindowImpl) createPasswordDialog(keyAction common.EncryptionKeyAction, ch chan bool) {
	notifyResult := func(v bool) {
		select {
		case ch <- v:
		default:
		}
	}
	switch keyAction {
	case common.EncryptionKeyAction_Generate:
		_, dg, _ := ui.newCertDialog("Generate Encryption Key", true, true, ch)
		dg.Resize(fyne.NewSize(600, 650))
		dg.Show()
	case common.EncryptionKeyAction_Decrypt:
		keyName, err := ui.getCurEncryptionKeyName()
		if err != nil {
			ui.ShowNotification(common.ERR_KEY_NOT_FOUND, err.Error())
			notifyResult(false)
			return
		}
		_, dg, _ := ui.loadCertDialog(keyName, "Decrypt Encryption Key", ch)
		dg.Resize(fyne.NewSize(500, 200))
		dg.Show()
	default:
		ui.ShowNotification(common.ERR_UNKNOWN_KEY_ACTION, "unknown key action")
		notifyResult(false)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Window lifecycle
// ──────────────────────────────────────────────────────────────────────────────

// CreateWindow initialises the main application window.
func (ui *MainWindowImpl) CreateWindow(title string, width, height float32, _ bool, options map[string]interface{}) {
	ui.ParseDefaultOptions(options)
	w := ui.app.NewWindow(title)
	ui.AddWindow("main", w)
	ui.w = w
	if ui.windowAspect == common.WindowAspect_FullScreen {
		w.SetFullScreen(true)
	} else {
		w.Resize(fyne.NewSize(width, height))
	}
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyF11 {
			ui.ToggleFullScreen(w)
		}
	})

	w.SetMainMenu(ui.createMainWindowMenu())
	mainLayout := ui.createWindowContainer()
	w.SetMaster()
	w.Show()
	_ = ui.createPasswordPopUp(w, mainLayout)
	w.SetContent(mainLayout)
}

// ──────────────────────────────────────────────────────────────────────────────
// Note list
// ──────────────────────────────────────────────────────────────────────────────

func (ui *MainWindowImpl) runNoteList() fyne.CanvasObject {
	titles := ui.noteService.GetTitles()
	if len(titles) == 0 {
		_, err := ui.noteService.GetNotes()
		if err != nil {
			return &widget.Card{
				Title:   "error",
				Content: widget.NewLabel(err.Error()),
			}
		}
		titles = ui.noteService.GetTitles()
		if len(titles) == 0 {
			ui.ShowNotification("", "Note list is empty")
		}
	}
	return ui.createNoteList(titles)
}

func (ui *MainWindowImpl) createNoteList(titles []string) fyne.CanvasObject {
	ui.titlesDataBinding = binding.BindStringList(&titles)
	noteList := widget.NewListWithData(ui.titlesDataBinding,
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	ui.AddWidget(common.WDG_NOTE_LIST, noteList)

	noteList.OnSelected = func(lii widget.ListItemID) {
		ui.selectedNoteID = ui.noteService.GetNoteIDFromTitle(titles[lii])
		note, err := ui.noteService.GetNoteWithContent(ui.selectedNoteID)
		if err != nil {
			if err.Error() == "cipher: message authentication failed" {
				ui.ShowNotification("", common.ERR_CANNOT_DECRYPT_MISSING_KEY)
			} else {
				ui.ShowNotification("", "Error getting note: "+err.Error())
			}
			return
		}
		ui.selectedNote = note
		ui.GetObserver().Notify(
			observer.EVENT_UPDATE_NOTE_WINDOW,
			ui.selectedNote,
			common.WindowMode_Edit,
			common.WindowAction_Update)
		ui.SetWindowVisibility(common.WIN_NOTE_DETAILS, true)
	}

	return noteList
}

// UpdateNoteListWidget is the observer listener that refreshes the note list
// whenever note titles change.
func (ui *MainWindowImpl) UpdateNoteListWidget() observer.Listener {
	return observer.Listener{
		OnNotify: func(titles interface{}, args ...interface{}) {
			if titles == nil {
				return
			}
			uiTitles, ok := titles.([]string)
			if !ok {
				log.Println("UpdateNoteList: invalid message value")
				return
			}
			if ui.titlesDataBinding != nil {
				if err := ui.titlesDataBinding.Set(uiTitles); err != nil {
					log.Println("UpdateNoteList: error setting data:", err)
				}
			}
		},
	}
}
