package service

import (
	"errors"
	"sort"
	"sync"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// NoteService ....
type NoteService interface {
	GetNote(id int) (Note, error)
	GetNotes() ([]Note, error)
	SearchNotes(query string, fuzzySearch bool) (map[string]int, error)
	CreateNote(note Note) error
	UpdateNote(note Note) error
	DeleteNote(id int) error
	EncryptNote(note *Note) error
	DecryptNote(note *Note) error
}

// NoteServiceImpl ....
type NoteServiceImpl struct {
	NoteRepo      NoteServiceRepository
	ConfigService ConfigService
	// titles a map representing all notes in db (key: note title hash, value: note ID)
	TitlesIDMap map[string]int
	// Titles an array with all note Titles in db
	Titles        []string
	TitlesIDMutex *sync.RWMutex
}

// Note ....
type Note struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// NewNoteService ....
func NewNoteService(
	noteRepo NoteServiceRepository,
	configService ConfigService,
) NoteService {
	return &NoteServiceImpl{
		NoteRepo:      noteRepo,
		ConfigService: configService,
		TitlesIDMap:   make(map[string]int),
		Titles:        make([]string, 0),
		TitlesIDMutex: &sync.RWMutex{},
	}
}

// GetNote ....
func (ns *NoteServiceImpl) GetNote(id int) (Note, error) {
	return ns.NoteRepo.GetNote(id)
}

// GetNotes ....
func (ns *NoteServiceImpl) GetNotes() ([]Note, error) {
	notes, err := ns.NoteRepo.GetAllNotes()
	if err != nil {
		return nil, err
	}
	ns.TitlesIDMutex.Lock()
	defer ns.TitlesIDMutex.Unlock()
	ns.Titles = make([]string, 0)
	ns.TitlesIDMap = make(map[string]int)
	for _, note := range notes {
		ns.Titles = append(ns.Titles, note.Title)
		titleHash := string(cryptoUtil.Hash(note.Title))
		ns.TitlesIDMap[titleHash] = note.ID
	}
	return notes, nil
}

// SearchNotes ....
func (ns *NoteServiceImpl) SearchNotes(query string, fuzzySearch bool) (map[string]int, error) {
	// get all notes if Titles is empty
	if ns.Titles == nil {
		_, err := ns.GetNotes()
		if err != nil {
			return nil, err
		}
	}
	// search the titles array and return the IDs of the notes that match the query
	if fuzzySearch {
		return ns.searchFuzzy(query)
	}
	return ns.searchExact(query)
}

func (ns *NoteServiceImpl) searchFuzzy(query string) (map[string]int, error) {
	matches := fuzzy.RankFind(query, ns.Titles)
	sort.Sort(matches)
	// for every result calculate the hash of the title and get the corresponding keys of titlesIDMap and return a subset of the titlesIDMap with the matching keys
	result := make(map[string]int)
	for _, match := range matches {
		result[string(cryptoUtil.Hash(match.Target))] = ns.TitlesIDMap[string(cryptoUtil.Hash(match.Target))]
	}
	return result, nil
}

func (ns *NoteServiceImpl) searchExact(query string) (map[string]int, error) {
	hashedTitle := string(cryptoUtil.Hash(query))
	if _, ok := ns.TitlesIDMap[hashedTitle]; ok {
		return map[string]int{hashedTitle: ns.TitlesIDMap[hashedTitle]}, nil
	}
	return map[string]int{}, nil
}

// CreateNote ....
func (ns *NoteServiceImpl) CreateNote(note Note) error {
	return ns.NoteRepo.CreateNote(note)
}

// UpdateNote ....
func (ns *NoteServiceImpl) UpdateNote(note Note) error {
	return ns.NoteRepo.UpdateNote(note)
}

// DeleteNote ....
func (ns *NoteServiceImpl) DeleteNote(id int) error {
	return ns.NoteRepo.DeleteNote(id)
}

// EncryptNote ....
func (ns *NoteServiceImpl) EncryptNote(note *Note) error {
	// make sure the note is not empty
	if note.Title == "" || note.Content == "" {
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
	note.Content = string(encryptedContent)
	return nil
}

// DecryptNote ....
func (ns *NoteServiceImpl) DecryptNote(note *Note) error {
	// make sure the note is not empty
	if note.Title == "" || note.Content == "" {
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
	note.Content = string(decryptedContent)
	return nil
}

func (ns *NoteServiceImpl) getEncryptionKey() (string, error) {
	// encKey is the encryption key in clear text (is the passphrase generated on first run)
	return ns.ConfigService.GetGlobal(common.CONFIG_ENCRYPTION_KEY)
}
