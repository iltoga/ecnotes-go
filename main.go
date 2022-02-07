package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
)

var (
	configService  service.ConfigService
	noteService    service.NoteService
	noteRepository service.NoteServiceRepository
	kvdbPath       string
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
	noteService = service.NewNoteService(noteRepository, configService)
}

func main() {
	SetupCloseHandler()

	fmt.Println("Starting...")
	notes, err := noteService.GetNotes()
	if err != nil && err.Error() != common.ERR_BUCKET_EMPTY {
		fmt.Println(err)
		os.Exit(1)
	}
	if err.Error() == common.ERR_BUCKET_EMPTY {
		newNote := &service.Note{
			Title:   "Welcome to EcNotes",
			Content: "This is your first note.\n\nYou can edit it by clicking on the title.",
		}
		if err := noteService.CreateNote(newNote); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	for _, note := range notes {
		fmt.Printf("%+v\n", note)
	}
	// ui.CreateMainWindow()

	// TODO: move to service (config validation)
	// keyFile, err := configService.GetConfig("file_key")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// crtFile, err := configService.GetConfig("file_crt")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// fmt.Println("key file:", keyFile)
	// fmt.Println("cert file:", crtFile)
	// infinite loop to keep the program running
	for {
		select {}
	}
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
