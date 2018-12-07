package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gophersch/tlgo"
	"github.com/gorilla/pat"
	"github.com/yageek/tl-ai/dataprovider"
	"github.com/yageek/tl-ai/storage"
)

var (
	store *storage.Store

	USERNAME string
	PASSWORD string

	tlClient *tlgo.Client
)

const (
	lastDataCache = "cache/apidata.gob"
)

func main() {

	// Configuration
	USERNAME = os.Getenv("USERNAME")
	PASSWORD = os.Getenv("PASSWORD")
	if USERNAME == "" || PASSWORD == "" {
		log.Fatalf("USERNAME and PASSWORD env variables should be set.")
	}

	// Load API data

	apiData := dataprovider.APIRawData{}
	b, err := hex.DecodeString(embeddedGOB)
	if err != nil {
		log.Panicf("Can not decode GOB: %s", err)
	}

	buff := bytes.NewBuffer(b)
	err = gob.NewDecoder(buff).Decode(&apiData)
	if err != nil {
		log.Fatalf("Can not load API data: %s\n", err)
	}

	// Store
	store = storage.NewStore(apiData)

	// Main client
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
