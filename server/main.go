package main

import (
	"encoding/gob"
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
	store storage.Store

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
	apiData, err := getCacheData()
	if err != nil {
		log.Fatalf("Can not get API data: %s\n", err)
	}

	// Store
	store = storage.NewLocalStorage()
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

func clearCache() error {
	return os.Remove(lastDataCache)
}

func getCacheData() (dataprovider.APIRawData, error) {

	rawData := dataprovider.APIRawData{}
	if _, err := os.Stat(lastDataCache); os.IsNotExist(err) {
		log.Printf("No cache files. Loading from the API...")

		data, err := dataprovider.GetAPIData()
		if err != nil {
			return dataprovider.APIRawData{}, err
		}

		file, err := os.Create(lastDataCache)
		if err != nil {
			return dataprovider.APIRawData{}, err
		}

		if err := gob.NewEncoder(file).Encode(&data); err != nil {
			return dataprovider.APIRawData{}, err
		}

		rawData = data

	} else {
		log.Printf("Reading cache files ....")
		file, err := os.Open(lastDataCache)
		if err != nil {
			return dataprovider.APIRawData{}, err
		}

		err = gob.NewDecoder(file).Decode(&rawData)
		if err != nil {
			return dataprovider.APIRawData{}, err
		}
	}

	return rawData, nil
}
