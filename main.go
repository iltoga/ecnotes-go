package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2/app"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/iltoga/ecnotes-go/ui"
)

var (
	configService  service.ConfigService
	noteService    service.NoteService
	noteRepository service.NoteServiceRepository
	kvdbPath       string
	obs            = observer.NewObserver()
	defaultBucket  = "notes"
)

func init() {
	var err error
	configService = service.NewConfigService()
	if err = configService.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	kvdbPath, err = configService.GetConfig(common.CONFIG_KVDB_PATH)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	noteRepository, err = service.NewNoteServiceRepository(kvdbPath, defaultBucket)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	noteService = service.NewNoteService(noteRepository, configService, obs)
}

func main() {
	// TODO: delete this if not needed (fyne framework already takes care of keeping the app running till we close the main window)
	// SetupCloseHandler()

	fmt.Println("Starting...")
	// create a new ui
	appUI := ui.NewUI(app.NewWithID("ec-notes"), configService, noteService, obs)
	mainWindow := ui.NewMainWindow(appUI)

	// add listener to ui service to trigger note list widget update whenever the note title array changes
	obs.AddListener(observer.EVENT_UPDATE_NOTE_TITLES, mainWindow.UpdateNoteListWidget())
	obs.AddListener(observer.EVENT_CREATE_NOTE, mainWindow.UpdateNoteListWidget())
	obs.AddListener(observer.EVENT_UPDATE_NOTE, mainWindow.UpdateNoteListWidget())

	// start the ui

	// TODO: load some defaults from configuration?
	emptyOptions := make(map[string]interface{})
	mainWindow.CreateWindow("EcNotes", 800, 800, true, emptyOptions)
	appUI.Run()
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		cleanup()
		os.Exit(0)
	}()
}

func cleanup() {
	fmt.Println("Cleanup")
}
