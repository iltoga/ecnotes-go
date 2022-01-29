package main

import (
	"fmt"
	"os"
	"sync"

	"fyne.io/fyne/app"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/ui"
)

// main function to be called from the test
func main() {
	configService := &service.ConfigServiceImpl{
		ResourcePath: "./integrationTests/mainWindow/testResources",
		Config:       make(map[string]string),
		Globals:      make(map[string]string),
		Loaded:       false,
		ConfigMux:    &sync.RWMutex{},
		GlobalsMux:   &sync.RWMutex{},
	}
	if err := configService.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// create a new ui
	testUI := ui.NewUI(app.New(), configService)
	testUI.CreateMainWindow()
}
