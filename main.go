package main

import (
	"log"
	"net/http"

	//"github.com/han-so1omon/graphtools/algorithms"
	"github.com/han-so1omon/graphtools/server"
)

func main() {
	log.Println("Starting graph app")
	// Get new router
	router := server.NewRouter()
	// Serve routes
	http.ListenAndServe(":8900", router)
}
