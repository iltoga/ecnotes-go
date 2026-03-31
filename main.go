package main

import (
	"context"
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

func main() {
	var err error

	// setup config service and load config
	configService, err := setupConfigService()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// load certificates to encrypt/decrypt messages
	certService, err := setupCerts(configService)
	if err != nil {
		fmt.Println("Error loading certificates:", err)
		os.Exit(1)
	}

	// initialize logger
	logger, err := setupLogger(configService)
	if err != nil {
		fmt.Println("Error setting up logger:", err)
		os.Exit(1)
	}

	cryptoService, err := setupCryptoService()
	if err != nil {
		logger.Fatal(err)
	}

	obs := observer.NewObserver()

	// setup db connection
	noteService, err := setupDb(configService, cryptoService, obs)
	if err != nil {
		logger.Fatal(err)
	}

	appCtx, cancel := context.WithCancel(context.Background())

	logger.Info("Starting...")
	setupCloseHandler(cancel, logger)

	// initialize external providers
	// We run this in a goroutine so it doesn't block the UI
	go func() {
		if err := setupProviders(appCtx, configService, noteService, obs, logger); err != nil {
			logger.Errorf("Error setting up providers: %v", err)
		}
	}()

	// create a new ui
	appUI := ui.NewUI(app.NewWithID("ec-notes"), configService, noteService, certService, obs)
	mainWindow := ui.NewMainWindow(appUI, cryptoService)

	// add listener to ui service to trigger note list widget update whenever the note title array changes
	obs.AddListener(observer.EVENT_UPDATE_NOTE_TITLES, mainWindow.UpdateNoteListWidget())

	// TODO: load some defaults from configuration?
	emptyOptions := make(map[string]interface{})
	mainWindow.CreateWindow("EcNotes", 600, 800, true, emptyOptions)
	mainWindow.GetWindow().SetOnClosed(
		func() {
			quitChan := make(chan os.Signal, 1)
			quitChan <- syscall.SIGQUIT
			setupCloseHandler(cancel, logger) // this will trigger cleanup
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

// setupCryptoService setup an empty crypto service
func setupCryptoService() (service.CryptoServiceFactory, error) {
	cryptoService := &service.CryptoServiceFactoryImpl{}
	return cryptoService, nil
}

// setupConfigService setup config service
func setupConfigService() (service.ConfigService, error) {
	configService, err := service.NewConfigService()
	if err != nil {
		return nil, err
	}
	if err = configService.LoadConfig(); err != nil {
		return nil, err
	}
	return configService, nil
}

// loadKeys loads the keys from the key_store.json file
func setupCerts(configService service.ConfigService) (service.CertService, error) {
	// get the key store path from the config
	keyFilePath, err := configService.GetConfig(common.CONFIG_KEY_FILE_PATH)
	if err != nil {
		return nil, err
	}
	certService := service.NewCertService(keyFilePath)
	return certService, nil
}

// setupDb setup the database
func setupDb(configService service.ConfigService, crypto service.CryptoServiceFactory, obs observer.Observer) (service.NoteService, error) {
	kvdbPath, err := configService.GetConfig(common.CONFIG_KVDB_PATH)
	if err != nil {
		return nil, err
	}
	defaultBucket := "notes"
	// TODO: pass env var to reset db (last parameter)
	noteRepository, err := service.NewNoteServiceRepository(kvdbPath, defaultBucket, false)
	if err != nil {
		return nil, err
	}
	noteService := service.NewNoteService(noteRepository, configService, obs, crypto)
	return noteService, nil
}

// setupLogger setup logrus logger with config
func setupLogger(configService service.ConfigService) (*log.Logger, error) {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	var level string
	// set log level according to config
	level, err := configService.GetConfig(common.CONFIG_LOG_LEVEL)
	if err != nil {
		return nil, err
	}
	levelParsed, err := log.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(levelParsed)

	// set logger path according to config
	logPath, err := configService.GetConfig(common.CONFIG_LOG_FILE_PATH)
	if err != nil {
		return nil, err
	}
	// if logpath directory does not exist, create it
	// note that logPath is the file name, not the directory
	logDir := filepath.Dir(logPath)
	if _, err = os.Stat(logDir); os.IsNotExist(err) {
		err = os.MkdirAll(logDir, 0755)
		if err != nil {
			return nil, err
		}
	}
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(mw)
	return logger, nil
}

// setupProviders setup external providers
func setupProviders(ctx context.Context, configService service.ConfigService, noteService service.NoteService, obs observer.Observer, logger *log.Logger) error {
	// if we have google_sheet_id in config, setup google sheets provider
	var (
		sheetID string
		gp      *provider.GoogleProvider
		err     error
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
		return err
	}
	logger.Info("Google provider activated: notes will be synced to google sheet")
	
	// Start the provider background worker
	gp.InitWorker(ctx)

	// add listeners to sync notes to google sheet when note is created/updated/deleted
	obs.AddListener(observer.EVENT_CREATE_NOTE, gp.UpdateNoteNotifier())
	obs.AddListener(observer.EVENT_UPDATE_NOTE, gp.UpdateNoteNotifier())
	obs.AddListener(observer.EVENT_DELETE_NOTE, gp.DeleteNoteNotifier())

	logger.Info("Syncing notes from google sheets...")
	var dbNotes []model.Note
	var downloadedNotes []model.Note
	if dbNotes, err = noteService.GetNotes(); err == nil {
		downloadedNotes, err = gp.SyncNotes(ctx, dbNotes)
		if err != nil {
			logger.Errorf("Error syncing notes from google sheets: %v", err)
			return err
		}
		// if downloaded notes are not empty, update db with downloaded notes
		if len(downloadedNotes) > 0 {
			err = noteService.SaveEncryptedNotes(downloadedNotes)
			if err != nil {
				return err
			}
		}
		logger.Info("Sync complete")
	} else {
		logger.Errorf("Error fetching local notes for sync: %v", err)
	}
	return nil
}

// setupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func setupCloseHandler(cancel context.CancelFunc, logger *log.Logger) {
	quitSignalChan := make(chan os.Signal, 1)
	signal.Notify(quitSignalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		<-quitSignalChan
		cancel()
		cleanup(logger)
	}()
}

func cleanup(logger *log.Logger) {
	if logger != nil {
		logger.Info("Cleanup...")
		// close logger file
		if logFile, ok := logger.Out.(*os.File); ok {
			if err := logFile.Close(); err != nil {
				fmt.Println("Error closing log file:", err)
			}
		}
	}
	os.Exit(0)
}

