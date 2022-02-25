package ui

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service/observer"
)

// NoteDetailsWindow ....
type NoteDetailsWindow interface {
	WindowInterface
	UpdateNoteDetailsWidget() observer.Listener
	Close(clearData bool)
}

// NoteDetailsWindowImpl ....
type NoteDetailsWindowImpl struct {
	UImpl
	WindowDefaultOptions
	note     *model.Note
	oldTitle string // in case we update the note title we need to save the old one to be able to save the note
	w        fyne.Window
}

// NewNoteDetailsWindow ....
func NewNoteDetailsWindow(ui *UImpl, note *model.Note) NoteDetailsWindow {
	return &NoteDetailsWindowImpl{
		UImpl: *ui,
		note:  note,
	}
}

// GetWindow returns the window object
func (ui *NoteDetailsWindowImpl) GetWindow() fyne.Window {
	return ui.w
}

// ParseDefaultOptions ....
func (ui *NoteDetailsWindowImpl) ParseDefaultOptions(options map[string]interface{}) {
	if val := common.GetMapValOrNil(options, common.OPT_WINDOW_ACTION); val != nil {
		if mode, ok := val.(common.WindowAction); ok {
			ui.windowAction = mode
		}
	}
	if val := common.GetMapValOrNil(options, common.OPT_WINDOW_MODE); val != nil {
		if mode, ok := val.(common.WindowMode); ok {
			ui.windowMode = mode
		}
	}
	if val := common.GetMapValOrNil(options, common.OPT_WINDOW_ASPECT); val != nil {
		if aspect, ok := val.(common.WindowAspect); ok {
			ui.windowAspect = aspect
		}
	}
}

// CreateWindow ....
func (ui *NoteDetailsWindowImpl) CreateWindow(
	title string,
	width, height float32,
	visible bool,
	options map[string]interface{},
) {
	// init window
	ui.ParseDefaultOptions(options)
	w := ui.app.NewWindow(title)
	ui.AddWindow(common.WIN_NOTE_DETAILS, w)
	ui.w = w
	if ui.windowAspect == common.WindowAspect_FullScreen {
		w.SetFullScreen(true)
	} else {
		w.Resize(fyne.NewSize(width, height))
	}

	w.SetContent(ui.createFormWidget(w))
	w.CenterOnScreen()
	// TODO: find a more elegant way to recreate the window when it is closed
	// this is to avoid the note details window to be destroyed when the user closes it
	w.SetOnClosed(func() {
		go func() {
			time.Sleep(time.Millisecond * 500)
			ui.CreateWindow(title, width, height, visible, options)
			ui.SetWindowVisibility(common.WIN_NOTE_DETAILS, false)
		}()
		// unselect all elements in the notes list, to allow the user to re-select the same note
		if wdg, err := ui.GetWidget(common.WDG_NOTE_LIST); err == nil {
			wdg.(*widget.List).UnselectAll()
		}
	})

	if visible {
		w.Show()
		if titleWidget, err := ui.GetWidget(common.WDG_NOTE_DETAILS_TITLE); err == nil {
			ui.SetFocusOnWidget(w, titleWidget.(*widget.Entry))
		} else {
			log.Printf("Error getting title widget: %v", err)
		}
	}
}

func (ui *NoteDetailsWindowImpl) updateWidgetsData(n *model.Note) {
	ui.note = n
	// save the note title in case we update it (we need the old one to be able to save the note)
	ui.oldTitle = n.Title
	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_TITLE); err == nil {
		w.(*widget.Entry).SetText(n.Title)
	} else {
		log.Printf("Error getting title widget: %v", err)
	}

	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_CONTENT); err == nil {
		w.(*widget.Entry).SetText(n.Content)
	} else {
		log.Printf("Error getting content widget: %v", err)
	}

	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_CONTENT_RICH_TEXT); err == nil {
		w.(*widget.RichText).ParseMarkdown(n.Content)
	} else {
		log.Printf("Error getting content widget: %v", err)
	}

	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_HIDDEN); err == nil {
		w.(*widget.Check).SetChecked(n.Hidden)
	} else {
		log.Printf("Error getting hidden checkbox widget: %v", err)
	}

	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_ENCRYPTED); err == nil {
		w.(*widget.Check).SetChecked(n.Encrypted)
	} else {
		log.Printf("Error getting encrypted checkbox widget: %v", err)
	}

	createdAtStr := common.FormatTime(common.TimestampToTime(n.CreatedAt))
	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_CREATED_AT); err == nil {
		w.(*widget.Label).SetText(fmt.Sprintf("Created:  %s", createdAtStr))
	} else {
		log.Printf("Error getting createdAt widget: %v", err)
	}

	updatedAtStr := common.FormatTime(common.TimestampToTime(n.UpdatedAt))
	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_UPDATED_AT); err == nil {
		w.(*widget.Label).SetText(fmt.Sprintf("Updated:  %s", updatedAtStr))
	} else {
		log.Printf("Error getting updatedAt widget: %v", err)
	}
}

// UpdateNoteDetailsWidget ....
func (ui *NoteDetailsWindowImpl) UpdateNoteDetailsWidget() observer.Listener {
	return observer.Listener{
		OnNotify: func(note interface{}, args ...interface{}) {
			if note == nil {
				return
			}
			n, ok := note.(*model.Note)
			if !ok {
				log.Printf("Error cannot cast note struct: %v", note)
				return
			}

			// parse args
			if len(args) > 0 {
				if windowMode, ok := args[0].(common.WindowMode); ok {
					ui.windowMode = windowMode
				}
				if len(args) > 1 {
					if windowAction, ok := args[1].(common.WindowAction); ok {
						ui.windowAction = windowAction
					}
				}
				if len(args) > 2 {
					if windowAspect, ok := args[2].(common.WindowAspect); ok {
						ui.windowAspect = windowAspect
					}
				}
			}
			ui.setWidgetsStatus()
			ui.updateWidgetsData(n)
		},
	}
}

// Close close note details window
func (ui *NoteDetailsWindowImpl) Close(clearData bool) {
	// just to make sure nothing is left in the window
	if clearData {
		ui.updateWidgetsData(new(model.Note))
	}
	ui.w.Close()
}

// saveNote save a new note
func (ui *NoteDetailsWindowImpl) saveNote() (note *model.Note, err error) {
	if err = ui.noteService.CreateNote(ui.note); err != nil {
		return
	}
	return ui.note, nil
}

// updateNote update an existing note
func (ui *NoteDetailsWindowImpl) updatenote() (noteID int, err error) {
	// try update the title, if changed
	if ui.note.Title != ui.oldTitle && ui.note.Title != "" && ui.oldTitle != "" {
		if noteID, err = ui.noteService.UpdateNoteTitle(ui.oldTitle, ui.note.Title); err != nil {
			err = fmt.Errorf("error updating note title: %v", err)
			return
		}
		// since we have deleted the old note from the database, we need to update the note id with the new one (the one
		// returned by updating the title)
		ui.note.ID = noteID
	}
	if err = ui.noteService.UpdateNoteContent(ui.note); err != nil {
		return
	}

	return
}

func (ui *NoteDetailsWindowImpl) createFormWidget(w fyne.Window) fyne.CanvasObject {
	// widgets
	titleWidget := widget.NewEntry()
	titleWidget.SetPlaceHolder("Title")
	titleWidget.OnChanged = func(text string) {
		ui.note.Title = text
	}

	// create a markdown widget to display the content
	contentWidgetRichText := widget.NewRichText()
	contentWidgetRichText.Wrapping = fyne.TextWrapWord
	contentWidgetRichText.Scroll = container.ScrollVerticalOnly

	contentWidget := widget.NewMultiLineEntry()
	contentWidget.SetPlaceHolder("Note Content")
	contentWidget.Wrapping = fyne.TextWrapWord
	contentWidget.OnChanged = func(text string) {
		ui.note.Content = text
		contentWidgetRichText.ParseMarkdown(text)
	}

	hiddenCheckbox := widget.NewCheck("Hidden", func(checked bool) {
		ui.note.Hidden = checked
	})

	encryptedCheckbox := widget.NewCheck("Encrypted", func(encrypted bool) {
		// if checked and the note is not encrypted, encrypt it, else decrypt it
		if encrypted && !ui.note.Encrypted {
			if err := ui.noteService.EncryptNote(ui.note); err != nil {
				ui.ShowNotification("Error encrypting note", err.Error())
				return
			}
		} else if !encrypted && ui.note.Encrypted {
			if err := ui.noteService.DecryptNote(ui.note); err != nil {
				ui.ShowNotification("Error decrypting note", err.Error())
				return
			}
		}
		ui.note.Encrypted = encrypted
		ui.updateWidgetsData(ui.note)
	})

	createdAtWidget := widget.NewLabel("")
	updatedAtWidget := widget.NewLabel("")

	// form buttons: create all buttons and show only the ones that are needed
	btnSaveNew := widget.NewButton("Save", func() {
		if _, err := ui.saveNote(); err != nil {
			ui.ShowNotification("Error saving note", err.Error())
			return
		}
		ui.ShowNotification("Note created", "")
		ui.Close(true)
	})
	ui.AddWidget(common.BTN_SAVE_NEW, btnSaveNew)

	btnSaveUpdated := widget.NewButton("Save", func() {
		if _, err := ui.updatenote(); err != nil {
			ui.ShowNotification("Error saving updated note", err.Error())
			return
		}
		ui.ShowNotification("Note updated", "")
		ui.Close(true)
	})
	ui.AddWidget(common.BTN_SAVE_UPDATED, btnSaveUpdated)

	btnDelete := widget.NewButton("Delete", func() {
		// double check that note ID is set and = note.Title (encoded)
		computedID := ui.noteService.GetNoteIDFromTitle(ui.note.Title)
		if computedID != ui.note.ID {
			ui.ShowNotification(
				"Error comparing ids",
				"Note ID does not match note title. For safety I am not deleting this note!",
			)
			return
		}
		if err := ui.noteService.DeleteNote(ui.note.ID); err != nil {
			ui.ShowNotification("Error deleting note", err.Error())
			return
		}
		ui.ShowNotification("Note deleted", "")
		ui.Close(true)
	})
	ui.AddWidget(common.BTN_DELETE, btnDelete)

	btnCancel := widget.NewButton("Cancel", func() {
		ui.Close(true)
	})
	ui.AddWidget(common.BTN_CANCEL, btnCancel)

	btnOk := widget.NewButton("Ok", func() {
		ui.Close(true)
	})
	ui.AddWidget(common.BTN_OK, btnOk)

	btnCopyEncrypted := widget.NewButton("Copy Encrypted Note", func() {
		// to not temper with the original note, we create a copy
		tmpNote := ui.note
		// if note content is not encrypted, encrypt it and copy it to clipboard
		if !ui.note.Encrypted {
			if err := ui.noteService.EncryptNote(tmpNote); err != nil {
				ui.ShowNotification("Error encrypting note", err.Error())
				return
			}
		}
		clipboardContent := tmpNote.Title + ":\n" + tmpNote.Content
		w.Clipboard().SetContent(clipboardContent)
		ui.ShowNotification("Note content copied to system clipboard", w.Clipboard().Content())
	})
	ui.AddWidget(common.BTN_COPY_ENCRYPTED, btnCopyEncrypted)

	btnPasteEncrypted := widget.NewButton("Paste Encrypted Note Content", func() {
		// to not temper with the original note, we create a copy
		note := ui.note
		// to not temper with the original note, we create a copy
		note.Content = w.Clipboard().Content()
		// decrypt the note
		if err := ui.noteService.DecryptNote(note); err != nil {
			ui.ShowNotification("Error decrypting note", err.Error())
			return
		}
		ui.updateWidgetsData(note)
		ui.ShowNotification("Clipboard content decrypted and pasted into note", w.Clipboard().Content())
	})
	ui.AddWidget(common.BTN_PASTE_ENCRYPTED, btnPasteEncrypted)

	// create a button to toggle between the two content widgets
	btnToggleContent := widget.NewButton("Toggle Content", func() {
		if contentWidget.Visible() {
			contentWidget.Hide()
			contentWidgetRichText.Show()
		} else {
			contentWidgetRichText.Hide()
			contentWidget.Show()
		}
	})
	ui.AddWidget(common.BTN_TOGGLE_CONTENT, btnToggleContent)

	// adding widgets to widget map
	ui.AddWidget(common.WDG_NOTE_DETAILS_TITLE, titleWidget)
	ui.AddWidget(common.WDG_NOTE_DETAILS_CONTENT, contentWidget)
	ui.AddWidget(common.WDG_NOTE_DETAILS_CONTENT_RICH_TEXT, contentWidgetRichText)
	ui.AddWidget(common.WDG_NOTE_DETAILS_HIDDEN, hiddenCheckbox)
	ui.AddWidget(common.WDG_NOTE_DETAILS_ENCRYPTED, encryptedCheckbox)
	ui.AddWidget(common.WDG_NOTE_DETAILS_CREATED_AT, createdAtWidget)
	ui.AddWidget(common.WDG_NOTE_DETAILS_UPDATED_AT, updatedAtWidget)

	// adding widgets to bottom part of layout
	bottomContainer := container.NewVBox(
		hiddenCheckbox,
		createdAtWidget,
		updatedAtWidget,
		widget.NewSeparator(),
		container.NewHBox(
			btnToggleContent,
			btnSaveNew,
			btnSaveUpdated,
			btnDelete,
			btnCancel,
			btnOk,
		),
		container.NewHBox(
			btnCopyEncrypted,
			btnPasteEncrypted,
		),
	)

	// hide buttons that are not needed
	ui.setWidgetsStatus()

	// define the window container as a horizontal container
	noteDetailsContainer := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			titleWidget,
			bottomContainer,
			nil,
			nil,
		),
		titleWidget,
		bottomContainer,
		contentWidget,
		contentWidgetRichText,
	)

	return noteDetailsContainer
}

func (ui *NoteDetailsWindowImpl) setWidgetsStatus() {
	switch ui.windowMode {
	case common.WindowMode_View:
		ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, false)
		ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, false)
		ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, false)

		ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT, false)
		ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT_RICH_TEXT, true)
		ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, true)
		ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, true)
		ui.SetWidgetVisibility(common.BTN_SAVE_NEW, false)
		ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, false)
		ui.SetWidgetVisibility(common.BTN_DELETE, false)
		ui.SetWidgetVisibility(common.BTN_CANCEL, false)
		ui.SetWidgetVisibility(common.BTN_OK, true)
		ui.SetWidgetVisibility(common.BTN_COPY_ENCRYPTED, true)
		ui.SetWidgetVisibility(common.BTN_PASTE_ENCRYPTED, false)
		ui.SetWidgetVisibility(common.BTN_TOGGLE_CONTENT, false)
	case common.WindowMode_Edit:
		fallthrough
	default:
		switch ui.windowAction {
		case common.WindowAction_Update:
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, true)

			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT, false)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT_RICH_TEXT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, true)
			ui.SetWidgetVisibility(common.BTN_SAVE_NEW, false)
			ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, true)
			ui.SetWidgetVisibility(common.BTN_DELETE, true)
			ui.SetWidgetVisibility(common.BTN_CANCEL, true)
			ui.SetWidgetVisibility(common.BTN_OK, false)
			ui.SetWidgetVisibility(common.BTN_COPY_ENCRYPTED, true)
			ui.SetWidgetVisibility(common.BTN_PASTE_ENCRYPTED, true)
			ui.SetWidgetVisibility(common.BTN_TOGGLE_CONTENT, true)
		case common.WindowAction_Delete:
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, false)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, false)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, false)

			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT, false)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT_RICH_TEXT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, true)
			ui.SetWidgetVisibility(common.BTN_SAVE_NEW, false)
			ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, false)
			ui.SetWidgetVisibility(common.BTN_DELETE, true)
			ui.SetWidgetVisibility(common.BTN_CANCEL, true)
			ui.SetWidgetVisibility(common.BTN_OK, false)
			ui.SetWidgetVisibility(common.BTN_COPY_ENCRYPTED, false)
			ui.SetWidgetVisibility(common.BTN_PASTE_ENCRYPTED, false)
			ui.SetWidgetVisibility(common.BTN_TOGGLE_CONTENT, false)
		case common.WindowAction_New:
			fallthrough
		default:
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, true)

			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CONTENT_RICH_TEXT, false)
			// hide updated and created fields if new note
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, false)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, false)
			ui.SetWidgetVisibility(common.BTN_SAVE_NEW, true)
			ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, false)
			ui.SetWidgetVisibility(common.BTN_DELETE, false)
			ui.SetWidgetVisibility(common.BTN_CANCEL, true)
			ui.SetWidgetVisibility(common.BTN_OK, false)
			ui.SetWidgetVisibility(common.BTN_COPY_ENCRYPTED, true)
			ui.SetWidgetVisibility(common.BTN_PASTE_ENCRYPTED, true)
			ui.SetWidgetVisibility(common.BTN_TOGGLE_CONTENT, true)
		}
	}
}
