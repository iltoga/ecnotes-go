package service

import (
	"errors"
	"sort"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// NoteService ....
type NoteService interface {
	GetNoteWithContent(id int) (*model.Note, error)
	GetNotes() ([]model.Note, error)
	GetTitles() []string
	SearchNotes(query string, fuzzySearch bool) ([]string, error)
	CreateNote(note *model.Note) error
	CreateEncryptedNotes(notes []model.Note) error
	UpdateNoteContent(note *model.Note) error

	UpdateNoteTitle(oldTitle, newTitle string) (noteID int, err error)
	DeleteNote(id int) error
	EncryptNote(note *model.Note) error
	DecryptNote(note *model.Note) error
	GetNoteIDFromTitle(title string) int
}

// NoteServiceImpl ....
type NoteServiceImpl struct {
	NoteRepo      NoteServiceRepository
	ConfigService ConfigService
	Observer      observer.Observer
	// Titles an array with all note Titles in db
	Titles []string
}

// NewNoteService ....
func NewNoteService(
	noteRepo NoteServiceRepository,
	configService ConfigService,
	observer observer.Observer,
) NoteService {
	return &NoteServiceImpl{
		NoteRepo:      noteRepo,
		ConfigService: configService,
		Observer:      observer,
		Titles:        []string{},
	}
}

// GetNoteIDFromTitle returns the note ID from the title
func (ns *NoteServiceImpl) GetNoteIDFromTitle(title string) int {
	return ns.NoteRepo.GetIDFromTitle(title)
}

// GetNote retreives a note from the db by id and decrypts it
func (ns *NoteServiceImpl) GetNoteWithContent(id int) (*model.Note, error) {
	note, err := ns.NoteRepo.GetNote(id)
	if err != nil {
		return nil, err
	}
	// decrypt content before returning
	if err := ns.DecryptNote(note); err != nil {
		return nil, err
	}
	return note, nil
}

// GetNotes returns all note titles from the db and populate Titles array and TitlesIDMap with the results
// note: the note content is returned encrypted
func (ns *NoteServiceImpl) GetNotes() ([]model.Note, error) {
	notes, err := ns.NoteRepo.GetAllNotes()
	if err != nil {
		// on empty bucket don't return an error
		if err.Error() == common.ERR_BUCKET_EMPTY {
			notes = []model.Note{}
		} else {
			return nil, err
		}
	}
	ns.Titles = []string{}
	for idx, note := range notes {
		ns.Titles = append(ns.Titles, note.Title)
		// swap the note with the decrypted one
		notes[idx] = note
	}
	// emit a note titles' update event
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, ns.Titles)
	return notes, nil
}

// GetTitles returns all note titles from memory
func (ns *NoteServiceImpl) GetTitles() []string {
	return ns.Titles
}

// SearchNotes ....
func (ns *NoteServiceImpl) SearchNotes(query string, fuzzySearch bool) ([]string, error) {
	// get all notes if Titles is empty
	if ns.Titles == nil || len(ns.Titles) == 0 {
		_, err := ns.GetNotes()
		if err != nil {
			return nil, err
		}
	}
	// search the titles array and return the IDs of the notes that match the query
	var searchResult []string
	if fuzzySearch {
		searchResult = ns.searchFuzzy(query, ns.Titles)
	} else {
		searchResult = ns.searchExact(query, ns.Titles)
	}
	// emit a note titles' update event
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, searchResult)
	return searchResult, nil
}

func (ns *NoteServiceImpl) searchFuzzy(query string, titles []string) []string {
	// case-insensitive search
	matches := fuzzy.RankFindFold(query, titles)
	sort.Sort(matches)
	// for every result calculate the hash of the title and get the corresponding keys of titlesIDMap and return a subset of the titlesIDMap with the matching keys
	result := []string{}
	for _, match := range matches {
		result = append(result, match.Target)
	}
	return result
}

func (ns *NoteServiceImpl) searchExact(query string, titles []string) []string {
	for i, title := range titles {
		if title == query {
			return []string{titles[i]}
		}
	}
	return []string{}
}

// CreateEncryptedNotes save to db a batch of (already) encrypted notes
// TODO: refactor this method to use a batch insert	instead of a loop
func (ns *NoteServiceImpl) CreateEncryptedNotes(notes []model.Note) error {
	// loop through the notes and save them to db
	for _, note := range notes {
		if err := ns.NoteRepo.CreateNote(&note); err != nil {
			return err
		}
	}
	// get all note titles from db and notify the observer
	_, err := ns.GetNotes()
	if err != nil {
		return err
	}
	return nil
}

// CreateNote ....
func (ns *NoteServiceImpl) CreateNote(note *model.Note) error {
	if note.Title == "" || note.Content == "" {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	// generate a new note ID if the note has no ID
	if note.ID == 0 {
		note.ID = ns.NoteRepo.GetIDFromTitle(note.Title)
	}
	// make sure the note ID is unique
	if exists, _ := ns.NoteRepo.NoteExists(note.ID); exists {
		return errors.New(common.ERR_NOTE_ALREADY_EXISTS)
	}
	note.CreatedAt = common.GetCurrentTimestamp()
	note.UpdatedAt = common.GetCurrentTimestamp()
	// add note title to titles array
	ns.Titles = append(ns.Titles, note.Title)
	// emit a note titles' update event that will update the note titles in the UI (main window)
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, ns.Titles)

	// if note is encrypted, decrypt content before saving to db, to be able to present unencrypted content in the UI
	// after saving the note to db
	if note.Encrypted {
		if err := ns.DecryptNote(note); err != nil {
			return err
		}
	}
	oldContent := note.Content
	if err := ns.EncryptNote(note); err != nil {
		return err
	}
	err := ns.NoteRepo.CreateNote(note)
	// notify both the unencrypted and the encrypted note
	decNote := *note
	decNote.Content = oldContent
	decNote.Encrypted = false
	// emit a note created event that will update the note details in the UI (note details window)
	ns.Observer.Notify(observer.EVENT_CREATE_NOTE, decNote, common.WindowMode_Edit, common.WindowAction_Update, note)
	return err
}

// UpdateNoteContent update the content of an existing note
func (ns *NoteServiceImpl) UpdateNoteContent(note *model.Note) error {
	if note.Title == "" || note.Content == "" || note.ID == 0 {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	// make sure the note already exists since we are updating the content
	if ok, err := ns.NoteRepo.NoteExists(note.ID); err != nil {
		return err
	} else if !ok {
		return errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	note.UpdatedAt = common.GetCurrentTimestamp()
	// if note is encrypted, decrypt content before saving to db, to be able to present unencrypted content in the UI
	// after saving the note to db
	if note.Encrypted {
		if err := ns.DecryptNote(note); err != nil {
			return err
		}
	}
	oldContent := note.Content
	if err := ns.EncryptNote(note); err != nil {
		return err
	}
	err := ns.NoteRepo.UpdateNote(note)
	// TODO: maybe refactor Crete/Update note to encrypt content but return the original (unencrypted) content
	// this is to avoid showing the note content encrypted in the UI
	decNote := *note
	decNote.Content = oldContent
	decNote.Encrypted = false

	// emit a note created event that will update the note details in the UI (note details window)
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE, decNote, common.WindowMode_Edit, common.WindowAction_Update, note)
	return err
}

// UpdateNoteTitle update the title of an existing UpdateNote
// since the title is used as a key in db we need to delete the old note and create a new one
// the new note is a copy of the old note with the new title
func (ns *NoteServiceImpl) UpdateNoteTitle(oldTitle, newTitle string) (noteID int, err error) {
	oldIndex := ns.NoteRepo.GetIDFromTitle(oldTitle)
	noteID = oldIndex
	if newTitle == "" {
		err = errors.New(common.ERR_NOTE_TITLE_EMPTY)
		return
	}
	// if oldTitle is empty or same as newTitle, it means that title hasn't been changed.	so we can just update the note content
	if oldTitle == "" || newTitle == oldTitle {
		return
	}

	// make sure the note already exists since we are updating the content
	var ok bool
	if ok, err = ns.NoteRepo.NoteExists(oldIndex); err != nil {
		return
	} else if !ok {
		err = errors.New(common.ERR_NOTE_NOT_FOUND)
		return
	}
	var note *model.Note
	note, err = ns.NoteRepo.GetNote(oldIndex)
	if err != nil {
		return
	}
	note.Title = newTitle
	newIndex := ns.NoteRepo.GetIDFromTitle(newTitle)
	note.ID = newIndex
	note.UpdatedAt = common.GetCurrentTimestamp()

	// TODO: wrap this block in a transaction
	if err = ns.NoteRepo.DeleteNote(oldIndex); err != nil {
		return
	}
	// if note is encrypted, decrypt content before saving to db, to be able to present unencrypted content in the UI
	// after saving the note to db
	if note.Encrypted {
		if err = ns.DecryptNote(note); err != nil {
			return
		}
	}
	oldContent := note.Content
	if err = ns.EncryptNote(note); err != nil {
		return
	}
	if err = ns.NoteRepo.CreateNote(note); err != nil {
		return
	}
	noteID = newIndex
	// always return unencrypted content after saving
	note.Content = oldContent
	note.Encrypted = false

	// update titles array
	for i, title := range ns.Titles {
		if title == oldTitle {
			ns.Titles[i] = newTitle
			break
		}
	}
	// emit a note titles' update event
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, ns.Titles)
	// Note: this time no need to emit a note updated event since everytime we update a note title, we also update the note content
	return
}

// DeleteNote ....
func (ns *NoteServiceImpl) DeleteNote(id int) error {
	if ok, err := ns.NoteRepo.NoteExists(id); err != nil {
		return err
	} else if !ok {
		return errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	note, err := ns.NoteRepo.GetNote(id)
	if err != nil {
		return err
	}
	// remove the note from the titles array
	for i, title := range ns.Titles {
		if title == note.Title {
			ns.Titles = append(ns.Titles[:i], ns.Titles[i+1:]...)
			break
		}
	}
	if err = ns.NoteRepo.DeleteNote(id); err != nil {
		return err
	}

	// emit a note titles' update event
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, ns.Titles)
	ns.Observer.Notify(observer.EVENT_DELETE_NOTE, note, common.WindowMode_Edit, common.WindowAction_Update)
	// Note: no need to emit a note update/delete event. since we are deleting a note, we don't need to update the note details in the UI, but just clear the data and hide the note details window
	return nil
}

// EncryptNote ....
func (ns *NoteServiceImpl) EncryptNote(note *model.Note) error {
	// make sure the note is not empty
	if note == nil || note.Title == "" || note.Content == "" {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	// make sure we have the encryption key
	encryptionKey, err := ns.getEncryptionKey()
	if err != nil {
		return err
	}
	encryptedContent, err := cryptoUtil.EncryptMessage(note.Content, encryptionKey)
	if err != nil {
		return err
	}
	note.Content = encryptedContent
	note.Encrypted = true
	return nil
}

// DecryptNote ....
func (ns *NoteServiceImpl) DecryptNote(note *model.Note) error {
	// make sure the note is not empty
	if note == nil || note.Title == "" || note.Content == "" {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	// make sure we have the encryption key
	encryptionKey, err := ns.getEncryptionKey()
	if err != nil {
		return err
	}
	decryptedContent, err := cryptoUtil.DecryptMessage(note.Content, encryptionKey)
	if err != nil {
		return err
	}
	note.Content = decryptedContent
	note.Encrypted = false
	return nil
}

func (ns *NoteServiceImpl) getEncryptionKey() (string, error) {
	// encKey is the encryption key in clear text (is the passphrase generated on first run)
	return ns.ConfigService.GetGlobal(common.CONFIG_ENCRYPTION_KEY)
}
