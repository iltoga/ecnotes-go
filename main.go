package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"fyne.io/fyne/v2/app"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/provider"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/iltoga/ecnotes-go/ui"
	log "github.com/sirupsen/logrus"
)

var (
	configService  service.ConfigService
	noteService    service.NoteService
	noteRepository service.NoteServiceRepository
	kvdbPath       string
	obs            = observer.NewObserver()
	defaultBucket  = "notes"
	logger         *log.Logger
	quitSignalChan chan os.Signal
	cryptoService  service.CryptoServiceFactory
)

func init() {
	var err error

	// setup config service and load config
	if err = setupConfigService(); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// initialize logger
	if err = setupLogger(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = setupCryptoService(); err != nil {
		logger.Fatal(err)
	}

	// setup db connection
	if err = setupDb(cryptoService); err != nil {
		logger.Fatal(err)
	}

	// initialize external providers
	if err = setupProviders(); err != nil {
		logger.Fatal(err)
	}
}

// setupCryptoService setup an empty crypto service
func setupCryptoService() (err error) {
	cryptoService = &service.CryptoServiceFactoryImpl{}
	return
}

// setupConfigService setup config service
func setupConfigService() (err error) {
	configService, err = service.NewConfigService()
	if err != nil {
		return
	}
	if err = configService.LoadConfig(); err != nil {
		return
	}
	return
}

// setupDb setup the database
func setupDb(crypto service.CryptoServiceFactory) (err error) {
	kvdbPath, err = configService.GetConfig(common.CONFIG_KVDB_PATH)
	if err != nil {
		return
	}
	// TODO: pass env var to reset db (last parameter)
	noteRepository, err = service.NewNoteServiceRepository(kvdbPath, defaultBucket, false)
	if err != nil {
		return
	}
	noteService = service.NewNoteService(noteRepository, configService, obs, crypto)
	return
}

func main() {
	logger.Info("Starting...")
	quitSignalChan = make(chan os.Signal, 1)
	setupCloseHandler(quitSignalChan)

	// create a new ui
	appUI := ui.NewUI(app.NewWithID("ec-notes"), configService, noteService, obs)
	mainWindow := ui.NewMainWindow(appUI, cryptoService)

	// add listener to ui service to trigger note list widget update whenever the note title array changes
	obs.AddListener(observer.EVENT_UPDATE_NOTE_TITLES, mainWindow.UpdateNoteListWidget())

	// TODO: load some defaults from configuration?
	emptyOptions := make(map[string]interface{})
	mainWindow.CreateWindow("EcNotes", 600, 800, true, emptyOptions)
	mainWindow.GetWindow().SetOnClosed(
		func() {
			quitSignalChan <- syscall.SIGQUIT
		},
	)
	noteDetailWindow := ui.NewNoteDetailsWindow(appUI, new(model.Note))
	// update note window when clicking on update note button
	obs.AddListener(observer.EVENT_UPDATE_NOTE, noteDetailWindow.UpdateNoteDetailsWidget())
	// update note window after updating note
	obs.AddListener(observer.EVENT_UPDATE_NOTE_WINDOW, noteDetailWindow.UpdateNoteDetailsWidget())
	// update note window when clicking on create note button
	obs.AddListener(observer.EVENT_CREATE_NOTE, noteDetailWindow.UpdateNoteDetailsWidget())
	// update note window after creating note
	obs.AddListener(observer.EVENT_CREATE_NOTE_WINDOW, noteDetailWindow.UpdateNoteDetailsWidget())
	// TODO: for now selcting a note opens is in 'update mode' and we probably don't need this event.
	//       we should probably just add a button to toggle view/edit mode in the note details window
	obs.AddListener(observer.EVENT_VIEW_NOTE, noteDetailWindow.UpdateNoteDetailsWidget())

	noteDetailWindow.CreateWindow("testNoteDetails", 600, 800, false, make(map[string]interface{}))
	appUI.Run()
}

// setupLogger setup logrus logger with config
func setupLogger() (err error) {
	logger = log.New()
	logger.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	var level string
	// set log level according to config
	level, err = configService.GetConfig(common.CONFIG_LOG_LEVEL)
	if err != nil {
		return
	}
	levelParsed, err := log.ParseLevel(level)
	if err != nil {
		return
	}
	logger.SetLevel(levelParsed)

	// set logger path according to config
	logPath, err := configService.GetConfig(common.CONFIG_LOG_FILE_PATH)
	if err != nil {
		return
	}
	// if logpath directory does not exist, create it
	// note that logPath is the file name, not the directory
	logDir := filepath.Dir(logPath)
	if _, err = os.Stat(logDir); os.IsNotExist(err) {
		err = os.MkdirAll(logDir, 0755)
		if err != nil {
			return
		}
	}
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(mw)
	return
}

// setupProviders setup external providers
func setupProviders() (err error) {
	// if we have google_sheet_id in config, setup google sheets provider
	var (
		sheetID string
		gp      *provider.GoogleProvider
	)
	sheetID, err = configService.GetConfig(common.CONFIG_GOOGLE_SHEET_ID)
	if err != nil {
		// no google sheet id in config, skip
		logger.Info("No google_sheet_id in config.toml, skipping google provider setup")
		logger.Info("Google provider not activated: notes will NOT be synced to google sheet")
		return nil
	}
	credFilePath, _ := configService.GetConfig(common.CONFIG_GOOGLE_CREDENTIALS_FILE_PATH)
	gp, err = provider.NewGoogleProvider("notes", sheetID, credFilePath, logger, obs)
	if err != nil {
		return
	}
	logger.Info("Google provider activated: notes will be synced to google sheet")
	// add listeners to sync notes to google sheet when note is created/updated/deleted
	obs.AddListener(observer.EVENT_CREATE_NOTE, gp.UpdateNoteNotifier())
	obs.AddListener(observer.EVENT_UPDATE_NOTE, gp.UpdateNoteNotifier())
	obs.AddListener(observer.EVENT_DELETE_NOTE, gp.DeleteNoteNotifier())

	// sync notes from google sheet to db (two way sync)
	// TODO: make this optional and start sync after ui is loaded, so we can show a loading screen
	var dbNotes []model.Note
	var downloadedNotes []model.Note
	if dbNotes, err = noteService.GetNotes(); err == nil {
		logger.Info("Syncing notes from google sheets...")
		downloadedNotes, err = gp.SyncNotes(dbNotes)
		if err != nil {
			return
		}
		// if downloaded notes are not empty, update db with downloaded notes
		if len(downloadedNotes) > 0 {
			err = noteService.SaveEncryptedNotes(downloadedNotes)
			if err != nil {
				return
			}
		}
		logger.Info("Sync complete")
	}
	return
}

// setupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func setupCloseHandler(c chan os.Signal) {
	signal.Notify(quitSignalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		<-c
		cleanup()
	}()
}

func cleanup() {
	logger.Info("Cleanup...")
	// close logger file
	if logFile, ok := logger.Out.(*os.File); ok {
		if err := logFile.Close(); err != nil {
			fmt.Println("Error closing log file:", err)
		}
	}
}
