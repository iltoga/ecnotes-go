package service_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
)

type NoteRepositoryMockImpl struct {
	mockedNotes []service.Note
}

func NewNoteRepositoryMock() *NoteRepositoryMockImpl {
	return &NoteRepositoryMockImpl{
		mockedNotes: []service.Note{
			{
				ID:      1,
				Title:   "Mandela quote",
				Content: "The greatest glory in living lies not in never falling, but in rising every time we fall. -Nelson Mandela",
			},
			{
				ID:      2,
				Title:   "The way to get started is to quit talking and begin doing",
				Content: "Disney is the best company ever. - Walt Disney",
			},
			{
				ID:      3,
				Title:   "Oprah Winfrey quote",
				Content: "If you look at what you have in life, you'll always have more. If you look at what you don't have in life, you'll never have enough",
			},
			{
				ID:      4,
				Title:   "The best is yet to come, Jhon Lennon",
				Content: "Life is what happens when you're busy making other plans",
			},
			{
				ID:      5,
				Title:   "The future belongs to those who believe in the beauty of their dreams",
				Content: "Eleanor Roosevelt",
			},
			{
				ID:      6,
				Title:   "The best is yet to come, Jhon Lennon",
				Content: "Life is what happens when you're busy making other plans",
			},
		},
	}
}

// GetAllNotes ....
func (nsr *NoteRepositoryMockImpl) GetAllNotes() ([]service.Note, error) {
	return nsr.mockedNotes, nil
}

// GetNote ....
func (nsr *NoteRepositoryMockImpl) GetNote(id int) (service.Note, error) {
	for _, note := range nsr.mockedNotes {
		if note.ID == id {
			return note, nil
		}
	}
	return service.Note{}, errors.New(common.ERR_NOTE_NOT_FOUND)
}

// CreateNote ....
func (nsr *NoteRepositoryMockImpl) CreateNote(note service.Note) error {
	nsr.mockedNotes = append(nsr.mockedNotes, note)
	return nil
}

// UpdateNote ....
func (nsr *NoteRepositoryMockImpl) UpdateNote(note service.Note) error {
	for i, n := range nsr.mockedNotes {
		if n.ID == note.ID {
			nsr.mockedNotes[i] = note
			return nil
		}
	}
	return errors.New(common.ERR_NOTE_NOT_FOUND)
}

// DeleteNote ....
func (nsr *NoteRepositoryMockImpl) DeleteNote(id int) error {
	for i, n := range nsr.mockedNotes {
		if n.ID == id {
			nsr.mockedNotes = append(nsr.mockedNotes[:i], nsr.mockedNotes[i+1:]...)
			return nil
		}
	}
	return errors.New(common.ERR_NOTE_NOT_FOUND)
}

func TestNoteServiceImpl_GetNote(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Note
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			got, err := ns.GetNote(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.GetNote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.NoteServiceImpl.GetNote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoteServiceImpl_GetNotes(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []service.Note
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			got, err := ns.GetNotes()
			if (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.GetNotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.NoteServiceImpl.GetNotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoteServiceImpl_CreateNote(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	type args struct {
		note service.Note
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			if err := ns.CreateNote(tt.args.note); (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.CreateNote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoteServiceImpl_UpdateNote(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	type args struct {
		note service.Note
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			if err := ns.UpdateNote(tt.args.note); (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.UpdateNote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoteServiceImpl_DeleteNote(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			if err := ns.DeleteNote(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.DeleteNote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoteServiceImpl_EncryptNote(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	type args struct {
		note *service.Note
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			if err := ns.EncryptNote(tt.args.note); (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.EncryptNote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoteServiceImpl_DecryptNote(t *testing.T) {
	type fields struct {
		TitlesIDMap map[string]int
		Titles      []string
	}
	type args struct {
		note *service.Note
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				TitlesIDMap: tt.fields.TitlesIDMap,
				Titles:      tt.fields.Titles,
			}
			if err := ns.DecryptNote(tt.args.note); (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.DecryptNote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNoteServiceImpl_SearchNotes(t *testing.T) {
	type fields struct {
		NoteRepo      service.NoteServiceRepository
		ConfigService service.ConfigService
		TitlesIDMap   map[string]int
		Titles        []string
		TitlesIDMutex *sync.RWMutex
	}
	type args struct {
		query       string
		fuzzySearch bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]int
		wantErr bool
	}{
		{
			name: "Search notes",
			fields: fields{
				NoteRepo:      NewNoteRepositoryMock(),
				ConfigService: service.NewConfigService(),
				TitlesIDMap:   map[string]int{},
				Titles:        []string{},
				TitlesIDMutex: &sync.RWMutex{},
			},
			args: args{
				query:       "Mandela quote",
				fuzzySearch: false,
			},
			want: map[string]int{
				"00628e854bd9b090d494b49b8a2b4063f622411593b34fba27a041d5e1a8c8a7": 1,
			},
			wantErr: false,
		},
		{
			name: "Search notes with fuzzy search",
			fields: fields{
				NoteRepo:      NewNoteRepositoryMock(),
				ConfigService: service.NewConfigService(),
				TitlesIDMap:   map[string]int{},
				Titles:        []string{},
				TitlesIDMutex: &sync.RWMutex{},
			},
			args: args{
				query:       "quote",
				fuzzySearch: true,
			},
			want: map[string]int{
				"00628e854bd9b090d494b49b8a2b4063f622411593b34fba27a041d5e1a8c8a7": 1,
				"82ade45b31514dc54ee6c48aaa20196d7cd8c1f85a24f800583e3b552a6bcdbb": 3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				NoteRepo:      tt.fields.NoteRepo,
				ConfigService: tt.fields.ConfigService,
				TitlesIDMap:   tt.fields.TitlesIDMap,
				Titles:        tt.fields.Titles,
				TitlesIDMutex: tt.fields.TitlesIDMutex,
			}
			got, err := ns.SearchNotes(tt.args.query, tt.args.fuzzySearch)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteServiceImpl.SearchNotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NoteServiceImpl.SearchNotes() = %v, want %v", got, tt.want)
			}
		})
	}
}
