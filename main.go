package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/ui"
)

var configService service.ConfigService

func init() {
	configService = service.NewConfigService()
	if err := configService.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	SetupCloseHandler()

	ui.CreateMainWindow()

	// TODO: move to service (config validation)
	keyFile, err := configService.GetConfig("file_key")
	if err != nil {
		fmt.Println(err)
	}
	crtFile, err := configService.GetConfig("file_crt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("key file:", keyFile)
	fmt.Println("cert file:", crtFile)
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
