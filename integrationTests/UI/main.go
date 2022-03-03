package main

import (
	"encoding/hex"
	"errors"
	"sync"

	"fyne.io/fyne/v2/app"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/iltoga/ecnotes-go/ui"
)

var (
	enckey = "d8fe4aa6f1579d7bf71f43da885947b25892d4015d89a08ce153d38a72567c4d04151525d1c43720bc578e2f2be3b5ba364eb571be6af7240d3929cc6d145a2bb4efb8b2fbd698b05e8962a6c2327ed931d97244aa301663290ee1fefdb4c217f6f4d21e090e228d19dfbecea2f6a69caa9190349c7a4e449e90de79e460220c2e3cc9fb99788d0b4e8a3fe527d1aa5bcb6a8fb791a596e549a5046a157ba6b1414493b8c678512ff2663120225371aabc52ea3b38e947754eeae58e730c8a9655b15152f9a37a22ab66fa3de1de16daeb9be652eb61f66907c0a7cc9f314754d36bea97cf71e97d0eb4d645f314b8e82188c4e7e9dffada184d75183cc4b85b3eec8bc95d36bf6dd3a37d01a3d47c248ec11429a3686d281ac6bb90"
	decKey = "HsNARwACWCF22HKtZEALH8YkvfFlOqvGnu1O0RVlJGA97nD5JtkEp0gpV6Pvb19zKdRtKbQ1dS1oVCGBdItpppwaS1za3yA3iidSay0TM1Rzda1tI6xsV3djwJpAKniQNZBej1Zvw6ltAB5v6yOUdRESjEqvLyuP2UUm6dJCdAGwBR2Su1UP9v19n5wmz9g8n8OGzNfAg3S6JX1cK5M7wDcncNUd2UUzNlYU242kS1bPUYT5Lfn4qq9d4LjieAZ6"
	obs    = observer.NewObserver()
	testUI *ui.UImpl
)

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
				ID:        1761572867,
				Title:     "Mandela quote",
				Content:   "The greatest glory in living lies not in never falling, but in rising every time we fall. -Nelson Mandela",
				CreatedAt: 1644832171924,
				UpdatedAt: 1644832171924,
			},
			{
				ID:        3652028006,
				Title:     "The way to get started is to quit talking and begin doing",
				Content:   "Disney is the best company ever. - Walt Disney",
				Hidden:    true,
				CreatedAt: 1644832181924,
				UpdatedAt: 1644832181924,
			},
			{
				ID:        2903686729,
				Title:     "Oprah Winfrey quote",
				Content:   "If you look at what you have in life, you'll always have more. If you look at what you don't have in life, you'll never have enough",
				CreatedAt: 1644832171924,
				UpdatedAt: 1644832171924,
			},
			{
				ID:        566982022,
				Title:     "The best is yet to come, Jhon Lennon",
				Content:   "Life is what happens when you're busy making other plans",
				Hidden:    true,
				CreatedAt: 1644832271924,
				UpdatedAt: 1644832274924,
			},
			{
				ID:        1442556606,
				Title:     "The future belongs to those who believe in the beauty of their dreams",
				Content:   "Eleanor Roosevelt",
				CreatedAt: 1644832171924,
				UpdatedAt: 1644832171924,
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
	mocks := NewNoteRepositoryMock()
	nsr.mockedNotes = mocks.mockedNotes
	// encrypt all notes in nsr.mockedNotes
	for i, note := range nsr.mockedNotes {
		eContent, err := cryptoUtil.EncryptMessage([]byte(note.Content), decKey)
		if err != nil {
			return nil, err
		}
		nsr.mockedNotes[i].Content = hex.EncodeToString(eContent)
		if err != nil {
			return nil, err
		}
	}
	nsr.mockedTitles = mocks.mockedTitles
	obs.Notify(observer.EVENT_UPDATE_NOTE_TITLES, nsr.mockedTitles)
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
	nsr.mockedTitles = append(nsr.mockedTitles, note.Title)
	return nil
}

// UpdateNote ....
func (nsr *NoteRepositoryMockImpl) UpdateNote(note *model.Note) error {
	for i, n := range nsr.mockedNotes {
		note.ID = nsr.GetIDFromTitle(note.Title)
		if n.ID == note.ID {
			nsr.mockedNotes[i] = *note
			return nil
		}
	}
	obs.Notify(observer.EVENT_UPDATE_NOTE_TITLES, nsr.mockedTitles)
	return errors.New(common.ERR_NOTE_NOT_FOUND)
}

// DeleteNote ....
func (nsr *NoteRepositoryMockImpl) DeleteNote(id int) error {
	for i, n := range nsr.mockedNotes {
		n.ID = nsr.GetIDFromTitle(n.Title)
		if n.ID == id {
			nsr.mockedNotes = append(nsr.mockedNotes[:i], nsr.mockedNotes[i+1:]...)
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
	configService.Config[common.CONFIG_ENCRYPTION_KEY] = enckey
	noteRepoMocked := NewNoteRepositoryMock()

	noteService := &service.NoteServiceImpl{
		NoteRepo:      noteRepoMocked,
		ConfigService: configService,
		Observer:      obs,
	}
	// create a new ui
	testUI = ui.NewUI(app.NewWithID("testAPP"), configService, noteService, obs)
	// new crypto service
	kms := service.NewKeyManagementServiceAES()
	key, _ := hex.DecodeString(decKey)
	kms.ImportKey(key)
	cryptoService := service.NewCryptoServiceAES(kms)

	mainWindow := ui.NewMainWindow(testUI, cryptoService)

	// add listeners
	obs.AddListener(observer.EVENT_UPDATE_NOTE_TITLES, mainWindow.UpdateNoteListWidget())

	// run the ui
	mainWindow.CreateWindow("EcNotesTest", 800, 800, true, map[string]interface{}{
		common.OPT_WINDOW_ASPECT: common.WindowAspect_Normal,
	})
	noteDetailWindow := ui.NewNoteDetailsWindow(testUI, new(model.Note))
	obs.AddListener(observer.EVENT_UPDATE_NOTE, noteDetailWindow.UpdateNoteDetailsWidget())
	obs.AddListener(observer.EVENT_CREATE_NOTE, noteDetailWindow.UpdateNoteDetailsWidget())
	// TODO: for now selcting a note opens is in 'update mode' and we probably don't need this event.
	//       we should probably just add a button to toggle view/edit mode in the note details window
	obs.AddListener(observer.EVENT_VIEW_NOTE, noteDetailWindow.UpdateNoteDetailsWidget())

	noteDetailWindow.CreateWindow("testNoteDetails", 600, 400, false, make(map[string]interface{}))

	// add some random notes at time interval
	// go func() {
	// 	time.Sleep(time.Second * 10)
	// 	for {
	// 		time.Sleep(time.Millisecond * 2000)
	// 		note := model.Note{
	// 			Title:   "Random note " + strconv.Itoa(rand.Intn(10000)),
	// 			Content: "Random content",
	// 		}
	// 		noteService.CreateNote(&note)
	// 	}
	// }()

	testUI.Run()
}
