package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/iltoga/ecnotes-go/lib/common"
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
	SaveEncryptedNotes(notes []model.Note) error
	ReEncryptNotes(notes []model.Note, cert model.EncKey) error
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
	Crypto        CryptoServiceFactory
	// Titles an array with all note Titles in db
	Titles []string
}

// NewNoteService ....
func NewNoteService(
	noteRepo NoteServiceRepository,
	configService ConfigService,
	observer observer.Observer,
	crypto CryptoServiceFactory,
) NoteService {
	return &NoteServiceImpl{
		NoteRepo:      noteRepo,
		ConfigService: configService,
		Observer:      observer,
		Crypto:        crypto,
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
	if len(ns.Titles) == 0 {
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
func (ns *NoteServiceImpl) SaveEncryptedNotes(notes []model.Note) error {
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

// ReEncryptNotes re-encrypts a batch of notes with a given key and encryption algorithm
func (ns *NoteServiceImpl) ReEncryptNotes(notes []model.Note, cert model.EncKey) error {
	oldSrv := ns.Crypto.GetSrv()
	if oldSrv == nil {
		return errors.New(common.ERR_NO_KEY)
	}

	newSrv := NewCryptoServiceFactory(cert.Algo)
	if newSrv == nil {
		return fmt.Errorf("unsupported encryption algorithm: %q", cert.Algo)
	}
	if err := newSrv.GetKeyManager().ImportKey(cert.Key, cert.Name); err != nil {
		return err
	}

	ns.Crypto.SetSrv(newSrv)
	for _, note := range notes {
		noteCopy := note
		if noteCopy.Encrypted {
			encryptedContent, err := hex.DecodeString(noteCopy.Content)
			if err != nil {
				return err
			}
			decryptedContent, err := oldSrv.Decrypt(encryptedContent)
			if err != nil {
				return err
			}
			noteCopy.Content = string(decryptedContent)
			noteCopy.Encrypted = false
		}
		if err := ns.UpdateNoteContent(&noteCopy); err != nil {
			return err
		}
	}
	_, err := ns.GetNotes()
	if err != nil {
		return err
	}
	return nil
}

// processAndSave centralizes decryption, snapshotting, encryption, and repo saving
func (ns *NoteServiceImpl) processAndSave(note *model.Note, action func(*model.Note) error) (savedNote *model.Note, decNote *model.Note, err error) {
	noteCopy := *note
	if noteCopy.Encrypted {
		if err := ns.DecryptNote(&noteCopy); err != nil {
			return nil, nil, err
		}
	}
	decNoteCopy := noteCopy
	if err := ns.EncryptNote(&noteCopy); err != nil {
		return nil, nil, err
	}
	if err := action(&noteCopy); err != nil {
		return nil, nil, err
	}
	savedNoteCopy := noteCopy
	return &savedNoteCopy, &decNoteCopy, nil
}

// emitNoteChanged triggers the notification for the UI observer
func (ns *NoteServiceImpl) emitNoteChanged(event observer.Event, decNote *model.Note, savedNote *model.Note) {
	ns.Observer.Notify(event, decNote, common.WindowMode_Edit, common.WindowAction_Update, savedNote)
}

// CreateNote adds a new note to the repo
func (ns *NoteServiceImpl) CreateNote(note *model.Note) error {
	if note.Title == "" || note.Content == "" {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	if note.ID == 0 {
		note.ID = ns.NoteRepo.GetIDFromTitle(note.Title)
	}
	if exists, _ := ns.NoteRepo.NoteExists(note.ID); exists {
		return errors.New(common.ERR_NOTE_ALREADY_EXISTS)
	}
	note.CreatedAt = common.GetCurrentTimestamp()
	note.UpdatedAt = common.GetCurrentTimestamp()

	savedNote, decNote, err := ns.processAndSave(note, ns.NoteRepo.CreateNote)
	if err == nil {
		ns.Titles = append(ns.Titles, savedNote.Title)
		ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, ns.Titles)
		ns.emitNoteChanged(observer.EVENT_CREATE_NOTE, decNote, savedNote)
	}
	return err
}

// UpdateNoteContent update the content of an existing note
func (ns *NoteServiceImpl) UpdateNoteContent(note *model.Note) error {
	if note.Title == "" || note.Content == "" || note.ID == 0 {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	if ok, err := ns.NoteRepo.NoteExists(note.ID); err != nil {
		return err
	} else if !ok {
		return errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	note.UpdatedAt = common.GetCurrentTimestamp()

	savedNote, decNote, err := ns.processAndSave(note, ns.NoteRepo.UpdateNote)
	if err == nil {
		ns.emitNoteChanged(observer.EVENT_UPDATE_NOTE, decNote, savedNote)
	}
	return err
}

// UpdateNoteTitle update the title of an existing UpdateNote
func (ns *NoteServiceImpl) UpdateNoteTitle(oldTitle, newTitle string) (noteID int, err error) {
	oldIndex := ns.NoteRepo.GetIDFromTitle(oldTitle)
	noteID = oldIndex
	if newTitle == "" {
		err = errors.New(common.ERR_NOTE_TITLE_EMPTY)
		return
	}
	if oldTitle == "" || newTitle == oldTitle {
		return
	}

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

	_, _, err = ns.processAndSave(note, func(n *model.Note) error {
		return ns.NoteRepo.RenameNote(oldIndex, n)
	})
	if err != nil {
		return noteID, err
	}
	noteID = newIndex

	// update titles array
	for i, title := range ns.Titles {
		if title == oldTitle {
			ns.Titles[i] = newTitle
			break
		}
	}
	ns.Observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, ns.Titles)
	return noteID, nil
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
	if err = ns.NoteRepo.DeleteNote(id); err != nil {
		return err
	}

	// remove the note from the titles array only after the repo delete succeeds
	for i, title := range ns.Titles {
		if title == note.Title {
			ns.Titles = append(ns.Titles[:i], ns.Titles[i+1:]...)
			break
		}
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
	note.EncKeyName = ns.Crypto.GetSrv().GetKeyManager().GetCertificate().Name
	encryptedContent, err := ns.Crypto.GetSrv().Encrypt([]byte(note.Content))
	if err != nil {
		return err
	}
	// hex encode the encrypted content
	note.Content = hex.EncodeToString(encryptedContent)
	note.Encrypted = true
	return nil
}

// DecryptNote ....
func (ns *NoteServiceImpl) DecryptNote(note *model.Note) error {
	// make sure the note is not empty
	if note == nil || note.Title == "" || note.Content == "" {
		return errors.New(common.ERR_NOTE_EMPTY)
	}
	// hex decode the encrypted content
	encryptedContent, err := hex.DecodeString(note.Content)
	if err != nil {
		return err
	}
	decryptedContent, err := ns.Crypto.GetSrv().Decrypt(encryptedContent)
	if err != nil {
		return err
	}
	note.Content = string(decryptedContent)
	note.Encrypted = false
	return nil
}
