package common

import (
	"log"
	"os/user"
)

// GetUserHomeDir returns the user's home directory.
func GetUserHomeDir() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return user.HomeDir
}
