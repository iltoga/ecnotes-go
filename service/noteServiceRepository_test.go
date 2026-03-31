package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/iltoga/ecnotes-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNoteRepository(t *testing.T) NoteServiceRepository {
	t.Helper()

	dir := t.TempDir()
	repo, err := NewNoteServiceRepository(dir, "notes", true)
	require.NoError(t, err)

	if impl, ok := repo.(*NoteServiceRepositoryImpl); ok {
		t.Cleanup(func() {
			require.NoError(t, impl.db.Close())
		})
	}

	return repo
}

func sampleRepoNote(id int, title string) *model.Note {
	return &model.Note{
		ID:         id,
		Title:      title,
		Content:    "content-" + title,
		Hidden:     false,
		Encrypted:  true,
		EncKeyName: "test-key",
		CreatedAt:  1000 + int64(id),
		UpdatedAt:  2000 + int64(id),
	}
}

func TestNoteServiceRepository_CRUDRoundTrip(t *testing.T) {
	repo := newTestNoteRepository(t)

	note := sampleRepoNote(1, "alpha")
	require.NoError(t, repo.CreateNote(note))

	exists, err := repo.NoteExists(note.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	got, err := repo.GetNote(note.ID)
	require.NoError(t, err)
	assert.Equal(t, note, got)

	allNotes, err := repo.GetAllNotes()
	require.NoError(t, err)
	require.Len(t, allNotes, 1)
	assert.Equal(t, *note, allNotes[0])

	note.Content = "updated-content"
	require.NoError(t, repo.UpdateNote(note))

	updated, err := repo.GetNote(note.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated-content", updated.Content)

	require.NoError(t, repo.DeleteNote(note.ID))
	exists, err = repo.NoteExists(note.ID)
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestNoteServiceRepository_RenameNoteAndLookupHelpers(t *testing.T) {
	repo := newTestNoteRepository(t)

	oldNote := sampleRepoNote(11, "old-title")
	require.NoError(t, repo.CreateNote(oldNote))

	renamed := sampleRepoNote(42, "new-title")
	require.NoError(t, repo.RenameNote(oldNote.ID, renamed))

	exists, err := repo.NoteExists(oldNote.ID)
	assert.Error(t, err)
	assert.False(t, exists)

	exists, err = repo.NoteExists(renamed.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	got, err := repo.GetNote(renamed.ID)
	require.NoError(t, err)
	assert.Equal(t, renamed, got)

	titleID := repo.GetIDFromTitle(renamed.Title)
	assert.Equal(t, titleID, repo.GetIDFromTitle(renamed.Title))
	assert.NotZero(t, titleID)
	assert.Equal(t, []byte(fmt.Sprintf("%d", renamed.ID)), repo.(*NoteServiceRepositoryImpl).getDBKeyFromID(renamed.ID))
}

func TestNoteServiceRepository_NewNoteServiceRepository_ResetsExistingFiles(t *testing.T) {
	dir := t.TempDir()
	junkFile := filepath.Join(dir, "junk.txt")
	require.NoError(t, os.WriteFile(junkFile, []byte("junk"), 0o600))

	repo, err := NewNoteServiceRepository(dir, "notes", true)
	require.NoError(t, err)

	if impl, ok := repo.(*NoteServiceRepositoryImpl); ok {
		t.Cleanup(func() {
			require.NoError(t, impl.db.Close())
		})
	}

	_, err = os.Stat(junkFile)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestNoteServiceRepository_ErrorPaths(t *testing.T) {
	repo := newTestNoteRepository(t)

	_, err := repo.GetNote(999)
	assert.Error(t, err)

	ok, err := repo.NoteExists(999)
	assert.Error(t, err)
	assert.False(t, ok)

	require.NoError(t, repo.DeleteNote(999))
}
