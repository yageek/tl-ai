package main

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/gophersch/tlgo"
)

type rawDataGob struct {
	Stops            map[string]*tlgo.Stop
	LineRoutesIndex  map[*tlgo.Route]*tlgo.Line
	RoutesLinesIndex map[*tlgo.Line][]*tlgo.Route
	Routes           map[*tlgo.Route]tlgo.RouteDetails
}

// Storage abstracts the source of the data
type Storage interface {
	GetCacheData() (rawDataGob, error)
	Clear() error
}

// LocalDevStorage is a way to keep a local cache
type LocalDevStorage struct {
	rawData rawDataGob
	client  *tlgo.Client
}

const lastDataCache = "cache/raw_data.gob"

// NewLocalDev returns a local cache storage
func NewLocalDevStorage() *LocalDevStorage {
	return &LocalDevStorage{
		rawData: rawDataGob{},
	}
}

func getAPIData() (rawDataGob, error) {

	client := tlgo.NewClient()
	stops, err := client.ListStops()
	if err != nil {
		return rawDataGob{}, err
	}

	stopsMap := make(map[string]*tlgo.Stop, len(stops))

	for i := range stops {
		stopsMap[stops[i].Name] = &stops[i]
	}

	lineRoutesIndex := make(map[*tlgo.Route]*tlgo.Line)
	routeLinesIndex := make(map[*tlgo.Line][]*tlgo.Route)

	log.Printf("List lines...\n")
	lines, err := client.ListLines()
	if err != nil {
		return rawDataGob{}, err
	}

	routesInfos := make(map[*tlgo.Route]tlgo.RouteDetails)

	for lineIndex := range lines {
		log.Printf("\tList route for %s ...\n", lines[lineIndex].Name)
		routes, err := client.ListRoutes(lines[lineIndex])
		if err != nil {
			return rawDataGob{}, err
		}

		routeLinesIndex[&lines[lineIndex]] = []*tlgo.Route{}

		for routeIndex := range routes {

			routeLinesIndex[&lines[lineIndex]] = append(routeLinesIndex[&lines[lineIndex]], &routes[routeIndex])
			lineRoutesIndex[&routes[routeIndex]] = &lines[lineIndex]
			log.Printf("\tGet details for %s ...\n", routes[routeIndex].ID)
			details, err := client.GetRouteDetails(routes[routeIndex])
			if err != nil {
				return rawDataGob{}, err
			}
			routesInfos[&routes[routeIndex]] = details

		}

	}
	data := rawDataGob{
		Stops:            stopsMap,
		LineRoutesIndex:  lineRoutesIndex,
		RoutesLinesIndex: routeLinesIndex,
		Routes:           routesInfos,
	}

	return data, nil

}
func (l *LocalDevStorage) Clear() error {
	return os.Remove(lastDataCache)
}

func (l *LocalDevStorage) GetCacheData() (rawDataGob, error) {
	if _, err := os.Stat(lastDataCache); os.IsNotExist(err) {
		log.Printf("No cache files. Loading from the API...")

		data, err := getAPIData()
		if err != nil {
			return rawDataGob{}, err
		}

		file, err := os.Create(lastDataCache)
		if err != nil {
			return rawDataGob{}, err
		}

		if err := gob.NewEncoder(file).Encode(&data); err != nil {
			return rawDataGob{}, err
		}

		l.rawData = data

	} else {
		log.Printf("Reading cache files ....")
		file, err := os.Open(lastDataCache)
		if err != nil {
			return rawDataGob{}, err
		}

		err = gob.NewDecoder(file).Decode(&l.rawData)
		if err != nil {
			return rawDataGob{}, err
		}
	}

	return l.rawData, nil
}
