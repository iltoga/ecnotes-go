package service_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/service"
	toml "github.com/pelletier/go-toml"
)

var testNote = &service.Note{
	ID:        1,
	Title:     "title1",
	Content:   "test content",
	CreatedAt: 1643614680013,
	UpdatedAt: 1643614680013,
}

type ConfigServiceMockImpl struct {
	Loaded  bool
	Globals map[string]string
	Config  map[string]string
}

type NoteRepositoryMockImpl struct {
	mockedNotes  []service.Note
	mockedTitles []string
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
		mockedTitles: []string{
			"Mandela quote",
			"The way to get started is to quit talking and begin doing",
			"Oprah Winfrey quote",
			"The best is yet to come, Jhon Lennon",
			"The future belongs to those who believe in the beauty of their dreams",
			"Eleanor Roosevelt",
		},
	}
}

// GetAllNotes ....
func (nsr *NoteRepositoryMockImpl) GetAllNotes() ([]service.Note, error) {
	return nsr.mockedNotes, nil
}

// GetNote ....
func (nsr *NoteRepositoryMockImpl) GetNote(id int) (*service.Note, error) {
	for _, note := range nsr.mockedNotes {
		if note.ID == id {
			return &note, nil
		}
	}
	return nil, errors.New(common.ERR_NOTE_NOT_FOUND)
}

// CreateNote ....
func (nsr *NoteRepositoryMockImpl) CreateNote(note *service.Note) error {
	nsr.mockedNotes = append(nsr.mockedNotes, *note)
	return nil
}

// UpdateNote ....
func (nsr *NoteRepositoryMockImpl) UpdateNote(note *service.Note) error {
	for i, n := range nsr.mockedNotes {
		if n.ID == note.ID {
			nsr.mockedNotes[i] = *note
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

func (nsr *NoteRepositoryMockImpl) NoteExists(id int) (bool, error) {
	for _, note := range nsr.mockedNotes {
		if note.ID == id {
			return true, nil
		}
	}
	return false, nil
}

func (nsr *NoteRepositoryMockImpl) GetIDFromTitle(title string) int {
	return int(cryptoUtil.IndexFromString(title))
}

type noteConfigServiceMockImpl struct {
	Config  map[string]string // configuration from config file
	Globals map[string]string // global variables (loaded in memory only)
	Loaded  bool
}

func (nsc *noteConfigServiceMockImpl) GetGlobal(key string) (string, error) {
	return nsc.Globals[key], nil
}

func (nsc *noteConfigServiceMockImpl) SetGlobal(key string, value string) {
	panic("not implemented") // TODO: Implement
}

func (nsc *noteConfigServiceMockImpl) GetConfig(key string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (nsc *noteConfigServiceMockImpl) SetConfig(key string, value string) error {
	panic("not implemented") // TODO: Implement
}

func (nsc *noteConfigServiceMockImpl) LoadConfig() error {
	nsc.Loaded = true
	return nil
}

func (nsc *noteConfigServiceMockImpl) ParseConfigTree(configTree *toml.Tree) {
	panic("not implemented") // TODO: Implement
}

func (nsc *noteConfigServiceMockImpl) SaveConfig() error {
	panic("not implemented") // TODO: Implement
}

func TestNoteServiceImpl_GetNote(t *testing.T) {
	type fields struct {
		Titles []string
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
				Titles: tt.fields.Titles,
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
		Titles []string
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
				Titles: tt.fields.Titles,
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

func TestNoteServiceImpl_EncryptNote(t *testing.T) {
	type fields struct {
		Titles        []string
		NoteRepo      service.NoteServiceRepository
		ConfigService service.ConfigService
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
		{
			name: "test",
			fields: fields{
				Titles:   []string{},
				NoteRepo: nil,
				ConfigService: &noteConfigServiceMockImpl{
					Globals: map[string]string{common.CONFIG_ENCRYPTION_KEY: "1234567890"},
					Loaded:  true,
				},
			},
			args: args{
				note: testNote,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				ConfigService: tt.fields.ConfigService,
				NoteRepo:      tt.fields.NoteRepo,
				Titles:        tt.fields.Titles,
			}
			if err := ns.EncryptNote(tt.args.note); (err != nil) != tt.wantErr {
				t.Errorf("service.NoteServiceImpl.EncryptNote() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := ns.DecryptNote(tt.args.note); err != nil {
				t.Errorf("service.NoteServiceImpl.DecryptNote() error = %v, wantErr %v", err, tt.wantErr)
				if !reflect.DeepEqual(tt.args.note, testNote) {
					t.Errorf("service.NoteServiceImpl.DecryptNote() = %v, want %v", tt.args.note, testNote)
				}
			}
		})
	}
}

func TestNoteServiceImpl_SearchNotes(t *testing.T) {
	noteRepositoryMock := NewNoteRepositoryMock()
	type fields struct {
		NoteRepo      service.NoteServiceRepository
		ConfigService service.ConfigService
		Titles        []string
	}
	type args struct {
		query       string
		fuzzySearch bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Search notes",
			fields: fields{
				NoteRepo: noteRepositoryMock,
				ConfigService: &noteConfigServiceMockImpl{
					Globals: map[string]string{common.CONFIG_ENCRYPTION_KEY: "1234567890"},
					Loaded:  true,
				},
				Titles: noteRepositoryMock.mockedTitles,
			},
			args: args{
				query:       "Mandela quote",
				fuzzySearch: false,
			},
			want: []string{
				"Mandela quote",
			},
			wantErr: false,
		},
		{
			name: "Search notes with fuzzy search",
			fields: fields{
				NoteRepo: noteRepositoryMock,
				ConfigService: &noteConfigServiceMockImpl{
					Globals: map[string]string{common.CONFIG_ENCRYPTION_KEY: "1234567890"},
					Loaded:  true,
				},
				Titles: noteRepositoryMock.mockedTitles,
			},
			args: args{
				query:       "quote",
				fuzzySearch: true,
			},
			want: []string{
				"Mandela quote",
				"Oprah Winfrey quote",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &service.NoteServiceImpl{
				NoteRepo:      tt.fields.NoteRepo,
				ConfigService: tt.fields.ConfigService,
				Titles:        tt.fields.Titles,
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
