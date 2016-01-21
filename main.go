package main

import (
	"log"
	"net/http"

	api "github.com/dklassen/chamba/api"
)

func main() {
	api.GetDB()
	// server := http.Server{
	// 	Addr:    ":443",
	// 	Handler: api.Handlers(),
	// }
	server := http.Server{
		Addr:    ":8000",
		Handler: api.Handlers(),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to setup and listen to ssl connections: ", err)
	}
	// err := server.ListenAndServeTLS("chamba.pem", "chamba.key")
	// if err != nil {
	// 	log.Fatal("Failed to setup and listen to ssl connections: ", err)
	// }
}
