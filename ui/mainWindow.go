package ui

import (
	"fmt"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
)

// UI ....
type UI interface {
	CreateMainWindow()
	AddWindow(name string, w fyne.Window)
	AddWidget(name string, w fyne.CanvasObject)
	GetWindow(name string) (fyne.Window, error)
	GetWidget(name string) (fyne.CanvasObject, error)
	UpdateNoteListWidget() observer.Listener
}

// UImpl Main ui configuration
type UImpl struct {
	app               fyne.App
	windows           map[string]fyne.Window
	winMux            *sync.Mutex
	widgets           map[string]fyne.CanvasObject
	widMux            *sync.Mutex
	confSrv           service.ConfigService
	noteService       service.NoteService
	titlesDataBinding binding.ExternalStringList
}

// NewUI UI constructor
func NewUI(
	app fyne.App,
	confSrv service.ConfigService,
	noteService service.NoteService,
) *UImpl {
	return &UImpl{
		app:         app,
		windows:     make(map[string]fyne.Window),
		widgets:     make(map[string]fyne.CanvasObject),
		winMux:      &sync.Mutex{},
		widMux:      &sync.Mutex{},
		confSrv:     confSrv,
		noteService: noteService,
	}
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

// CreateMainWindow ....
func (ui *UImpl) CreateMainWindow() {
	// define main windows
	w := ui.app.NewWindow("EcNotes")
	ui.AddWindow("main", w)
	w.Resize(fyne.NewSize(800, 800))

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
		fmt.Println("searchBox changed:", text)
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

	ui.app.Run()
}

func (ui *UImpl) runNoteList() fyne.CanvasObject {
	// load notes into a fyne.List
	titles := ui.noteService.GetTitles()
	if len(titles) == 0 {
		// load notes from db (and populate titles array)
		_, err := ui.noteService.GetNotes()
		if err != nil {
			ui.showNotification("Error", err.Error())
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

func (ui *UImpl) createNoteList(titles []string) fyne.CanvasObject {
	ui.titlesDataBinding = binding.BindStringList(&titles)
	noteList := widget.NewListWithData(ui.titlesDataBinding,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})
	// noteList := widget.NewList(
	// 	// lets change item count from 3 to 30
	// 	func() int {
	// 		notesCount := len(titles)
	// 		return notesCount
	// 	},
	// 	func() fyne.CanvasObject {
	// 		return widget.NewLabel("")
	// 	},
	// 	// last one
	// 	func(lii widget.ListItemID, co fyne.CanvasObject) {
	// 		// update data of widget
	// 		co.(*widget.Label).SetText(titles[lii])
	// 	},
	// )
	ui.AddWidget("note_list", noteList)
	noteList.OnSelected = func(lii widget.ListItemID) {
		// when item is selected
		fmt.Println("DEBUG: note list item selected:", lii)
	}

	// add random titles to data in a goroutine
	// go func() {
	// 	for i := 0; i < 10; i++ {
	// 		time.Sleep(time.Second)
	// 		ui.titlesDataBinding.Append(fmt.Sprintf("title %d", i))
	// 	}
	// }()

	return noteList
}

// UpdateNoteList listener (observer) triggered when the note tiles are updated
func (ui *UImpl) UpdateNoteListWidget() observer.Listener {
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
			fmt.Printf("DEBUG: updateNoteListWidget called. note titles are:\n%#v\n", titles)
			// update note list widget
			// if w, ok := ui.widgets["note_list"]; ok {
			// 	widgetList := w.(*widget.List)
			// 	widgetList.Refresh()
			// }
		},
	}
}

func (ui *UImpl) runPasswordPopUp(
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
					ui.showNotification("Error generating encryption key", err.Error())
					return
				}
				ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_KEY, decKey)
				// encrypt the key with password input in the password entry
				if encKey, err = cryptoUtil.EncryptMessage(decKey, pwdWg.Text); err != nil {
					ui.showNotification("Error encrypting encryption key", err.Error())
					return
				}
				// save encrypted encryption key to config file
				ui.confSrv.SetConfig(common.CONFIG_ENCRYPTION_KEY, encKey)
				if err := ui.confSrv.SaveConfig(); err != nil {
					ui.showNotification("Error saving configuration", err.Error())
					return
				}
				ui.showNotification("Encryption key generated", "")
			case common.EncryptionKeyAction_Decrypt:
				// decrypt the key with password input in the password entry
				if encKey, err = ui.confSrv.GetConfig(common.CONFIG_ENCRYPTION_KEY); err != nil {
					ui.showNotification("Error loading encryption key from app configuration", err.Error())
					return
				}
				if decKey, err = cryptoUtil.DecryptMessage(encKey, pwdWg.Text); err != nil {
					ui.showNotification("Error decrypting encryption key", err.Error())
					return
				}
				ui.confSrv.SetGlobal(common.CONFIG_ENCRYPTION_KEY, decKey)
				ui.showNotification("Encryption key decrypted and stored in memory till app is closed", "")
			default:
				ui.showNotification("Error", "Unknown key action")
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

func (ui *UImpl) showNotification(title, contentStr string) {
	time.Sleep(time.Millisecond * 500)
	ui.app.SendNotification(fyne.NewNotification(title, contentStr))
}
