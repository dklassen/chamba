package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	api "github.com/dklassen/chamba/api"
)

func main() {
	api.GetDB()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: api.Handlers(),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
