package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	SetupCloseHandler()

	CreateMainWindow()
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
		Cleanup()
		os.Exit(0)
	}()
}

func Cleanup() {
	fmt.Println("Cleanup")
}
