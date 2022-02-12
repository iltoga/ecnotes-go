package main

import (
	"errors"
	"sync"

	"fyne.io/fyne/v2/app"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/iltoga/ecnotes-go/ui"
)

var (
	encKey = "d8fe4aa6f1579d7bf71f43da885947b25892d4015d89a08ce153d38a72567c4d04151525d1c43720bc578e2f2be3b5ba364eb571be6af7240d3929cc6d145a2bb4efb8b2fbd698b05e8962a6c2327ed931d97244aa301663290ee1fefdb4c217f6f4d21e090e228d19dfbecea2f6a69caa9190349c7a4e449e90de79e460220c2e3cc9fb99788d0b4e8a3fe527d1aa5bcb6a8fb791a596e549a5046a157ba6b1414493b8c678512ff2663120225371aabc52ea3b38e947754eeae58e730c8a9655b15152f9a37a22ab66fa3de1de16daeb9be652eb61f66907c0a7cc9f314754d36bea97cf71e97d0eb4d645f314b8e82188c4e7e9dffada184d75183cc4b85b3eec8bc95d36bf6dd3a37d01a3d47c248ec11429a3686d281ac6bb90"
	decKey = "HsNARwACWCF22HKtZEALH8YkvfFlOqvGnu1O0RVlJGA97nD5JtkEp0gpV6Pvb19zKdRtKbQ1dS1oVCGBdItpppwaS1za3yA3iidSay0TM1Rzda1tI6xsV3djwJpAKniQNZBej1Zvw6ltAB5v6yOUdRESjEqvLyuP2UUm6dJCdAGwBR2Su1UP9v19n5wmz9g8n8OGzNfAg3S6JX1cK5M7wDcncNUd2UUzNlYU242kS1bPUYT5Lfn4qq9d4LjieAZ6"
	obsrv  = observer.NewObserver()
	testUI *ui.UImpl
)

// NoteRepositoryMockImpl ....
type NoteRepositoryMockImpl struct {
	mockedNotes  []service.Note
	mockedTitles []string
}

// NewNoteRepositoryMock ....
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
	mocks := NewNoteRepositoryMock()
	nsr.mockedNotes = mocks.mockedNotes
	// encrypt all notes in nsr.mockedNotes
	var err error
	for i, note := range nsr.mockedNotes {
		nsr.mockedNotes[i].Content, err = cryptoUtil.EncryptMessage(note.Content, decKey)
		if err != nil {
			return nil, err
		}
	}
	nsr.mockedTitles = mocks.mockedTitles
	obsrv.Notify(observer.EVENT_UPDATE_NOTE_TITLES, nsr.mockedTitles)
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
	nsr.mockedTitles = append(nsr.mockedTitles, note.Title)
	obsrv.Notify(observer.EVENT_UPDATE_NOTE_TITLES, nsr.mockedTitles)
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
	obsrv.Notify(observer.EVENT_UPDATE_NOTE_TITLES, nsr.mockedTitles)
	return errors.New(common.ERR_NOTE_NOT_FOUND)
}

// DeleteNote ....
func (nsr *NoteRepositoryMockImpl) DeleteNote(id int) error {
	for i, n := range nsr.mockedNotes {
		if n.ID == id {
			nsr.mockedNotes = append(nsr.mockedNotes[:i], nsr.mockedNotes[i+1:]...)
			for j, t := range nsr.mockedTitles {
				if t == n.Title {
					nsr.mockedTitles = append(nsr.mockedTitles[:j], nsr.mockedTitles[j+1:]...)
					break
				}
			}
			obsrv.Notify(observer.EVENT_UPDATE_NOTE_TITLES, nsr.mockedTitles)
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

// main this main mocks db service and runs the UI
func main() {
	configService := &service.ConfigServiceImpl{
		ResourcePath: "./integrationTests/mainWindow/testResources",
		// ResourcePath: "./testResources",
		Config:     make(map[string]string),
		Globals:    make(map[string]string),
		Loaded:     true,
		ConfigMux:  &sync.RWMutex{},
		GlobalsMux: &sync.RWMutex{},
	}
	// mocked configuration (no need to load from file)
	// if err := configService.LoadConfig(); err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	configService.Config[common.CONFIG_ENCRYPTION_KEY] = encKey
	noteRepoMocked := NewNoteRepositoryMock()

	noteService := &service.NoteServiceImpl{
		NoteRepo:      noteRepoMocked,
		ConfigService: configService,
		Observer:      obsrv,
	}
	// create a new ui
	testUI = ui.NewUI(app.NewWithID("testAPP"), configService, noteService)

	// add listeners
	obsrv.AddListener(observer.EVENT_UPDATE_NOTE_TITLES, testUI.UpdateNoteListWidget())

	// add some random notes at time interval
	// go func() {
	// 	time.Sleep(time.Second * 10)
	// 	for {
	// 		time.Sleep(time.Millisecond * 2000)
	// 		note := service.Note{
	// 			Title:   "Random note " + strconv.Itoa(rand.Intn(10000)),
	// 			Content: "Random content",
	// 		}
	// 		noteService.CreateNote(&note)
	// 	}
	// }()
	testUI.CreateMainWindow()
}
