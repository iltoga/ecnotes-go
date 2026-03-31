package provider

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service/observer"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleProvider struct {
	BaseSyncNoteProvider
	sheetsService  *sheets.Service
	client         *http.Client
	sheetName      string
	sheetID        string
	credFilePath   string
	noteIds        map[int]int
	notesUpdatedAt map[int]int64
	idsMux         *sync.RWMutex
	updAtMux       *sync.RWMutex
	updateQueue    chan *model.Note
	deleteQueue    chan int
	ctx            context.Context
	logger         *log.Logger
	observer       observer.Observer
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(
	sheetName string,
	sheetID string,
	credFilePath string,
	logger *log.Logger,
	observer observer.Observer,
) (*GoogleProvider, error) {
	gp := &GoogleProvider{
		sheetName:      sheetName,
		sheetID:        sheetID,
		credFilePath:   credFilePath,
		noteIds:        make(map[int]int),
		notesUpdatedAt: make(map[int]int64),
		idsMux:         &sync.RWMutex{},
		updAtMux:       &sync.RWMutex{},
		updateQueue:    make(chan *model.Note, 100),
		deleteQueue:    make(chan int, 100),
		logger:         logger,
		observer:       observer,
	}
	if err := gp.Init(); err != nil {
		return nil, err
	}
	return gp, nil
}

// CacheIDSet update the note ID map
func (gp *GoogleProvider) CacheIDSet(noteID int, noteIDx int, nonBlocked bool) {
	if !nonBlocked {
		gp.idsMux.Lock()
		defer gp.idsMux.Unlock()
	}
	gp.noteIds[noteID] = noteIDx
}

// CacheIDGet returns the note ID from the cache
func (gp *GoogleProvider) CacheIDGet(noteID int) (int, bool) {
	gp.idsMux.RLock()
	defer gp.idsMux.RUnlock()
	idx, ok := gp.noteIds[noteID]
	return idx, ok
}

// CacheIDUnset removes the note ID from the cache
func (gp *GoogleProvider) CacheIDUnset(noteID int) {
	gp.idsMux.Lock()
	defer gp.idsMux.Unlock()
	delete(gp.noteIds, noteID)
}

// CacheUpdAtSet update the note updAt map
func (gp *GoogleProvider) CacheUpdAtSet(noteID int, updAt int64, nonBlocked bool) {
	if !nonBlocked {
		gp.updAtMux.Lock()
		defer gp.updAtMux.Unlock()
	}
	gp.notesUpdatedAt[noteID] = updAt
}

// CacheUpdAtGet returns the note updAt from the cache
func (gp *GoogleProvider) CacheUpdAtGet(noteID int) (int64, bool) {
	gp.updAtMux.RLock()
	defer gp.updAtMux.RUnlock()
	idx, ok := gp.notesUpdatedAt[noteID]
	return idx, ok
}

// CacheUpdAtUnset removes the note updAt from the cache
func (gp *GoogleProvider) CacheUpdAtUnset(noteID int) {
	gp.updAtMux.Lock()
	defer gp.updAtMux.Unlock()
	delete(gp.notesUpdatedAt, noteID)
}

// GetNotes fetch from the provider notes with given id or all if no ids is given
func (gp *GoogleProvider) GetNotes(ids ...int) ([]model.Note, error) {
	readRange := fmt.Sprintf("%s!A2:H", gp.sheetName)
	ctx, cancel := context.WithTimeout(gp.ctx, 10*time.Second)
	defer cancel()
	resp, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	notes := make([]model.Note, 0)
	if len(resp.Values) == 0 {
		log.Println("No data found in google sheet.")
	} else {
		for _, row := range resp.Values {
			// map the sheet row to a Note object
			note := gp.ParseSheetRow(row)
			notes = append(notes, note)
		}
	}
	// if ids is not empty, filter the notes
	if len(ids) > 0 {
		notes = gp.FilterNotes(notes, ids)
	}
	return notes, nil
}

// FilterNotes filters the notes by the given ids
func (*GoogleProvider) FilterNotes(notes []model.Note, ids []int) []model.Note {
	filteredNotes := make([]model.Note, 0)
	for _, id := range ids {
		for _, note := range notes {
			if note.ID == id {
				filteredNotes = append(filteredNotes, note)
				break
			}
		}
	}
	return filteredNotes
}

// GetNoteIDs returns a map of the note IDs and their index in the sheet
// Note: populates another map with the note IDs and their UpdatedAt field to be used for the sync
// TODO: find a way to use a single api call to get both the note IDs and the UpdatedAt fields
func (gp *GoogleProvider) GetNoteIDs(forceRemote bool) (map[int]int, error) {
	// always return the map from the local cache, unless forceRemote is true
	if len(gp.noteIds) > 0 && !forceRemote {
		return gp.noteIds, nil
	}
	ctx, cancel := context.WithTimeout(gp.ctx, 10*time.Second)
	defer cancel()
	// range used to read note IDs and UpdatedAt fields from the sheet
	readRangeID := fmt.Sprintf("%s!A2:A", gp.sheetName)
	respGetIDs, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRangeID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	readRangeUpdAt := fmt.Sprintf("%s!G2:G", gp.sheetName)
	respGetUpdAt, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRangeUpdAt).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	// add all IDs to a slice (make it a thread-safe map)
	gp.updAtMux.Lock()
	defer gp.updAtMux.Unlock()
	gp.idsMux.Lock()
	defer gp.idsMux.Unlock()
	gp.noteIds = make(map[int]int)
	gp.notesUpdatedAt = make(map[int]int64)
	for idx, row := range respGetIDs.Values {
		if len(row) < 1 {
			continue
		}
		// populate the note ID map with the note ID and its index in the slice
		noteID := common.StringToInt(row[0].(string))
		gp.CacheIDSet(noteID, idx, true)
		// populate the note updated at map with the note ID and its updated at field
		updAtVal := common.StringToInt64(respGetUpdAt.Values[idx][0].(string))
		gp.CacheUpdAtSet(noteID, updAtVal, true)
	}
	return gp.noteIds, nil
}

// GetNote returns the note with the given id
func (gp *GoogleProvider) GetNote(id int) (*model.Note, error) {
	// if noteIds map is empty, populate it
	if len(gp.noteIds) == 0 {
		_, err := gp.GetNoteIDs(true)
		if err != nil {
			return nil, err
		}
	}
	// check if the note ID is in the map
	noteIDx, ok := gp.CacheIDGet(id)
	if !ok {
		return nil, errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	noteIDx += 2 // add 2 to the index to get the correct row
	ctx, cancel := context.WithTimeout(gp.ctx, 10*time.Second)
	defer cancel()
	readRangeRow := fmt.Sprintf("%s!A%d:H%d", gp.sheetName, noteIDx, noteIDx)
	// read the note from sheet in readRangeRow
	respGetNote, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRangeRow).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	// parse the note from the sheet
	if len(respGetNote.Values) == 0 {
		return nil, errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	row := respGetNote.Values[0]
	note := gp.ParseSheetRow(row)
	return &note, nil
}

// PutNote pushes a note to the provider
// if the note does not exist, it will be created
func (gp *GoogleProvider) PutNote(note *model.Note) error {
	// if noteIds map is empty, populate it
	if len(gp.noteIds) == 0 {
		_, err := gp.GetNoteIDs(true)
		if err != nil {
			return err
		}
	}
	// check if the note ID is in the map
	noteIDx, ok := gp.CacheIDGet(note.ID)
	if !ok {
		// if not, create a new row at the end of the sheet
		cacheIdx := len(gp.noteIds)
		noteIDx = cacheIdx + 2 // add 2 to the index to get the correct row
		// add the note to the map
		gp.CacheIDSet(note.ID, cacheIdx, false)
	} else {
		// if yes, update the row
		noteIDx += 2 // add 2 to the index to get the correct row
	}
	// create/update a new row in the sheet
	writeRange := fmt.Sprintf("%s!A%d:H%d", gp.sheetName, noteIDx, noteIDx)
	values := [][]interface{}{
		{note.ID, note.Title, note.Content, note.Hidden, note.Encrypted, note.EncKeyName, note.CreatedAt, note.UpdatedAt},
	}
	ctx, cancel := context.WithTimeout(gp.ctx, 10*time.Second)
	defer cancel()
	_, err := gp.sheetsService.Spreadsheets.Values.Update(gp.sheetID, writeRange, &sheets.ValueRange{
		Values: values,
	}).Context(ctx).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

// DeleteNote deletes the note with the given id
func (gp *GoogleProvider) DeleteNote(id int) error {
	// if noteIds map is empty, populate it
	if len(gp.noteIds) == 0 {
		_, err := gp.GetNoteIDs(true)
		if err != nil {
			return err
		}
	}
	// check if the note ID is in the map
	noteIDx, ok := gp.CacheIDGet(id)
	if !ok {
		return errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	noteIDx += 2 // add 2 to the index to get the correct row
	// delete the row from the sheet
	ctx, cancel := context.WithTimeout(gp.ctx, 10*time.Second)
	defer cancel()
	deleteRange := fmt.Sprintf("%s!A%d:H%d", gp.sheetName, noteIDx, noteIDx)
	_, err := gp.sheetsService.Spreadsheets.Values.Clear(gp.sheetID, deleteRange, &sheets.ClearValuesRequest{}).
		Context(ctx).
		Do()
	if err != nil {
		return err
	}
	// delete the note from the map
	gp.CacheIDUnset(id)
	return nil
}

// SyncNotes syncs the notes from the provider to the local database and vice versa
// to correctly sync, we need to get all note ID (columb A of the sheet) and UpdatedAt fields (column G of the sheet) from the provider, then we need to get all notes from the local database, then we need to compare the two lists and sync the notes
// return the notes to be added to the local database
func (gp *GoogleProvider) SyncNotes(syncCtx context.Context, dbNotes []model.Note) (downloaded []model.Note, err error) {
	// get ids and updated at from the provider
	noteIds, err := gp.GetNoteIDs(true)
	if err != nil {
		return nil, err
	}
	gp.updAtMux.RLock()
	noteUpdAt := gp.notesUpdatedAt
	gp.updAtMux.RUnlock()

	// loop through the local notes and check if they exist in the provider.
	// if they do not exist, put them in the provider
	// if they do exist, check if they have been updated since the last sync and update them in the provider if they have
	for idx, dbNote := range dbNotes {
		// check if the note exists in the provider
		_, ok := noteIds[dbNote.ID]
		if !ok {
			// if not, create it
			err := gp.PutNote(&dbNote)
			if err != nil {
				return nil, err
			}
			continue
		}
		// if yes, check if it has been updated since the last sync
		if dbNote.UpdatedAt > noteUpdAt[dbNote.ID] {
			// if yes, update it in the provider
			err := gp.PutNote(&dbNote)
			if err != nil {
				return nil, err
			}
			continue
		}
		// if the the note from the provider has been updated since the last sync, update the local note with the new data
		if dbNote.UpdatedAt < noteUpdAt[dbNote.ID] {
			// get the note from the provider
			note, err := gp.GetNote(dbNote.ID)
			if err != nil {
				return nil, err
			}
			// update the local note with the new data
			dbNotes[idx] = *note
		}
	}
	// loop through the provider notes and check if they exist in the local database.
	// if they do not exist, get them from the provider and put them in the local database
	toAdd := make([]model.Note, 0)
	for noteID, _ := range noteIds {
		// check if the noteID exists dbNotes
		if len(dbNotes) == 0 {
			// if dbNotes is empty, get the note from the provider
			note, err := gp.GetNote(noteID)
			if err != nil {
				return nil, err
			}
			// add the note to the local database
			toAdd = append(toAdd, *note)
			continue
		}
		// if the noteID does not exist in dbNotes, get it from the provider
		found := false
		for _, dbNote := range dbNotes {
			if dbNote.ID == noteID {
				found = true
				break
			}
		}
		if !found {
			// get the note from the provider
			note, err := gp.GetNote(noteID)
			if err != nil {
				return nil, err
			}
			// add the note to the local database
			toAdd = append(toAdd, *note)
		}

	}

	// build the note titles array
	noteTitles := make([]string, len(dbNotes))
	for idx, note := range dbNotes {
		noteTitles[idx] = note.Title
	}
	gp.observer.Notify(observer.EVENT_UPDATE_NOTE_TITLES, noteTitles)
	return toAdd, nil
}

// Init initializes the provider
func (gp *GoogleProvider) Init() error {
	if gp.sheetID == "" || gp.sheetName == "" || gp.credFilePath == "" {
		return errors.New(common.ERR_INVALID_GOOGLE_PROVIDER_CONFIG)
	}
	gp.ctx = context.Background()
	client, err := getClientWithJWTToken(gp.ctx, gp.credFilePath)
	if err != nil {
		return err
	}
	gp.client = client
	gp.sheetsService, err = sheets.NewService(gp.ctx, option.WithHTTPClient(client))
	return err
}

// getClientWithJWTToken gets a http client with jwt token from service account
func getClientWithJWTToken(ctx context.Context, credFilePath string) (*http.Client, error) {
	c, err := ioutil.ReadFile(credFilePath)
	if err != nil {
		return nil, err
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.JWTConfigFromJSON(c, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config.Client(ctx), nil
}

// ParseSheetRow maps the sheet row to a Note object
func (*GoogleProvider) ParseSheetRow(row []interface{}) model.Note {
	note := model.Note{
		ID:         common.StringToInt(row[0].(string)),
		Title:      row[1].(string),
		Content:    row[2].(string),
		Hidden:     common.StringToBool(row[3].(string)),
		Encrypted:  common.StringToBool(row[4].(string)),
		EncKeyName: row[5].(string),
		CreatedAt:  common.StringToInt64(row[6].(string)),
		UpdatedAt:  common.StringToInt64(row[7].(string)),
	}
	return note
}

// UpdateNoteNotifier creates a new note observer to notify the provider when a note is created or updated
func (gp *GoogleProvider) UpdateNoteNotifier() observer.Listener {
	return observer.Listener{
		OnNotify: func(note interface{}, args ...interface{}) {
			if note == nil {
				return
			}
			// the encrypted note is args[2]. Check if exists and if it is encrypted
			if len(args) < 3 {
				gp.logger.Error("The note is not encrypted. Cannot push it to google sheets")
				return
			}
			n, ok := args[2].(*model.Note)
			if !ok {
				gp.logger.Errorf("Error cannot cast note struct: %v", note)
				return
			}

			// Add note to update queue instead of pulling synchronously
			select {
			case gp.updateQueue <- n:
			default:
				gp.logger.Warn("Update queue full, dropping Google Sheet sync for note ID", n.ID)
			}
		},
	}
}

// DeleteNoteNotifier creates a new note observer to notify the provider when a note is deleted
func (gp *GoogleProvider) DeleteNoteNotifier() observer.Listener {
	return observer.Listener{
		OnNotify: func(note interface{}, args ...interface{}) {
			if note == nil {
				return
			}
			n, ok := note.(*model.Note)
			if !ok {
				gp.logger.Errorf("Error cannot cast note struct: %v", note)
				return
			}

			// add the note ID to the delete queue
			select {
			case gp.deleteQueue <- n.ID:
			default:
				gp.logger.Warn("Delete queue full, dropping Google Sheet sync for note ID", n.ID)
			}
		},
	}
}

// InitWorker starts a background worker that processes sync queues
func (gp *GoogleProvider) InitWorker(ctx context.Context) {
	go func() {
		gp.logger.Info("Starting Google Sheets sync worker")
		for {
			select {
			case <-ctx.Done():
				gp.logger.Info("Shutting down Google Sheets sync worker")
				return
			case note := <-gp.updateQueue:
				if err := gp.PutNote(note); err != nil {
					gp.logger.Errorf("Worker error pushing note to google sheets (ID %d): %v", note.ID, err)
				}
			case id := <-gp.deleteQueue:
				if err := gp.DeleteNote(id); err != nil {
					gp.logger.Errorf("Worker error deleting from google sheets (ID %d): %v", id, err)
				}
			}
		}
	}()
}
