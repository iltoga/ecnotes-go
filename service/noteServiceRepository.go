package service

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/xujiajun/nutsdb"
)

// NoteServiceRepository interface for querying notes
type NoteServiceRepository interface {
	GetAllNotes() ([]model.Note, error)
	GetNote(id int) (*model.Note, error)
	CreateNote(note *model.Note) error
	UpdateNote(note *model.Note) error
	DeleteNote(id int) error
	NoteExists(id int) (bool, error)
	GetIDFromTitle(title string) int
}

// NoteServiceRepositoryImpl implementation of NoteServiceRepository that uses nutsdb
type NoteServiceRepositoryImpl struct {
	dbPath string
	db     *nutsdb.DB
	bucket string
}

// NewNoteServiceRepository constructor for NoteServiceRepositoryImpl
func NewNoteServiceRepository(
	dbPath string,
	bucket string,
	resetDB bool,
) (NoteServiceRepository, error) {
	db, err := openDBConnection(dbPath, resetDB)
	if err != nil {
		return nil, err
	}
	return &NoteServiceRepositoryImpl{
		dbPath: dbPath,
		db:     db,
		bucket: bucket,
	}, nil
}

func openDBConnection(dbPath string, resetDB bool) (*nutsdb.DB, error) {
	opt := nutsdb.DefaultOptions

	if resetDB {
		files, _ := ioutil.ReadDir(dbPath)
		for _, f := range files {
			name := f.Name()
			if name != "" {
				fmt.Println(dbPath + "/" + name)
				err := os.RemoveAll(dbPath + "/" + name)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	opt.Dir = dbPath
	opt.SegmentSize = 1024 * 1024 // 1MB
	return nutsdb.Open(opt)
}

// GetAllNotes retreives all notes from the db (already decrypted)
func (nsr *NoteServiceRepositoryImpl) GetAllNotes() ([]model.Note, error) {
	notes := []model.Note{}
	if err := nsr.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(nsr.bucket)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				var note model.Note
				if err := common.UnmarshalJSON(entry.Value, &note); err != nil {
					return err
				}
				notes = append(notes, note)
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return notes, nil
}

// GetNote retreives a note from the db by its ID (already decrypted)
func (nsr *NoteServiceRepositoryImpl) GetNote(id int) (*model.Note, error) {
	var note *model.Note
	if err := nsr.db.View(
		func(tx *nutsdb.Tx) error {
			dbEntry, err := tx.Get(nsr.bucket, nsr.getDBKeyFromID(id))
			if err != nil {
				return err
			}
			// unmarshal note
			return common.UnmarshalJSON(dbEntry.Value, &note)
		}); err != nil {
		return nil, err
	}
	return note, nil
}

// CreateNote adds a new note to the db
// model.Note: note's content has already been encrypted at service layer
func (nsr *NoteServiceRepositoryImpl) CreateNote(note *model.Note) error {
	var (
		key        = []byte(fmt.Sprintf("%d", note.ID))
		value, err = common.MarshalJSON(note)
	)
	if err != nil {
		return err
	}
	return nsr.db.Update(
		func(tx *nutsdb.Tx) error {
			key := key
			val := value
			return tx.Put(nsr.bucket, key, val, 0)
		})
}

// UpdateNote update a note in the db
func (nsr *NoteServiceRepositoryImpl) UpdateNote(note *model.Note) error {
	if exists, _ := nsr.NoteExists(note.ID); !exists {
		return errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	return nsr.db.Update(
		func(tx *nutsdb.Tx) error {
			key := nsr.getDBKeyFromID(note.ID)
			val, err := common.MarshalJSON(note)
			if err != nil {
				return err
			}
			return tx.Put(nsr.bucket, key, val, 0)
		})
}

// DeleteNote deletes a note from the db
func (nsr *NoteServiceRepositoryImpl) DeleteNote(id int) error {
	// delete note by id
	return nsr.db.Update(
		func(tx *nutsdb.Tx) error {
			key := nsr.getDBKeyFromID(id)
			return tx.Delete(nsr.bucket, key)
		})
}

// NoteExists checks if a note exists in the db
func (nsr *NoteServiceRepositoryImpl) NoteExists(id int) (bool, error) {
	if err := nsr.db.View(
		func(tx *nutsdb.Tx) error {
			_, err := tx.Get(nsr.bucket, nsr.getDBKeyFromID(id))
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
		return false, err
	}
	return true, nil
}

// GetIDFromTitle retreives a note's ID from its title
func (nsr *NoteServiceRepositoryImpl) GetIDFromTitle(title string) int {
	return int(cryptoUtil.IndexFromString(title))
}

// getDBKeyFromID returns the key formatted for nutsdb
func (nsr *NoteServiceRepositoryImpl) getDBKeyFromID(id int) []byte {
	return []byte(fmt.Sprintf("%d", id))
}
