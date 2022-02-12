package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/service/observer"
)

type MainWindow interface {
	WindowInterface
	UpdateNoteListWidget() observer.Listener
}

type MainWindowImpl struct {
	UImpl
	titlesDataBinding binding.ExternalStringList
}

func NewMainWindow(ui *UImpl) MainWindow {
	return &MainWindowImpl{
		UImpl: *ui,
	}
}

// CreateMainWindow ....
func (ui *MainWindowImpl) CreateWindow(title string, width, height float32, _ bool) {
	// define main windows
	w := ui.app.NewWindow(title)
	ui.AddWindow("main", w)
	w.Resize(fyne.NewSize(width, height))

	mainWinLoaderLabel := widget.NewLabel("Loading main window...")
	mainWinLoader := func(msg string) *widget.Label {
		mainWinLoaderLabel.SetText(msg)
		mainWinLoaderLabel.Alignment = fyne.TextAlignCenter
		return mainWinLoaderLabel
	}

	ch := make(chan bool)
	var winLoaderText string
	// if we have encryption key, show password entry to decrypt and save to global map
	if _, err := ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_KEY); err == nil {
		winLoaderText = "Decrypting encryption key..."
	} else {
		winLoaderText = "Generating encryption key..."
	}

	// create main layout
	searchBox := widget.NewEntry()
	searchBox.SetPlaceHolder("Search in titles")
	ui.AddWidget("search_box", searchBox)
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
		// when new note button is clicked
		fmt.Println("new note button clicked")
	})
	hideBtn := widget.NewButton("Hide", func() {
		// when hide button is clicked
		fmt.Println("hide button clicked")
	})
	deleteNoteBtn := widget.NewButton("Delete", func() {
		// when delete note button is clicked
		fmt.Println("delete note button clicked")
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
	btnContainer := fyne.NewContainerWithLayout(
		btnBarLayout,
		btnBar,
	)

	// horizontal separator
	hSep := widget.NewSeparator()

	// render main layout
	mainLayout := container.NewVBox(
		searchBox,
		// container.NewHBox(
		// 	widget.NewLabel("Search:"),
		// 	searchBox,
		// ),
		btnContainer,
		hSep,
		mainWinLoader(winLoaderText),
	)
	w.SetContent(mainLayout)
	w.SetMaster()
	w.CenterOnScreen()
	w.Show()
	// create the password modal window
	modal := ui.runPasswordPopUp(w, common.EncryptionKeyAction_Decrypt, mainWinLoaderLabel, ch)
	modal.Show()

	go func() {
		<-ch
		noteContainer := container.NewScroll(ui.runNoteList())
		noteContainer.SetMinSize(w.Canvas().Size().Subtract(fyne.NewSize(100, 200)))
		mainLayout.AddObject(noteContainer)
	}()
}

func (ui *MainWindowImpl) runNoteList() fyne.CanvasObject {
	// load notes into a fyne.List
	titles := ui.noteService.GetTitles()
	if len(titles) == 0 {
		// load notes from db (and populate titles array)
		_, err := ui.noteService.GetNotes()
		if err != nil {
			ui.ShowNotification("Error", err.Error())
			return &widget.Card{
				Title: "Error",
				Content: widget.NewLabel(
					err.Error(),
				),
			}
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

	ui.AddWidget("note_list", noteList)
	noteList.OnSelected = func(lii widget.ListItemID) {
		// when item is selected
		fmt.Println("DEBUG: note list item selected:", lii)
	}

	return noteList
}

// UpdateNoteList listener (observer) triggered when the note tiles are updated
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
			// TODO: take in accoung the search box
			if ui.titlesDataBinding != nil {
				if err := ui.titlesDataBinding.Set(uiTitles); err != nil {
					log.Println("UpdateNoteList: error setting data:", err)
					return
				}
			}
			// update note list widget
			// if w, ok := ui.widgets["note_list"]; ok {
			// 	widgetList := w.(*widget.List)
			// 	widgetList.Refresh()
			// }
		},
	}
}

func (ui *MainWindowImpl) runPasswordPopUp(
	w fyne.Window,
	keyAction common.EncryptionKeyAction,
	mainWinLoader *widget.Label,
	ch chan bool,
) (modal *widget.PopUp) {
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
				decKey, err = cryptoUtil.SecureRandomStr(common.ENCRYPTION_KEY_LENGTH)
				if err != nil {
					ui.ShowNotification("Error generating encryption key", err.Error())
					return
				}
				ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_KEY, decKey)
				// encrypt the key with password input in the password entry
				if encKey, err = cryptoUtil.EncryptMessage(decKey, pwdWg.Text); err != nil {
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
				if decKey, err = cryptoUtil.DecryptMessage(encKey, pwdWg.Text); err != nil {
					ui.ShowNotification("Error decrypting encryption key", err.Error())
					return
				}
				ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_KEY, decKey)
				ui.ShowNotification("Encryption key decrypted and stored in memory till app is closed", "")
			default:
				ui.ShowNotification("Error", "Unknown key action")
			}

			modal.Hide()
			// reset password entry for security
			pwdWg.SetText("")
			// hide main window loader
			mainWinLoader.Hidden = true
			ch <- true
		})
	)

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
