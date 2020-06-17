package main

import (
	"log"
	"net/http"

	//"github.com/han-so1omon/graphtools/algorithms"
	"github.com/han-so1omon/graphtools/server"
)

//todo graphstore
func main() {
	log.Println("Starting graph app")

	// Initialize graph manager store
	store := server.InMemoryGraphStore{}

	// Get new router
	router := server.NewRouter(&store)

	// Serve routes
	http.ListenAndServe(":8900", router)
}
