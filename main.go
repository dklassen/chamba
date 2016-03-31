package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/dklassen/chamba/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	api "github.com/dklassen/chamba/api"
)

func main() {
	log.Info("Chamba application started: ", time.Now())
	api.GetDB()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: api.Handlers(),
	}
	log.Info("listening on port:", port)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
