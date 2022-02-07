package service

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/xujiajun/nutsdb"
)

// NoteServiceRepository interface for querying notes
type NoteServiceRepository interface {
	GetAllNotes() ([]Note, error)
	GetNote(id int) (*Note, error)
	CreateNote(note *Note) error
	UpdateNote(note *Note) error
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
) (NoteServiceRepository, error) {
	db, err := openDBConnection(dbPath)
	if err != nil {
		return nil, err
	}
	return &NoteServiceRepositoryImpl{
		dbPath: dbPath,
		db:     db,
		bucket: bucket,
	}, nil
}

func openDBConnection(dbPath string) (*nutsdb.DB, error) {
	opt := nutsdb.DefaultOptions

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
	opt.Dir = dbPath
	opt.SegmentSize = 1024 * 1024 // 1MB
	return nutsdb.Open(opt)
}

// GetAllNotes retreives all notes from the db (already decrypted)
func (nsr *NoteServiceRepositoryImpl) GetAllNotes() ([]Note, error) {
	notes := []Note{}
	if err := nsr.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(nsr.bucket)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				var note Note
				if err := common.UnmarshalJSON(entry.Value, &note); err != nil {
					return err
				}
				notes = append(notes, note)
				// fmt.Println(string(entry.Key), string(entry.Value))
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return notes, nil
}

// GetNote retreives a note from the db by its ID (already decrypted)
func (nsr *NoteServiceRepositoryImpl) GetNote(id int) (*Note, error) {
	var note *Note
	if err := nsr.db.View(
		func(tx *nutsdb.Tx) error {
			dbEntry, err := tx.Get(nsr.bucket, nsr.getDBKeyFromID(id))
			if err != nil {
				return err
			}
			// unmarshal note
			if err := common.UnmarshalJSON(dbEntry.Value, &note); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return note, nil
}

// CreateNote adds a new note to the db
// Note: note's content has already been encrypted at service layer
func (nsr *NoteServiceRepositoryImpl) CreateNote(note *Note) error {
	var (
		key        = []byte(fmt.Sprintf("%d", note.ID))
		value, err = common.MarshalJSON(note)
	)
	if err != nil {
		return err
	}
	if err = nsr.db.Update(
		func(tx *nutsdb.Tx) error {
			key := key
			val := value
			return tx.Put(nsr.bucket, key, val, 0)
		}); err != nil {
		return err
	}
	return nil
}

// UpdateNote update a note in the db
func (nsr *NoteServiceRepositoryImpl) UpdateNote(note *Note) error {
	if exists, _ := nsr.NoteExists(note.ID); !exists {
		return errors.New(common.ERR_NOTE_NOT_FOUND)
	}
	if err := nsr.db.Update(
		func(tx *nutsdb.Tx) error {
			key := nsr.getDBKeyFromID(note.ID)
			val, err := common.MarshalJSON(note)
			if err != nil {
				return err
			}
			return tx.Put(nsr.bucket, key, val, 0)
		}); err != nil {
		return err
	}
	return nil
}

// DeleteNote deletes a note from the db
func (nsr *NoteServiceRepositoryImpl) DeleteNote(id int) error {
	// delete note by id
	if err := nsr.db.Update(
		func(tx *nutsdb.Tx) error {
			key := nsr.getDBKeyFromID(id)
			return tx.Delete(nsr.bucket, key)
		}); err != nil {
		return err
	}
	return nil
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
