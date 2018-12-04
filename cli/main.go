package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/gophersch/tlgo"
	"github.com/yageek/tl-ai/search"
)

const (
	lastDataCache = "cache/raw_data.gob"
)

var (
	rawData rawDataGob
)

type rawDataGob struct {
	Stops  []tlgo.Stop
	Lines  []tlgo.Line
	Routes map[string]tlgo.RouteDetails
}

func main() {

	if _, err := os.Stat(lastDataCache); os.IsNotExist(err) {
		log.Printf("No cache files. Loading from the API...")
		client := tlgo.NewClient()

		log.Printf("List stops...\n")
		stops, err := client.ListStops()
		if err != nil {
			log.Fatalf("Can not get stops: %v", err)
		}

		log.Printf("List lines...\n")
		lines, err := client.ListLines()
		if err != nil {
			log.Fatalf("Can not get lists: %v", err)
		}

		lineInfos := make(map[string]tlgo.RouteDetails)

		for _, line := range lines {
			log.Printf("\tList route for %s ...\n", line.Name)
			routes, err := client.ListRoutes(line)
			if err != nil {
				log.Printf("Can not fetch routes for %s: %v", line.ID, err)
			}

			for _, route := range routes {
				log.Printf("\tGet details for %s ...\n", route.ID)
				details, err := client.GetRouteDetails(route)
				if err != nil {
					log.Printf("Can not fetch routes details for %s: %v", route.ID, err)
				}
				lineInfos[line.ID] = details
			}
		}

		file, err := os.Create(lastDataCache)
		if err != nil {
			log.Fatalf("Can not create cache file: %v", err)
		}

		data := rawDataGob{
			Stops:  stops,
			Lines:  lines,
			Routes: lineInfos,
		}

		if err := gob.NewEncoder(file).Encode(&data); err != nil {
			log.Fatalf("Can not write file: %v", err)
		}

		rawData = data
	} else {
		log.Printf("Reading cache files ....")
		file, err := os.Open(lastDataCache)
		if err != nil {
			log.Fatalf("Can not cache file: %v", err)
		}

		err = gob.NewDecoder(file).Decode(&rawData)
		if err != nil {
			log.Fatalf("Can not read decoded file: %v", err)
		}
	}

	// Creating the tree
	request := search.NewBFS(rawData.Stops, rawData.Lines, rawData.Routes)
	steps, err := request.FindStopToStopPath("Sablons", "Longemalle")
	if err != nil {
		log.Printf("Error during the search: %v\n", err)
	}

	for _, step := range steps {
		fmt.Println(step)
	}

}
