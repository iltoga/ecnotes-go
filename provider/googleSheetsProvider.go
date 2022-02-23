package provider

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/iltoga/ecnotes-go/model"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	BaseSyncNoteProvider
	client       *http.Client
	sheetName    string
	sheetID      string
	credFilePath string
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
	}
	if err := gp.Init(); err != nil {
		return nil, err
	}
	return gp, nil
}

// GetNoteIDs returns the list of note IDs from the provider
func (gp *GoogleProvider) GetNoteIDs() ([]int, error) {
	panic("Not implemented")
}

// GetNote returns the note with the given id
func (gp *GoogleProvider) GetNote(id int) (*model.Note, error) {
	panic("Not implemented")
}

// PutNote puts the given note into the provider
func (gp *GoogleProvider) PutNote(note *model.Note) error {
	panic("Not implemented")
}

// DeleteNote deletes the note with the given id
func (gp *GoogleProvider) DeleteNote(id int) error {
	panic("Not implemented")
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
	ctx := context.Background()
	client, err := getClientWithJWTToken(ctx, gp.credFilePath)
	if err != nil {
		return err
	}
	gp.client = client
	return nil
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
