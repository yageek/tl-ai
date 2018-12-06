package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gophersch/tlgo"
	"github.com/gorilla/pat"
)

var (
	rawData rawDataGob
	storage Storage

	USERNAME string
	PASSWORD string

	tlClient *tlgo.Client
)

func main() {

	// Configuration
	USERNAME = os.Getenv("USERNAME")
	PASSWORD = os.Getenv("PASSWORD")
	if USERNAME == "" || PASSWORD == "" {
		log.Fatalf("USERNAME and PASSWORD env variables should be set.")
	}

	// Storage
	storage = NewLocalDevStorage()
	data, err := storage.GetCacheData()
	if err != nil {
		log.Fatalf("Can not initialise storage: %s", err)
	}

	rawData = data

	tlClient = tlgo.NewClient()

	// Main app
	router := pat.New()

	router.Post("/dialogflow_interactions", basicAuth(USERNAME, PASSWORD, dialogFlowHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
