package service_test

import (
	"encoding/hex"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testNote = &model.Note{
		ID:        1,
		Title:     "title1",
		Content:   "test content",
		CreatedAt: 1643614680013,
		UpdatedAt: 1643614680013,
	}
	aesKeyTest         = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456"
	aesKeyTestBytes, _ = hex.DecodeString(aesKeyTest)
	_, _               = cryptoUtil.DecryptAES256(aesKeyTestBytes, []byte(aesKeyTest))
)

type CertServiceMockImpl struct {
	certs map[string]model.EncKey
}

// NewCertServiceMock ....
func NewCertServiceMock() *CertServiceMockImpl {
	return &CertServiceMockImpl{
		certs: map[string]model.EncKey{
			"testKey1": {
				Name: "testKey1",
				Key:  aesKeyTestBytes,
				Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
			},
		},
	}
}

// LoadCerts ....
func (cs *CertServiceMockImpl) LoadCerts(pwd string) error {
	return nil
}

// SaveCerts ....
func (cs *CertServiceMockImpl) SaveCerts(pwd string) error {
	return nil
}

// GetCert ....
func (cs *CertServiceMockImpl) GetCert(name string) (*model.EncKey, error) {
	if cert, ok := cs.certs[name]; ok {
		return &cert, nil
	}
	return nil, errors.New(common.ERR_CERT_NOT_FOUND)
}

// AddCert ....
func (cs *CertServiceMockImpl) AddCert(cert model.EncKey) error {
	cs.certs[cert.Name] = cert
	return nil
}

// RemoveCert ....
func (cs *CertServiceMockImpl) RemoveCert(name string) error {
	if _, ok := cs.certs[name]; ok {
		delete(cs.certs, name)
		return nil
	}
	return errors.New(common.ERR_CERT_NOT_FOUND)
}

// ObserverMockImpl ....
type ObserverMockImpl struct{}

// AddListener ....
func (obmock *ObserverMockImpl) AddListener(event observer.Event, listener observer.Listener) {
}

// Remove ....
func (obmock *ObserverMockImpl) Remove(event observer.Event) {
}

// Notify ....
func (obmock *ObserverMockImpl) Notify(event observer.Event, data interface{}, args ...interface{}) {
}

type capturedNotification struct {
	event observer.Event
	data  interface{}
	args  []interface{}
}

type capturingObserver struct {
	mu     sync.Mutex
	events []capturedNotification
}

func (co *capturingObserver) AddListener(event observer.Event, listener observer.Listener) {}
func (co *capturingObserver) Remove(event observer.Event)                                  {}
func (co *capturingObserver) Notify(event observer.Event, data interface{}, args ...interface{}) {
	copiedArgs := append([]interface{}(nil), args...)
	co.mu.Lock()
	defer co.mu.Unlock()
	co.events = append(co.events, capturedNotification{
		event: event,
		data:  data,
		args:  copiedArgs,
	})
}

func newTestNoteService(t *testing.T) (*service.NoteServiceImpl, *NoteRepositoryMockImpl) {
	cryptoSrv := service.NewCryptoServiceAES(service.NewKeyManagementServiceAES())
	require.NoError(t, cryptoSrv.GetKeyManager().ImportKey([]byte("1234567890123456"), "testKey1"))

	repo := &NoteRepositoryMockImpl{}
	obs := &capturingObserver{}
	ns := &service.NoteServiceImpl{
		NoteRepo: repo,
		Observer:  obs,
		Crypto: &service.CryptoServiceFactoryImpl{
			Srv: cryptoSrv,
		},
	}
	return ns, repo
}

// ConfigServiceMockImpl ....
type ConfigServiceMockImpl struct {
	Loaded  bool
	Globals map[string]string
	Config  map[string]string
}

// NoteRepositoryMockImpl ....
type NoteRepositoryMockImpl struct {
	mockedNotes  []model.Note
	mockedTitles []string
}

// NewNoteRepositoryMock ....
func NewNoteRepositoryMock() *NoteRepositoryMockImpl {
	return &NoteRepositoryMockImpl{
		mockedNotes: []model.Note{
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
		},
	}
}

// GetAllNotes ....
func (nsr *NoteRepositoryMockImpl) GetAllNotes() ([]model.Note, error) {
	return nsr.mockedNotes, nil
}

// GetNote ....
func (nsr *NoteRepositoryMockImpl) GetNote(id int) (*model.Note, error) {
	for _, note := range nsr.mockedNotes {
		if note.ID == id {
			return &note, nil
		}
	}
	return nil, errors.New(common.ERR_NOTE_NOT_FOUND)
}

// CreateNote ....
func (nsr *NoteRepositoryMockImpl) CreateNote(note *model.Note) error {
	nsr.mockedNotes = append(nsr.mockedNotes, *note)
	return nil
}

// UpdateNote ....
func (nsr *NoteRepositoryMockImpl) UpdateNote(note *model.Note) error {
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

// RenameNote ....
func (nsr *NoteRepositoryMockImpl) RenameNote(oldID int, note *model.Note) error {
	for i, n := range nsr.mockedNotes {
		if n.ID == oldID {
			nsr.mockedNotes[i] = *note
			return nil
		}
	}
	return errors.New(common.ERR_NOTE_NOT_FOUND)
}

// NoteExists ....
func (nsr *NoteRepositoryMockImpl) NoteExists(id int) (bool, error) {
	for _, note := range nsr.mockedNotes {
		if note.ID == id {
			return true, nil
		}
	}
	return false, nil
}

// GetIDFromTitle ....
func (nsr *NoteRepositoryMockImpl) GetIDFromTitle(title string) int {
	return int(cryptoUtil.IndexFromString(title))
}

type noteConfigServiceMockImpl struct {
	Config  map[string]string // configuration from config file
	Globals map[string]string // global variables (loaded in memory only)
	Loaded  bool
}

// GetResourcePath ....
func (nsc *noteConfigServiceMockImpl) GetResourcePath() string {
	return "./resource"
}

// GetGlobal ....
func (nsc *noteConfigServiceMockImpl) GetGlobal(key string) (string, error) {
	return nsc.Globals[key], nil
}

// GetGlobalBytes ....
func (nsc *noteConfigServiceMockImpl) GetGlobalBytes(key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// SetGlobal ....
func (nsc *noteConfigServiceMockImpl) SetGlobal(key string, value string) {
	panic("not implemented") // TODO: Implement
}

// GetGlobalBytes ....
func (nsc *noteConfigServiceMockImpl) SetGlobalBytes(key string, value []byte) {
	panic("not implemented") // TODO: Implement
}

// GetConfig ....
func (nsc *noteConfigServiceMockImpl) GetConfig(key string) (string, error) {
	panic("not implemented") // TODO: Implement
}

// GetConfigBytes ....
func (nsc *noteConfigServiceMockImpl) GetConfigBytes(key string) ([]byte, error) {
	panic("not implemented") // TODO: Implement
}

// SetConfig ....
func (nsc *noteConfigServiceMockImpl) SetConfig(key string, value string) error {
	panic("not implemented") // TODO: Implement
}

// SetConfigBytes
func (nsc *noteConfigServiceMockImpl) SetConfigBytes(key string, value []byte) error {
	panic("not implemented") // TODO: Implement
}

// LoadConfig ....
func (nsc *noteConfigServiceMockImpl) LoadConfig() error {
	nsc.Loaded = true
	return nil
}

// ParseConfigTree ....
func (nsc *noteConfigServiceMockImpl) ParseConfigTree(configTree *toml.Tree) {
	panic("not implemented") // TODO: Implement
}

// SaveConfig ....
func (nsc *noteConfigServiceMockImpl) SaveConfig() error {
	panic("not implemented") // TODO: Implement
}

// TestNoteServiceImpl_EncryptNote ....
func TestNoteServiceImpl_EncryptNote(t *testing.T) {
	cryptoSrv := service.NewCryptoServiceAES(service.NewKeyManagementServiceAES())
	cryptoSrv.GetKeyManager().ImportKey([]byte("1234567890123456"), "testKey1")
	type fields struct {
		Titles        []string
		NoteRepo      service.NoteServiceRepository
		ConfigService service.ConfigService
		Crypto        service.CryptoServiceFactory
		Observer      observer.Observer
	}
	type args struct {
		note *model.Note
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
					Config: map[string]string{common.CONFIG_CUR_ENCRYPTION_KEY_NAME: "testKey1"},
					Loaded: true,
				},
				Crypto: &service.CryptoServiceFactoryImpl{
					Srv: cryptoSrv,
				},
				Observer: &ObserverMockImpl{},
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
				Observer:      tt.fields.Observer,
				Crypto:        tt.fields.Crypto,
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

func TestNoteServiceImpl_CreateNote_EmitsSeparateSnapshots(t *testing.T) {
	ns, _ := newTestNoteService(t)
	obs := ns.Observer.(*capturingObserver)

	note := &model.Note{
		ID:      100,
		Title:   "Snapshot Note",
		Content: "plain content",
	}

	require.NoError(t, ns.CreateNote(note))

	obs.mu.Lock()
	events := append([]capturedNotification(nil), obs.events...)
	obs.mu.Unlock()

	var createEvent *capturedNotification
	for i := range events {
		if events[i].event == observer.EVENT_CREATE_NOTE {
			createEvent = &events[i]
			break
		}
	}
	require.NotNil(t, createEvent)

	decNote, ok := createEvent.data.(*model.Note)
	require.True(t, ok)
	savedNote, ok := createEvent.args[2].(*model.Note)
	require.True(t, ok)

	assert.False(t, decNote.Encrypted)
	assert.Equal(t, "plain content", decNote.Content)
	assert.True(t, savedNote.Encrypted)
	assert.NotEqual(t, "plain content", savedNote.Content)
	assert.Equal(t, note.Title, decNote.Title)
	assert.Equal(t, note.Title, savedNote.Title)
}

func TestNoteServiceImpl_UpdateNoteContent_And_GetNoteWithContent(t *testing.T) {
	ns, _ := newTestNoteService(t)

	note := &model.Note{
		Title:   "Update Title",
		Content: "Original content",
	}
	require.NoError(t, ns.CreateNote(note))

	loaded, err := ns.GetNoteWithContent(note.ID)
	require.NoError(t, err)
	assert.Equal(t, "Original content", loaded.Content)
	assert.False(t, loaded.Encrypted)

	loaded.Content = "Updated content"
	require.NoError(t, ns.UpdateNoteContent(loaded))

	updated, err := ns.GetNoteWithContent(note.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated content", updated.Content)
	assert.Equal(t, "Update Title", updated.Title)
}

func TestNoteServiceImpl_UpdateNoteTitle_RenamesAndUpdatesTitles(t *testing.T) {
	ns, _ := newTestNoteService(t)

	note := &model.Note{
		Title:   "Old Title",
		Content: "Some content",
	}
	require.NoError(t, ns.CreateNote(note))

	newID, err := ns.UpdateNoteTitle("Old Title", "New Title")
	require.NoError(t, err)
	assert.NotEqual(t, 0, newID)

	titles := ns.GetTitles()
	assert.Contains(t, titles, "New Title")
	assert.NotContains(t, titles, "Old Title")

	updated, err := ns.GetNoteWithContent(newID)
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "Some content", updated.Content)
}

func TestNoteServiceImpl_DeleteNote_RemovesNoteAndTitle(t *testing.T) {
	ns, repo := newTestNoteService(t)

	note := &model.Note{
		Title:   "Delete Me",
		Content: "Some content",
	}
	require.NoError(t, ns.CreateNote(note))

	require.NoError(t, ns.DeleteNote(note.ID))

	titles := ns.GetTitles()
	assert.NotContains(t, titles, "Delete Me")

	exists, err := repo.NoteExists(note.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestNoteServiceImpl_SaveEncryptedNotes_AppendsAndRefreshesTitles(t *testing.T) {
	ns, _ := newTestNoteService(t)

	notes := []model.Note{
		{ID: 200, Title: "Batch One", Content: "content one"},
		{ID: 201, Title: "Batch Two", Content: "content two"},
	}

	require.NoError(t, ns.SaveEncryptedNotes(notes))

	titles := ns.GetTitles()
	assert.Contains(t, titles, "Batch One")
	assert.Contains(t, titles, "Batch Two")

	saved, err := ns.GetNotes()
	require.NoError(t, err)
	assert.Len(t, saved, 2)
}

func TestNoteServiceImpl_ReEncryptNotes_MigratesEncryptedNotes(t *testing.T) {
	ns, _ := newTestNoteService(t)

	note := &model.Note{
		Title:   "Rotate Me",
		Content: "Top secret content",
	}
	require.NoError(t, ns.CreateNote(note))

	encryptedNotes, err := ns.GetNotes()
	require.NoError(t, err)
	require.Len(t, encryptedNotes, 1)
	oldEncryptedContent := encryptedNotes[0].Content

	newCert := model.EncKey{
		Name: "newKey",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  []byte("new-key-32-bytes-new-key-32-bytes"),
	}
	require.NoError(t, ns.ReEncryptNotes(encryptedNotes, newCert))

	rotated, err := ns.GetNotes()
	require.NoError(t, err)
	require.Len(t, rotated, 1)
	assert.Equal(t, "newKey", rotated[0].EncKeyName)
	assert.NotEqual(t, oldEncryptedContent, rotated[0].Content)

	decrypted, err := ns.GetNoteWithContent(note.ID)
	require.NoError(t, err)
	assert.Equal(t, "Top secret content", decrypted.Content)
}

// TestNoteServiceImpl_SearchNotes ....
func TestNoteServiceImpl_SearchNotes(t *testing.T) {
	noteRepositoryMock := NewNoteRepositoryMock()
	type fields struct {
		NoteRepo      service.NoteServiceRepository
		ConfigService service.ConfigService
		Titles        []string
		Observer      observer.Observer
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
					Config: map[string]string{common.CONFIG_CUR_ENCRYPTION_KEY_NAME: "testKey1"},
					Loaded: true,
				},
				Titles:   noteRepositoryMock.mockedTitles,
				Observer: &ObserverMockImpl{},
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
					Config: map[string]string{common.CONFIG_CUR_ENCRYPTION_KEY_NAME: "testKey1"},
					Loaded: true,
				},
				Titles:   noteRepositoryMock.mockedTitles,
				Observer: &ObserverMockImpl{},
			},
			args: args{
				query:       "mand",
				fuzzySearch: true,
			},
			want: []string{
				"Mandela quote",
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
				Observer:      tt.fields.Observer,
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
