package googleSync_test

import (
	"fmt"
	"testing"

	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/provider"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	testSheetID   = "10s0pRlrvskhGfGkJYKeJidEZIsBPPmhMvHDEOtuT_20"
	testSheetName = "test_notes"
	defaultNotes  = []model.Note{
		{
			ID:        98304983,
			Title:     "second note",
			Content:   "e890660c9204ab0feb9835879f66bb0896880d29012d5ebc13fe84a1b2d9001ffe24255f1fbfb2a65eefe89540f8958a09344e9a",
			Hidden:    false,
			Encrypted: true,
			CreatedAt: int64(1645412392244),
			UpdatedAt: int64(1645435461544),
		},
		{
			ID:        3782526374,
			Title:     "third note",
			Content:   "b0a45d2207da56cb6ce5757fd441c51f3fd63614e2febaa10a1bd3f34109f744ff76dd34e2c5dadd3e596e015d71945d5b91ea",
			Hidden:    false,
			Encrypted: true,
			CreatedAt: int64(1645516749891),
			UpdatedAt: int64(1645516749891),
		},
		{
			ID:        1839475811,
			Title:     "fourth note",
			Content:   "740c6a7c6b1125f431cfa3a1c80cdc6ad1250b3f4cc809972fa8d8fffe9c4a5ae1ec453503f506a091859b98065674648b0b4bbb",
			CreatedAt: int64(1645516773308),
			UpdatedAt: int64(1645516773308),
		},
	}
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
		logrus.New(),           //TODO: mock logger
		observer.NewObserver(), // TODO: mock observer
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
	assert.Equal(
		s.T(),
		note.Content,
		"b0a45d2207da56cb6ce5757fd441c51f3fd63614e2febaa10a1bd3f34109f744ff76dd34e2c5dadd3e596e015d71945d5b91ea",
	)
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

// TestGetNotes ....
func (s *googleSyncTest) TestGetNotes_Filtered() {
	t := s.T()
	// mock some ids
	ids := []int{3782526374, 1839475811}
	// read and print all notes and assert the filtered ones
	notes, err := getGP().GetNotes(ids...)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(notes), 2)
	assert.Equal(t, notes[0].ID, 3782526374)
	assert.Equal(t, notes[1].ID, 1839475811)
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

// TestFilterNotes ....
func (s *googleSyncTest) TestFilterNotes() {
	t := s.T()
	gp := getGP()
	// mock some notes
	notes := []model.Note{
		{
			ID:        98304983,
			Title:     "first note",
			Content:   "b0a45d2207da56cb6ce5757fd441c51f3fd63614e2febaa10a1bd3f34109f744ff76dd34e2c5dadd3e596e015d71945d5b91ea",
			CreatedAt: int64(1645516749891),
			UpdatedAt: int64(1645516749891),
		},
		{
			ID:        1839475811,
			Title:     "second note",
			Content:   "b0a45d2207da56cb6ce5757fd441c51f3fd63614e2febaa10a1bd3f34109f744ff76dd34e2c5dadd3e596e015d71945d5b91ea",
			CreatedAt: int64(1645516749891),
			UpdatedAt: int64(1645516749891),
		},
	}
	// mock some ids
	ids := []int{
		98304983,
		11111111,
		1839475811,
	}
	// filter notes and assert it
	filteredNotes := gp.FilterNotes(notes, ids)
	assert.Equal(t, len(filteredNotes), 2)
	assert.Equal(t, filteredNotes[0].ID, 98304983)
	assert.Equal(t, filteredNotes[1].ID, 1839475811)
}

// TestSyncNotes ....
func (s *googleSyncTest) TestSyncNotes() {
	t := s.T()
	gp := getGP()
	// mock some notes
	newNotes := []model.Note{
		{
			ID:        11111111,
			Title:     "first note",
			Content:   "blahblahblah",
			CreatedAt: int64(1645516749891),
			UpdatedAt: int64(1645516749891),
		},
	}
	// combine new notes and defaultNotes
	dbNotes := append(defaultNotes, newNotes...)
	// sync notes and assert it
	_, err := gp.SyncNotes(dbNotes)
	assert.Nil(t, err)
	// read and print all notes
	notes, err := gp.GetNotes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, len(notes), 4)
	// asssert notes
	assert.Equal(t, notes[3].ID, 11111111)
	assert.Equal(t, notes[3].Title, "first note")
	assert.Equal(t, notes[3].Content, "blahblahblah")
	assert.Equal(t, notes[3].CreatedAt, int64(1645516749891))
	assert.Equal(t, notes[3].UpdatedAt, int64(1645516749891))

	// delete the new note
	err = gp.DeleteNote(11111111)
	assert.Nil(t, err)
}
