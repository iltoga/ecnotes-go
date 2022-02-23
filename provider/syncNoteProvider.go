// Interfaces to be implemented by different providers
package provider

import "github.com/iltoga/ecnotes-go/model"

// SyncNoteProvider is the interface that must be implemented by a sync-note provider implementation
// Note: the relative service must be able to get/put/delete/find notes from the provider
type SyncNoteProvider interface {
	// GetNoteIDs returns the list of note IDs from the provider
	GetNoteIDs() ([]int, error)
	// GetNote returns the note with the given id
	GetNote(id int) (*model.Note, error)
	// PutNote puts the given note into the provider
	PutNote(note *model.Note) error
	// DeleteNote deletes the note with the given id
	DeleteNote(id int) error
	// FindNotes returns all notes that match the given query
	FindNotes(query string) ([]model.Note, error)
	// SyncNotes syncs the notes from the provider to the local database and vice versa
	SyncNotes() error
	// Init initializes the provider
	Init() error
}

// BaseSyncNoteProvider ....
type BaseSyncNoteProvider struct {
	// NoteIDs is the list of note IDs from the provider
	NoteIDs []int
	// Notes is the list of notes from the provider
	Notes []model.Note
}
