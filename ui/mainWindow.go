package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
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
	selectedNote      *service.Note
	w                 fyne.Window
}

func NewMainWindow(ui *UImpl) MainWindow {
	return &MainWindowImpl{
		UImpl: *ui,
	}
}

// createWindowContainer creates a container with the window content
func (ui *MainWindowImpl) createWindowContainer() *fyne.Container {
	var winLoaderText string
	// if we have encryption key, show password entry to decrypt and save to global map
	if _, err := ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_KEY); err == nil {
		winLoaderText = "Decrypting encryption key..."
	} else {
		winLoaderText = "Generating encryption key..."
	}

	// create main layout
	searchBox := widget.NewEntry()
	searchBox.SetPlaceHolder("Search note titles")
	ui.AddWidget(common.WDG_SEARCH_BOX, searchBox)
	searchBox.OnChanged = func(text string) {
		// when search box is changed
		// use fuzzy search to find titles that match the search text
		_, err := ui.noteService.SearchNotes(text, true)
		if err != nil {
			ui.ShowNotification("Error searching notes", err.Error())
			return
		}
	}

	// create buttons
	newNoteBtn := widget.NewButton("New", func() {
		ui.GetObserver().
			Notify(observer.EVENT_CREATE_NOTE, new(service.Note), common.WindowMode_Edit, common.WindowAction_New)
		// set note details window to be visible
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
			// this is to allow user to see the note content unencrypted in the note details window
			ui.selectedNote.Content = unEncContent
		}
	})
	deleteNoteBtn := widget.NewButton("Delete", func() {
		if ui.selectedNote != nil {
			// delete note from db
			err := ui.noteService.DeleteNote(ui.selectedNote.ID)
			if err != nil {
				ui.ShowNotification("Error deleting note", err.Error())
				return
			}
		}
	})

	btnBar := container.New(
		layout.NewHBoxLayout(),
		newNoteBtn,
		hideBtn,
		deleteNoteBtn,
	)

	btnBarLayout := layout.NewBorderLayout(
		nil,
		nil,
		nil,
		btnBar,
	)
	btnContainer := container.New(
		btnBarLayout,
		btnBar,
	)

	// horizontal separator
	hSep := widget.NewSeparator()

	// TODO: delete this if never shown
	mainWinLoaderLabel := widget.NewLabel("Loading main window...")
	mainWinLoaderLabel.Hidden = true
	mainWinLoader := func(msg string) *widget.Label {
		mainWinLoaderLabel.SetText(msg)
		mainWinLoaderLabel.Alignment = fyne.TextAlignCenter
		return mainWinLoaderLabel
	}

	// render main layout
	return container.NewVBox(
		searchBox,
		btnContainer,
		hSep,
		mainWinLoader(winLoaderText),
	)
}

// createPasswordPopUp creates a modal dialog to enter password
func (ui *MainWindowImpl) createPasswordPopUp(w fyne.Window, c *fyne.Container) *widget.PopUp {
	ch := make(chan bool)
	// check if we have encryption key in the config
	keyAction := common.EncryptionKeyAction_Decrypt
	if _, err := ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_KEY); err != nil {
		keyAction = common.EncryptionKeyAction_Generate
	}
	modal := ui.runPasswordPopUp(w, keyAction, ch)
	modal.Show()
	if pwdWg, err := ui.GetWidget(common.WDG_PASSWORD_MODAL); err == nil {
		ui.SetFocusOnWidget(w, pwdWg.(*widget.Entry))
	}

	go func() {
		<-ch
		noteContainer := container.NewScroll(ui.runNoteList())
		noteContainer.SetMinSize(w.Canvas().Size().Subtract(fyne.NewSize(100, 200)))
		c.Add(noteContainer)
	}()

	return modal
}

// createMainWindowMenu creates the main window menu
func (ui *MainWindowImpl) createMainWindowMenu() *fyne.MainMenu {
	menuItemCopyEncKey := &fyne.MenuItem{
		Label: "Copy encryption key to clipboard",
		Action: func() {
			// get the (password encrypted) encryption key
			encKey, err := ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_KEY)
			if err != nil {
				ui.ShowNotification("", "It looks like encryption key has not been generated yet")
				return
			}
			ui.w.Clipboard().SetContent(encKey)
		},
	}
	menuItemImportEncKey := &fyne.MenuItem{
		Label: "Import encryption key",
		Action: func() {
			onConfirm := func(s string) {
				if s == "" {
					ui.ShowNotification("", "Encrypted key is empty! Canceling...")
					return
				}
				if err := ui.confSrv.SetConfig(common.CONFIG_ENCRYPTION_KEY, s); err != nil {
					ui.ShowNotification("Error saving encryption key to configuration", err.Error())
					return
				}
				if err := ui.confSrv.SaveConfig(); err != nil {
					ui.ShowNotification("Error updating configuration file", err.Error())
					return
				}
				ui.ShowNotification("", "Encryption key has been imported successfully")
			}
			dg := dialog.NewEntryDialog(
				"Import Encryption Key. Attention! This will overwrite the existing key in configuration file",
				"",
				onConfirm,
				ui.w,
			)
			dg.Resize(fyne.NewSize(600, 200))
			dg.Show()
		},
	}

	menuItems := []*fyne.MenuItem{menuItemCopyEncKey, menuItemImportEncKey}
	menu := &fyne.Menu{
		Label: "File",
		Items: menuItems,
	}
	return fyne.NewMainMenu(menu)
}

// CreateWindow ....
func (ui *MainWindowImpl) CreateWindow(title string, width, height float32, _ bool, options map[string]interface{}) {
	// init window
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

	// create main window menu
	w.SetMainMenu(ui.createMainWindowMenu())

	// create window container
	mainLayout := ui.createWindowContainer()
	w.SetMaster()
	w.Show()
	_ = ui.createPasswordPopUp(w, mainLayout)
	w.SetContent(mainLayout)
	// w.CenterOnScreen()
}

func (ui *MainWindowImpl) runNoteList() fyne.CanvasObject {
	// load notes into a fyne.List
	titles := ui.noteService.GetTitles()
	if len(titles) == 0 {
		// load notes from db (and populate titles array)
		_, err := ui.noteService.GetNotes()
		if err != nil {
			ui.ShowNotification("", "Note list is empty")
			// ui.ShowNotification("Error", err.Error())
			// return &widget.Card{
			// 	Title: "Error",
			// 	Content: widget.NewLabel(
			// 		err.Error(),
			// 	),
			// }
		}
		titles = ui.noteService.GetTitles()
	}
	return ui.createNoteList(titles)
}

func (ui *MainWindowImpl) createNoteList(titles []string) fyne.CanvasObject {
	ui.titlesDataBinding = binding.BindStringList(&titles)
	noteList := widget.NewListWithData(ui.titlesDataBinding,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	ui.AddWidget(common.WDG_NOTE_LIST, noteList)

	noteList.OnSelected = func(lii widget.ListItemID) {
		var err error
		// get note from db
		noteID := ui.noteService.GetNoteIDFromTitle(titles[lii])
		ui.selectedNote, err = ui.noteService.GetNoteWithContent(noteID)
		if err != nil {
			ui.ShowNotification("Error Loading note from db", err.Error())
			return
		}
		ui.GetObserver().Notify(
			observer.EVENT_UPDATE_NOTE,
			ui.selectedNote,
			common.WindowMode_Edit,
			common.WindowAction_Update)
		ui.SetWindowVisibility(common.WIN_NOTE_DETAILS, true)
	}

	return noteList
}

// UpdateNoteList listener (observer) triggered when the note tiles are u
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
			// TODO: take in account the search box
			if ui.titlesDataBinding != nil {
				if err := ui.titlesDataBinding.Set(uiTitles); err != nil {
					log.Println("UpdateNoteList: error setting data:", err)
					return
				}
			}
		},
	}
}

func (ui *MainWindowImpl) runPasswordPopUp(
	w fyne.Window,
	keyAction common.EncryptionKeyAction,
	ch chan bool,
) (modal *widget.PopUp) {
	var (
		pwdWg = widget.NewPasswordEntry()
		// close the modal window when the password is submitted
		clearAndClose = func() {
			if modal.Visible() {
				modal.Hide()
			}
			// reset password entry for security
			pwdWg.SetText("")

			// set focus on serach box
			if wg, err := ui.GetWidget(common.WDG_SEARCH_BOX); err == nil {
				ui.SetFocusOnWidget(w, wg.(*widget.Entry))
			}
			ch <- true
		}

		popUpText = widget.NewLabel("Enter password")
		btnWg     = widget.NewButton("OK", func() {
			ui.submitPassword(keyAction, pwdWg.Text)
			clearAndClose()
		})
	)

	pwdWg.OnSubmitted = func(pwd string) {
		ui.submitPassword(keyAction, pwdWg.Text)
		clearAndClose()
	}

	// add widgets to widgets map
	ui.AddWidget(common.WDG_PASSWORD_MODAL, pwdWg)
	ui.AddWidget(common.BTN_PASSWORD_MODAL, btnWg)

	if keyAction == common.EncryptionKeyAction_Decrypt {
		btnWg.SetText("Enter password to Decrypt Key")
	} else {
		btnWg.SetText("Enter password to Encrypt generated Key")
	}

	modal = widget.NewModalPopUp(
		container.New(
			layout.NewVBoxLayout(),
			popUpText,
			pwdWg,
			btnWg,
		),
		w.Canvas(),
	)
	return modal
}

func (ui *MainWindowImpl) submitPassword(keyAction common.EncryptionKeyAction, password string) {
	var (
		encKey, decKey string
		err            error
	)
	// generate encryption key, encrypt with password and save to file
	switch keyAction {
	case common.EncryptionKeyAction_Generate:
		// generate encryption key
		decKey, err = cryptoUtil.SecureRandomStr(common.ENCRYPTION_KEY_LENGTH)
		if err != nil {
			ui.ShowNotification("Error generating encryption key", err.Error())
			return
		}
		ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_KEY, decKey)
		// encrypt the key with password input in the password entry
		if encKey, err = cryptoUtil.EncryptMessage(decKey, password); err != nil {
			ui.ShowNotification("Error encrypting encryption key", err.Error())
			return
		}
		// save encrypted encryption key to config file
		ui.confSrv.SetConfig(common.CONFIG_ENCRYPTION_KEY, encKey)
		if err := ui.confSrv.SaveConfig(); err != nil {
			ui.ShowNotification("Error saving configuration", err.Error())
			return
		}
		ui.ShowNotification("Encryption key generated", "")
	case common.EncryptionKeyAction_Decrypt:
		// decrypt the key with password input in the password entry
		if encKey, err = ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_KEY); err != nil {
			ui.ShowNotification("Error loading encryption key from app configuration", err.Error())
			return
		}
		if decKey, err = cryptoUtil.DecryptMessage(encKey, password); err != nil {
			ui.ShowNotification("Error decrypting encryption key", err.Error())
			return
		}
		ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_KEY, decKey)
		ui.ShowNotification("Encryption key decrypted and stored in memory till app is closed", "")
	default:
		ui.ShowNotification("Error", "Unknown key action")
	}
}
