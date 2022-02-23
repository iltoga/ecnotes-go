package provider

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleProvider struct {
	BaseSyncNoteProvider
	sheetsService *sheets.Service
	client        *http.Client
	sheetName     string
	sheetID       string
	credFilePath  string
	noteIds       map[int]int
	mux           *sync.RWMutex
	ctx           context.Context
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(
	sheetName string,
	sheetID string,
	credFilePath string,
) (*GoogleProvider, error) {
	gp := &GoogleProvider{
		sheetName:    sheetName,
		sheetID:      sheetID,
		credFilePath: credFilePath,
		noteIds:      make(map[int]int),
		mux:          &sync.RWMutex{},
	}
	if err := gp.Init(); err != nil {
		return nil, err
	}
	return gp, nil
}

// CacheIDSet update the note ID map
func (gp *GoogleProvider) CacheIDSet(noteID int, noteIDx int, nonBlocked bool) {
	if !nonBlocked {
		gp.mux.Lock()
		defer gp.mux.Unlock()
	}
	gp.noteIds[noteID] = noteIDx
}

// CacheIDGet returns the note ID from the cache
func (gp *GoogleProvider) CacheIDGet(noteID int) (int, bool) {
	gp.mux.RLock()
	defer gp.mux.RUnlock()
	idx, ok := gp.noteIds[noteID]
	return idx, ok
}

// CacheIDUnset removes the note ID from the cache
func (gp *GoogleProvider) CacheIDUnset(noteID int) {
	gp.mux.Lock()
	defer gp.mux.Unlock()
	delete(gp.noteIds, noteID)
}

// GetNotes returns all notes from the provider
func (gp *GoogleProvider) GetNotes() ([]model.Note, error) {
	readRange := fmt.Sprintf("%s!A2:G", gp.sheetName)
	resp, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	notes := make([]model.Note, 0)
	if len(resp.Values) == 0 {
		log.Println("No data found in google sheet.")
	} else {
		for _, row := range resp.Values {
			if len(row) < 5 {
				continue
			}
			// map the sheet row to a Note object
			note := model.Note{
				ID:        common.StringToInt(row[0].(string)),
				Title:     row[1].(string),
				Content:   row[2].(string),
				CreatedAt: common.StringToInt64(row[3].(string)),
				UpdatedAt: common.StringToInt64(row[4].(string)),
			}
			notes = append(notes, note)
		}
	}
	return notes, nil
}

// GetNoteIDs returns the list of note IDs from the provider
func (gp *GoogleProvider) GetNoteIDs(forceRemote bool) (map[int]int, error) {
	// always return the map from the local cache, unless forceRemote is true
	if len(gp.noteIds) > 0 && !forceRemote {
		return gp.noteIds, nil
	}
	// range used to read note IDs
	readRangeID := fmt.Sprintf("%s!A2:A", gp.sheetName)
	respGetIDs, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRangeID).Do()
	if err != nil {
		return nil, err
	}
	// add all IDs to a slice (make it a thread-safe map)
	gp.mux.Lock()
	defer gp.mux.Unlock()
	gp.noteIds = make(map[int]int)
	for idx, row := range respGetIDs.Values {
		if len(row) < 1 {
			continue
		}
		// populate the note ID map with the note ID and its index in the slice
		gp.CacheIDSet(common.StringToInt(row[0].(string)), idx, true)
	}
	return gp.noteIds, nil
}

// GetNote returns the note with the given id
func (gp *GoogleProvider) GetNote(id int) (*model.Note, error) {
	// if noteIds map is empty, populate it
	if len(gp.noteIds) == 0 {
		_, err := gp.GetNoteIDs(false)
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
	readRangeRow := fmt.Sprintf("%s!A%d:G%d", gp.sheetName, noteIDx, noteIDx)
	// read the note from sheet in readRangeRow
	respGetNote, err := gp.sheetsService.Spreadsheets.Values.Get(gp.sheetID, readRangeRow).Do()
	if err != nil {
		return nil, err
	}
	// parse the note from the sheet
	if len(respGetNote.Values) == 0 {
		return nil, errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	row := respGetNote.Values[0]
	note := &model.Note{
		ID:        common.StringToInt(row[0].(string)),
		Title:     row[1].(string),
		Content:   row[2].(string),
		Hidden:    common.StringToBool(row[3].(string)),
		Encrypted: common.StringToBool(row[4].(string)),
		CreatedAt: common.StringToInt64(row[5].(string)),
		UpdatedAt: common.StringToInt64(row[6].(string)),
	}
	return note, nil
}

// PutNote pushes a note to the provider
// if the note does not exist, it will be created
func (gp *GoogleProvider) PutNote(note *model.Note) error {
	// if noteIds map is empty, populate it
	if len(gp.noteIds) == 0 {
		_, err := gp.GetNoteIDs(false)
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
	writeRange := fmt.Sprintf("%s!A%d:G%d", gp.sheetName, noteIDx, noteIDx)
	values := [][]interface{}{
		{note.ID, note.Title, note.Content, note.Hidden, note.Encrypted, note.CreatedAt, note.UpdatedAt},
	}
	_, err := gp.sheetsService.Spreadsheets.Values.Update(gp.sheetID, writeRange, &sheets.ValueRange{
		Values: values,
	}).Context(gp.ctx).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

// DeleteNote deletes the note with the given id
func (gp *GoogleProvider) DeleteNote(id int) error {
	// if noteIds map is empty, populate it
	if len(gp.noteIds) == 0 {
		_, err := gp.GetNoteIDs(false)
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
	deleteRange := fmt.Sprintf("%s!A%d:G%d", gp.sheetName, noteIDx, noteIDx)
	_, err := gp.sheetsService.Spreadsheets.Values.Clear(gp.sheetID, deleteRange, &sheets.ClearValuesRequest{}).Context(gp.ctx).Do()
	if err != nil {
		return err
	}
	// delete the note from the map
	gp.CacheIDUnset(id)
	return nil
}

// FindNotes returns all notes that match the given query
func (gp *GoogleProvider) FindNotes(query string) ([]model.Note, error) {
	panic("Not implemented")
}

// SyncNotes syncs the notes from the provider to the local database and vice versa
func (gp *GoogleProvider) SyncNotes() error {
	panic("Not implemented")
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
