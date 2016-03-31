package api

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

var (
	goenv string
)

func environment() string {
	if goenv == "" {
		goenv = os.Getenv("GOENV")
	}

	if goenv == "" {
		log.Fatal("GOENV environment variable is not set. Please set and re-launch")
	}
	return goenv
}
