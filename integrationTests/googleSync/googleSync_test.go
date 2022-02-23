package googleSync_test

import (
	"fmt"
	"testing"

	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	testSheetID   = "10s0pRlrvskhGfGkJYKeJidEZIsBPPmhMvHDEOtuT_20"
	testSheetName = "test_notes"
)

type googleSyncTest struct {
	suite.Suite
}

// SetupTest ....
func (s *googleSyncTest) SetupTest() {
	fmt.Println("SetupTest...")
	// initDB()
}

// TearDownTest ....
func (s *googleSyncTest) TearDownTest() {
	fmt.Println("TearDownTest...")
	// cleanup()
}

// TestSuite ....
func TestSuite(t *testing.T) {
	suite.Run(t, new(googleSyncTest))
}

func getGP() *provider.GoogleProvider {
	gp, err := provider.NewGoogleProvider(
		testSheetName,
		testSheetID,
		"/home/demo/.config/ecnotes/providers/google/cred_serviceaccount.json",
	)
	if err != nil {
		panic(err)
	}
	return gp
}

// TestGetNote ....
func (s *googleSyncTest) TestGetNote() {
	gp := getGP()
	note, err := gp.GetNote(3782526374)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), note.ID, 3782526374)
	assert.Equal(s.T(), note.Title, "third note")
	assert.Equal(s.T(), note.Content, "b0a45d2207da56cb6ce5757fd441c51f3fd63614e2febaa10a1bd3f34109f744ff76dd34e2c5dadd3e596e015d71945d5b91ea")
	assert.Equal(s.T(), note.CreatedAt, int64(1645516749891))
	assert.Equal(s.T(), note.UpdatedAt, int64(1645516749891))
}

// TestGetNotes ....
func (s *googleSyncTest) TestGetNotes() {
	t := s.T()
	// read and print all notes
	notes, err := getGP().GetNotes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(notes), 3)
}

// TestPutNote ....
func (s *googleSyncTest) TestPutAndDeleteNote() {
	t := s.T()
	gp := getGP()
	// create a new note
	note := &model.Note{
		Title:     "test note",
		Content:   "test content",
		CreatedAt: int64(1645516749891),
		UpdatedAt: int64(1645516749891),
	}
	// save the note
	err := gp.PutNote(note)
	assert.Nil(t, err)
	// read and print all notes
	notes, err := gp.GetNotes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(notes), 4)

	// update the note
	note.Title = "updated test note"
	note.Content = "updated test content"
	note.UpdatedAt = int64(1645516749891)
	err = gp.PutNote(note)
	assert.Nil(t, err)
	// read updated note and assert it
	note, err = gp.GetNote(note.ID)
	assert.Nil(t, err)
	assert.Equal(t, note.Title, "updated test note")
	assert.Equal(t, note.Content, "updated test content")
	assert.Equal(t, note.CreatedAt, int64(1645516749891))
	assert.Equal(t, note.UpdatedAt, int64(1645516749891))

	// delete the note
	err = gp.DeleteNote(note.ID)
	assert.Nil(t, err)
	// read and print all notes
	notes, err = gp.GetNotes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(notes), 3)
	// get the deleted note and assert it
	noteID := note.ID
	note, err = gp.GetNote(note.ID)
	assert.NotNil(t, err)
	assert.Nil(t, note)
	// also check that the note ID is not in the cache
	_, ok := gp.CacheIDGet(noteID)
	assert.False(t, ok)
}

// TestGetNoteIDs get note IDs from google sheet and assert it
func (s *googleSyncTest) TestGetNoteIDs() {
	t := s.T()
	gp := getGP()
	// read all note IDs from google sheet
	ids, err := gp.GetNoteIDs(true)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(ids), 3)
	// asssert ids
	assert.Equal(t, ids[98304983], 0)
	assert.Equal(t, ids[3782526374], 1)
	assert.Equal(t, ids[1839475811], 2)

	// read all note IDs from cache
	// remove one element from cache
	gp.CacheIDUnset(3782526374)
	ids, err = gp.GetNoteIDs(false)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(ids), 2)
	// asssert ids
	assert.Equal(t, ids[98304983], 0)
	assert.Equal(t, ids[1839475811], 2)
	assert.Empty(t, ids[3782526374])
}
