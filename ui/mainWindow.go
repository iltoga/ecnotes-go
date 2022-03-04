package ui

import (
	"errors"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
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
	w                 fyne.Window
	cryptoService     service.CryptoServiceFactory
}

func NewMainWindow(
	ui *UImpl,
	cryptoService service.CryptoServiceFactory) MainWindow {
	return &MainWindowImpl{
		UImpl:         *ui,
		cryptoService: cryptoService,
	}
}

// GetWindow returns the window object
func (ui *MainWindowImpl) GetWindow() fyne.Window {
	return ui.w
}

// createWindowContainer creates a container with the window content
func (ui *MainWindowImpl) createWindowContainer() *fyne.Container {
	var winLoaderText string
	// if we have encryption key, show password entry to decrypt and save to global map
	if _, err := ui.confSrv.GetConfigBytes(common.CONFIG_ENCRYPTION_KEY); err == nil {
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
			Notify(observer.EVENT_CREATE_NOTE_WINDOW, new(model.Note), common.WindowMode_Edit, common.WindowAction_New)
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
func (ui *MainWindowImpl) createPasswordPopUp(w fyne.Window, c *fyne.Container) error {
	ch := make(chan bool)
	// check if we have encryption key in the config
	keyAction := common.EncryptionKeyAction_Decrypt
	if _, err := ui.confSrv.GetConfigBytes(common.CONFIG_ENCRYPTION_KEY); err != nil {
		keyAction = common.EncryptionKeyAction_Generate
	}
	ui.createPasswordDialog(keyAction, ch)

	go func() {
		<-ch
		noteContainer := container.NewScroll(ui.runNoteList())
		noteContainer.SetMinSize(w.Canvas().Size().Subtract(fyne.NewSize(100, 200)))
		c.Add(noteContainer)
	}()

	return nil
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
			// get encryption algorithm
			encAlgo, err := ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_ALGORITHM)
			if err != nil {
				ui.ShowNotification("", "It looks like encryption algorithm has not been set yet")
				return
			}
			// content to be copied to clipboard
			content := fmt.Sprintf("%s:%s", encAlgo, encKey)
			ui.w.Clipboard().SetContent(content)
		},
	}
	menuItemImportEncKey := &fyne.MenuItem{
		Label: "Import encryption key",
		Action: func() {
			onConfirm := func(encKey, encAlgo, keyPwd string) {
				if encKey == "" {
					ui.ShowNotification("", "Encrypted key is empty! Canceling...")
					return
				}
				// try to decrypt the key with the password using aes-256-cbc in crytpoUtils
				key, err := cryptoUtil.DecryptAES256([]byte(keyPwd), []byte(encKey))
				if err != nil {
					ui.ShowNotification("Invalid Key", "Error decrypting key: "+err.Error())
					return
				}

				if err := ui.confSrv.SetConfig(common.CONFIG_ENCRYPTION_KEY, encKey); err != nil {
					ui.ShowNotification("Error saving encryption key to configuration", err.Error())
					return
				}
				if err := ui.confSrv.SaveConfig(); err != nil {
					ui.ShowNotification("Error updating configuration file", err.Error())
					return
				}
				// validate encryption algorithm against supported algorithms (in constants.go)
				if !common.IsSupportedEncryptionAlgorithm(encAlgo) {
					ui.ShowNotification("Error", "Unsupported encryption algorithm")
					return
				}
				if err := ui.confSrv.SetConfig(common.CONFIG_ENCRYPTION_ALGORITHM, encAlgo); err != nil {
					ui.ShowNotification("Error saving encryption algorithm to configuration", err.Error())
					return
				}

				ui.ShowNotification(
					"",
					"Encryption key has been imported successfully. Now all notes will be re-encrypted",
				)

				// now reflect the changes in the app by setting the new encryption algo and key and re-encrypting all notes
				// get encryption algorithm and key from configuration

				// get all notes
				notes, err := ui.noteService.GetNotes()
				if err != nil {
					ui.ShowNotification("Error", "Error getting all notes: "+err.Error())
					return
				}
				// re-encrypt all notes with the new encryption key
				if err := ui.noteService.ReEncryptNotes(notes, encAlgo, key); err != nil {
					ui.ShowNotification("Error", "Error re-encrypting notes: "+err.Error())
					return
				}
				ui.ShowNotification("", "All notes have been re-encrypted successfully")
			}

			// create a widget with vertical layout and as a content: a label, a text input field, a slect list and a button
			// to confirm the input
			encKeyWdg := widget.NewEntry()
			// select widget with supported encryption algorithms
			encAlgoWdg := widget.NewSelect(common.SUPPORTED_ENCRYPTION_ALGORITHMS, func(s string) {
				encKeyWdg.SetPlaceHolder(fmt.Sprintf("Enter %s key", s))
			})
			keyPasswordWdg := widget.NewPasswordEntry()
			wdg := container.NewVBox(
				widget.NewLabel("Enter the encryption key"),
				encAlgoWdg,
				encKeyWdg,
				widget.NewLabel("Enter the password to decrypt the key"),
				keyPasswordWdg,
				widget.NewLabel(
					"Attention! By confirming, the key will be saved in the configuration file.\n"+
						"If there is already one, it will be overwritten.\n"+
						"The above encryption key must have been generated by EcNotes, to guarantee that will work with the application.",
				),
				widget.NewButton("Confirm", func() {
					onConfirm(encKeyWdg.Text, encAlgoWdg.Selected, keyPasswordWdg.Text)
				}),
			)
			dg := dialog.NewCustom("Import Encryption Key", "Cancel", wdg, ui.w)
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
			return &widget.Card{
				Title: "error",
				Content: widget.NewLabel(
					err.Error(),
				),
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
			observer.EVENT_UPDATE_NOTE_WINDOW,
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

func (ui *MainWindowImpl) createPasswordDialog(keyAction common.EncryptionKeyAction, ch chan bool) {
	var (
		// encryptedKey, decryptedKey []byte
		// decryptedKeyStr            string
		wdg fyne.CanvasObject
		dg  dialog.Dialog
		// err     error
		exitApp = true
		dgTitle string
	)

	doReturn := func(keyAction common.EncryptionKeyAction, ch chan bool, err error, algo string, key []byte) {
		if err != nil {
			ui.ShowNotification("Error", err.Error())
			dg.Hide()
			return
		}
		if keyAction == common.EncryptionKeyAction_Decrypt {
			ui.ShowNotification("Success", "Key decrypted successfully")
		} else {
			ui.ShowNotification("Success", "Key encrypted successfully")
		}

		exitApp = false
		dg.Hide()
		ch <- true
	}

	switch keyAction {
	case common.EncryptionKeyAction_Generate:
		dgTitle = "Generate Encryption Key"
		onConfirm := func(algo string, pwd string) {
			ui.cryptoService.SetSrv(service.NewCryptoServiceFactory(algo))
			// generate encryption key
			decryptedKey, err := ui.cryptoService.GetSrv().GetKeyManager().GenerateKey()
			if err != nil {
				doReturn(keyAction, ch, err, algo, nil)
				return
			}
			// encrypt the key with password input in the password entry
			encryptedKey, err := cryptoUtil.EncryptMessage(decryptedKey, pwd)
			if err != nil {
				err = fmt.Errorf("error encrypting encryption key: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			// save encrypted encryption key to config file
			err = ui.confSrv.SetConfigBytes(common.CONFIG_ENCRYPTION_KEY, encryptedKey)
			if err != nil {
				err = fmt.Errorf("error setting configuration item: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			ui.confSrv.SetGlobalBytes(common.CONFIG_ENCRYPTION_KEY, decryptedKey)

			// save encryption algo to config file
			err = ui.confSrv.SetConfig(common.CONFIG_ENCRYPTION_ALGORITHM, algo)
			if err != nil {
				err = fmt.Errorf("error setting configuration item: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			if err = ui.confSrv.SaveConfig(); err != nil {
				err = fmt.Errorf("error saving configuration: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_ALGORITHM, algo)
			doReturn(keyAction, ch, err, algo, decryptedKey)
		}
		// select widget with supported encryption algorithms
		encAlgoWdg := widget.NewSelect(common.SUPPORTED_ENCRYPTION_ALGORITHMS, func(s string) {
		})
		keyPasswordWdg := widget.NewPasswordEntry()
		wdg = container.NewVBox(
			widget.NewLabel("Select encryption algorithm you want to use"),
			encAlgoWdg,
			widget.NewLabel("Enter the password to encrypt the key"),
			keyPasswordWdg,
			widget.NewLabel(
				"Attention!\n"+
					"keep the password in mind or write it down and put it in a safe place.\n"+
					"If you lose it the only way to read your notes will be\n"+
					"to brute force the encrypted key ;)",
			),
			widget.NewButton("Confirm", func() {
				// validate input fields first
				if encAlgoWdg.Selected == "" {
					ui.ShowNotification("Error", "please select an encryption algorithm")
					return
				}
				if keyPasswordWdg.Text == "" {
					ui.ShowNotification("Error", "password cannot be empty")
					return
				}
				onConfirm(encAlgoWdg.Selected, keyPasswordWdg.Text)
			}),
		)
	case common.EncryptionKeyAction_Decrypt:
		dgTitle = "Decrypt Encryption Key"
		onConfirm := func(pwd string) {
			// get algo from config file
			algo, err := ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_ALGORITHM)
			if err != nil {
				err = fmt.Errorf("error getting configuration item: %s", err.Error())
				doReturn(keyAction, ch, err, algo, nil)
				return
			}

			// decrypt the key with password input in the password entry
			encryptedKey, err := ui.confSrv.GetConfigBytes(common.CONFIG_ENCRYPTION_KEY)
			if err != nil {
				err = fmt.Errorf("error loading encryption key from app configuration: %s", err.Error())
				doReturn(keyAction, ch, err, algo, nil)
				return
			}
			decryptedKey, err := cryptoUtil.DecryptMessage(encryptedKey, pwd)
			if err != nil {
				err = fmt.Errorf("error decrypting encryption key: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			ui.cryptoService.SetSrv(service.NewCryptoServiceFactory(algo))
			// import decrypted key to crypto service to validate it
			if err = ui.cryptoService.GetSrv().GetKeyManager().ImportKey(decryptedKey); err != nil {
				err = fmt.Errorf("error importing key: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			// TODO: this is probably no necessary as the key is already imported and we can use the one above
			// get the key from crypto service
			if decryptedKey, err = ui.cryptoService.GetSrv().GetKeyManager().GetPrivateKey(); err != nil {
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			// save decrypted encryption key to config file
			ui.confSrv.SetGlobalBytes(common.CONFIG_ENCRYPTION_KEY, decryptedKey)
			if err = ui.confSrv.SaveConfig(); err != nil {
				err = fmt.Errorf("error saving configuration: %s", err.Error())
				doReturn(keyAction, ch, err, algo, decryptedKey)
				return
			}
			doReturn(keyAction, ch, err, algo, decryptedKey)
		}
		keyPasswordWdg := widget.NewPasswordEntry()
		wdg = container.NewBorder(
			widget.NewLabel("Enter the password to decrypt the key"),
			widget.NewButton("Confirm", func() {
				// validate input fields first
				if keyPasswordWdg.Text == "" {
					ui.ShowNotification("Error", "password cannot be empty")
					return
				}
				onConfirm(keyPasswordWdg.Text)
			}),
			nil, nil,
			keyPasswordWdg,
		)
	default:
		err := errors.New("unknown key action")
		doReturn(keyAction, ch, err, "", nil)
		return
	}

	dg = dialog.NewCustom(dgTitle, "Cancel", wdg, ui.w)
	dg.SetOnClosed(func() {
		if exitApp {
			ui.Stop()
		}
	})
	dg.Resize(fyne.NewSize(500, 200))
	dg.Show()
}
