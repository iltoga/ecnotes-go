package ui

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
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
	note     *service.Note
	oldTitle string // in case we update the note title we need to save the old one to be able to save the note
}

// NewNoteDetailsWindow ....
func NewNoteDetailsWindow(ui *UImpl, note *service.Note) NoteDetailsWindow {
	return &NoteDetailsWindowImpl{
		UImpl: *ui,
		note:  note,
	}
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
	if ui.windowAspect == common.WindowAspect_FullScreen {
		w.SetFullScreen(true)
	} else {
		w.Resize(fyne.NewSize(width, height))
	}

	w.SetContent(ui.createFormWidget())
	w.CenterOnScreen()
	// TODO: find a more elegant way to recreate the window when it is closed
	// this is to avoid the note details window to be destroyed when the user closes it
	w.SetOnClosed(func() {
		go func() {
			time.Sleep(time.Millisecond * 500)
			ui.CreateWindow(title, width, height, visible, options)
			ui.SetWindowVisibility(common.WIN_NOTE_DETAILS, false)
		}()
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

func (ui *NoteDetailsWindowImpl) updateNoteData(n *service.Note) {
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

	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_HIDDEN); err == nil {
		w.(*widget.Check).SetChecked(n.Hidden)
	} else {
		log.Printf("Error getting hidden widget: %v", err)
	}

	createdAtStr := common.FormatTime(common.TimestampToTime(n.CreatedAt))
	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_CREATED_AT); err == nil {
		w.(*widget.Entry).SetText(createdAtStr)
	} else {
		log.Printf("Error getting createdAt widget: %v", err)
	}

	updatedAtStr := common.FormatTime(common.TimestampToTime(n.UpdatedAt))
	if w, err := ui.GetWidget(common.WDG_NOTE_DETAILS_UPDATED_AT); err == nil {
		w.(*widget.Entry).SetText(updatedAtStr)
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
			n, ok := note.(*service.Note)
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
			ui.updateNoteData(n)
		},
	}
}

// Close close note details window
func (ui *NoteDetailsWindowImpl) Close(clearData bool) {
	if clearData {
		ui.updateNoteData(new(service.Note))
	}
	w, err := ui.GetWindow(common.WIN_NOTE_DETAILS)
	if err != nil {
		ui.ShowNotification("Error getting window instance", err.Error())
	}
	w.Hide()
}

func (ui *NoteDetailsWindowImpl) createFormWidget() fyne.CanvasObject {
	// widgets
	titleWidget := widget.NewEntry()
	titleWidget.OnChanged = func(text string) {
		ui.note.Title = text
	}

	contentWidget := widget.NewMultiLineEntry()
	contentWidget.Wrapping = fyne.TextWrapWord
	contentWidget.OnChanged = func(text string) {
		ui.note.Content = text
	}

	hiddenCheckbox := widget.NewCheck("Hidden", func(checked bool) {
		ui.note.Hidden = checked
	})

	createdAtWidget := widget.NewEntry()
	createdAtWidget.Disable()

	updatedAtWidget := widget.NewEntry()
	updatedAtWidget.Disable()

	// form buttons: create all buttons and show only the ones that are needed
	btnSaveNew := widget.NewButton("Save", func() {
		unEncContent := ui.note.Content
		if err := ui.noteService.CreateNote(ui.note); err != nil {
			ui.ShowNotification("Error saving new note", err.Error())
			return
		}
		// this is to allow the user to see the content unencrypted in the text editor
		ui.note.Content = unEncContent
		ui.ShowNotification("Note created", "")
	})
	ui.AddWidget(common.BTN_SAVE_NEW, btnSaveNew)
	btnSaveUpdated := widget.NewButton("Save", func() {
		var (
			noteID int
			err    error
		)
		// try update the title, if changed
		if ui.note.Title != ui.oldTitle && ui.note.Title != "" && ui.oldTitle != "" {
			if noteID, err = ui.noteService.UpdateNoteTitle(ui.oldTitle, ui.note.Title); err != nil {
				ui.ShowNotification("Error updating note title", err.Error())
				return
			}
		}
		unEncContent := ui.note.Content
		ui.note.ID = noteID
		if err := ui.noteService.UpdateNoteContent(ui.note); err != nil {
			ui.ShowNotification("Error updating note content", err.Error())
			return
		}
		// this is to allow the user to see the content unencrypted in the text editor
		ui.note.Content = unEncContent
		ui.ShowNotification("Note updated", "")
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

	// TODO: move names to constants
	// adding widgets to widget map
	ui.AddWidget(common.WDG_NOTE_DETAILS_TITLE, titleWidget)
	ui.AddWidget(common.WDG_NOTE_DETAILS_CONTENT, contentWidget)
	ui.AddWidget(common.WDG_NOTE_DETAILS_HIDDEN, hiddenCheckbox)
	ui.AddWidget(common.WDG_NOTE_DETAILS_CREATED_AT, createdAtWidget)
	ui.AddWidget(common.WDG_NOTE_DETAILS_UPDATED_AT, updatedAtWidget)

	// define content
	noteDetails := widget.NewForm()
	noteDetails.Append("Title", titleWidget)
	noteDetails.Append("Content", contentWidget)
	noteDetails.Append("Hidden", hiddenCheckbox)
	noteDetails.Append("Created", createdAtWidget)
	noteDetails.Append("Updated", updatedAtWidget)

	// hide buttons that are not needed

	btnBar := container.NewHBox(btnCancel, btnSaveNew, btnSaveUpdated, btnDelete, btnOk)
	ui.setWidgetsStatus()
	noteDetails.Append("", btnBar)
	return noteDetails
}

func (ui *NoteDetailsWindowImpl) setWidgetsStatus() {
	switch ui.windowMode {
	case common.WindowMode_View:
		ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, false)
		ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, false)
		ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, false)

		ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, true)
		ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, true)
		ui.SetWidgetVisibility(common.BTN_SAVE_NEW, false)
		ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, false)
		ui.SetWidgetVisibility(common.BTN_DELETE, false)
		ui.SetWidgetVisibility(common.BTN_CANCEL, false)
		ui.SetWidgetVisibility(common.BTN_OK, true)
	case common.WindowMode_Edit:
		fallthrough
	default:
		switch ui.windowAction {
		case common.WindowAction_Update:
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, true)

			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, true)
			ui.SetWidgetVisibility(common.BTN_SAVE_NEW, false)
			ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, true)
			ui.SetWidgetVisibility(common.BTN_DELETE, true)
			ui.SetWidgetVisibility(common.BTN_CANCEL, true)
			ui.SetWidgetVisibility(common.BTN_OK, false)
		case common.WindowAction_Delete:
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, true)

			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, true)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, true)
			ui.SetWidgetVisibility(common.BTN_SAVE_NEW, false)
			ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, false)
			ui.SetWidgetVisibility(common.BTN_DELETE, true)
			ui.SetWidgetVisibility(common.BTN_CANCEL, true)
			ui.SetWidgetVisibility(common.BTN_OK, false)
		case common.WindowAction_New:
			fallthrough
		default:
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_TITLE, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_CONTENT, true)
			ui.SetWidgetEnabled(common.WDG_NOTE_DETAILS_HIDDEN, true)

			// hide updated and created fields if new note
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_CREATED_AT, false)
			ui.SetWidgetVisibility(common.WDG_NOTE_DETAILS_UPDATED_AT, false)
			ui.SetWidgetVisibility(common.BTN_SAVE_NEW, true)
			ui.SetWidgetVisibility(common.BTN_SAVE_UPDATED, false)
			ui.SetWidgetVisibility(common.BTN_DELETE, false)
			ui.SetWidgetVisibility(common.BTN_CANCEL, true)
			ui.SetWidgetVisibility(common.BTN_OK, false)
		}
	}
}
