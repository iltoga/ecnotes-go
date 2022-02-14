package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
)

// NoteDetailsWindow ....
type NoteDetailsWindow interface {
	WindowInterface
	UpdateNoteDetailsWidget() observer.Listener
}

// NoteDetailsWindowImpl ....
type NoteDetailsWindowImpl struct {
	UImpl
	note *service.Note
}

// NewNoteDetailsWindow ....
func NewNoteDetailsWindow(ui *UImpl, note *service.Note) NoteDetailsWindow {
	return &NoteDetailsWindowImpl{
		UImpl: *ui,
		note:  note,
	}
}

// CreateWindow ....
func (ui *NoteDetailsWindowImpl) CreateWindow(title string, width, height float32, visible bool) {
	// define main windows
	w := ui.app.NewWindow(title)
	ui.AddWindow("note_details", w)
	w.Resize(fyne.NewSize(width, height))
	w.SetContent(ui.createFormWidget())
	w.CenterOnScreen()
	if visible {
		w.Show()
		if titleWidget, err := ui.GetWidget("note_details_title"); err == nil {
			ui.SetFocusOnWidget(w, titleWidget.(*widget.Entry))
		} else {
			log.Printf("Error getting title widget: %v", err)
		}
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
			ui.note = n
			if w, err := ui.GetWidget("note_details_title"); err == nil {
				w.(*widget.Entry).SetText(n.Title)
			} else {
				log.Printf("Error getting title widget: %v", err)
			}

			if w, err := ui.GetWidget("note_details_content"); err == nil {
				w.(*widget.Entry).SetText(n.Content)
			} else {
				log.Printf("Error getting content widget: %v", err)
			}

			if w, err := ui.GetWidget("note_details_hidden"); err == nil {
				w.(*widget.Check).SetChecked(n.Hidden)
			} else {
				log.Printf("Error getting hidden widget: %v", err)
			}

			createdAtStr := common.FormatTime(common.TimestampToTime(n.CreatedAt))
			if w, err := ui.GetWidget("note_details_created_at"); err == nil {
				w.(*widget.Entry).SetText(createdAtStr)
			} else {
				log.Printf("Error getting createdAt widget: %v", err)
			}

			updatedAtStr := common.FormatTime(common.TimestampToTime(n.UpdatedAt))
			if w, err := ui.GetWidget("note_details_updated_at"); err == nil {
				w.(*widget.Entry).SetText(updatedAtStr)
			} else {
				log.Printf("Error getting updatedAt widget: %v", err)
			}
		},
	}
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

	// TODO: move names to constants
	// adding widgets to widget map
	ui.AddWidget("note_details_title", titleWidget)
	ui.AddWidget("note_details_content", contentWidget)
	ui.AddWidget("note_details_hidden", hiddenCheckbox)
	ui.AddWidget("note_details_created_at", createdAtWidget)
	ui.AddWidget("note_details_updated_at", updatedAtWidget)

	// define content
	noteDetails := widget.NewForm()
	noteDetails.Append("Title", titleWidget)
	noteDetails.Append("Content", contentWidget)
	noteDetails.Append("Hidden", hiddenCheckbox)
	noteDetails.Append("Created", createdAtWidget)
	noteDetails.Append("Updated", updatedAtWidget)

	return noteDetails
}
