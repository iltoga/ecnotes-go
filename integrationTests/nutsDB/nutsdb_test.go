package nutsdb_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	configService  service.ConfigService
	noteService    service.NoteService
	noteRepository service.NoteServiceRepository
	kvdbPath       string
	defaultBucket  = "notes"
)

type nutsDBSuiteTest struct {
	suite.Suite
}

// SetupTest ....
func (s *nutsDBSuiteTest) SetupTest() {
	fmt.Println("SetupTest...")
	initDB()
}

// TearDownTest ....
func (s *nutsDBSuiteTest) TearDownTest() {
	fmt.Println("TearDownTest...")
	cleanup()
}

// TestSuite ....
func TestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip category mysql repository test")
	}
	suite.Run(t, new(nutsDBSuiteTest))
}

// TestCreateAndReadNote test if we can create a note, save it to nutsDB kv store and read it back
func (s *nutsDBSuiteTest) TestCreateAndReadNote() {
	var (
		t = s.T()
		// create a new note to have something in the db
		newNote = &model.Note{
			Title:   "Welcome to EcNotes",
			Content: "This is your first note.\n\nYou can edit it by clicking on the title.",
		}
	)
	// test creating a new note
	if err := noteService.CreateNote(newNote); err != nil {
		t.Error(err)
	}
	// this is because when created, the note is encrypted and when read it is decrypted
	newNote.Encrypted = false
	// get same note from db
	note, err := noteService.GetNoteWithContent(newNote.ID)
	if err != nil {
		t.Error(err)
	}
	newNote.Content = "This is your first note.\n\nYou can edit it by clicking on the title."
	assert.Equal(t, newNote, note, "Note should be the same")
}

// TestUpdateNote test if we can update a note (since updating the note also deletes the old one) we can test delete too
func (s *nutsDBSuiteTest) TestUpdateDeleteNote() {
	var (
		t = s.T()
		// create a new note to have something in the db
		newNote = &model.Note{
			Title:   "Welcome to EcNotes",
			Content: "This is your first note.\n\nYou can edit it by clicking on the title.",
		}
	)
	// test creating a new note
	err := noteService.CreateNote(newNote)
	assert.NoError(t, err, "Error creating note")

	// test update note's title (deletes old note and creates a new one)
	newTitle := "Welcome to EcNotes - updated"
	newContent := "This is your first note.\n\nYou can edit it by clicking on the title.\n\nUpdated!"
	newID := noteRepository.GetIDFromTitle(newTitle)
	oldID := noteRepository.GetIDFromTitle(newNote.Title)
	_, err = noteService.UpdateNoteTitle(newNote.Title, newTitle)
	assert.NoError(t, err)
	ok, err := noteRepository.NoteExists(oldID)
	assert.Error(t, err, "key not found in the bucket")
	assert.False(t, ok, "Old note should not exist anymore")
	// get same note from db
	updatedNote, err := noteService.GetNoteWithContent(newID)
	assert.NoError(t, err, "Error getting note with new title")
	updatedNote.Content = newContent
	// update note content
	err = noteService.UpdateNoteContent(updatedNote)
	assert.NoError(t, err, "Error updating note content")

	// restore unencrypted content to compare
	updatedNote.Content = newContent
	// this is because when created or updated, the note is encrypted and when read it is decrypted
	updatedNote.Encrypted = false
	// get same note from db
	updatedNoteFromDB, err := noteService.GetNoteWithContent(newID)
	assert.NoError(t, err, "Error getting note with new content from db")
	assert.Equal(t, updatedNote, updatedNoteFromDB, "Note should be the same")
}

func mockConfig() {
	// make sure we have the encryption key
	key, err := configService.GetConfigBytes(common.CONFIG_ENCRYPTION_KEY)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// decrypt the key

	decKey, err := cryptoUtil.DecryptMessage(key, "1234")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// add decrypted key to globals so we can use it to encrypt/decrypt notes before storing them
	configService.SetGlobal(common.CONFIG_ENCRYPTION_KEY, string(decKey))
	kvdbPath, err = configService.GetConfig(common.CONFIG_KVDB_PATH)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initDB() {
	var err error
	// mocking ressource path for integration tests
	configService = &service.ConfigServiceImpl{
		// ResourcePath: "./integrationTests/nutsDB/testResources",
		ResourcePath: "./testResources",
		Config:       make(map[string]string),
		Globals:      make(map[string]string),
		Loaded:       false,
		ConfigMux:    &sync.RWMutex{},
		GlobalsMux:   &sync.RWMutex{},
	}
	if err = configService.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// mock configuration so that we can use the db for testing without a UI
	mockConfig()
	noteRepository, err = service.NewNoteServiceRepository(kvdbPath, defaultBucket, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cryptoSrvF := &service.CryptoServiceFactoryImpl{
		Srv: service.NewCryptoServiceAES(service.NewKeyManagementServiceAES()),
	}
	cryptoSrvF.Srv.GetKeyManager().ImportKey([]byte("1234"))
	noteService = service.NewNoteService(noteRepository, configService, observer.NewObserver(), cryptoSrvF)
}

func cleanup() {
	fmt.Println("Cleanup...")
	kvDBPath, err := configService.GetConfig(common.CONFIG_KVDB_PATH)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// delete db to cleanup test enviroment
	if err := os.RemoveAll(kvDBPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}
