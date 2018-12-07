package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gophersch/tlgo"
)

const (
	dialogFlowNextDepartureIntent = "NextDepartureQuery"
	LineNameKey                   = "line-name"
	StopOriginKey                 = "stop-origin"
	StopDirectionKey              = "stop-direction"
)

type fullfillment struct {
	QueryResult queryResult `json:"queryResult"`
}
type queryResult struct {
	Query              string                 `json:"queryText"`
	Parameters         map[string]interface{} `json:"parameters"`
	AllRequiredPresent bool                   `json:"allRequiredParamsPresent"`
	Intent             intent                 `json:"intent"`
	Confidence         float32                `json:"intentDetectionConfidence"`
}

type intent struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type fullFillementResponse struct {
	Text string `json:"fulfillmentText"`
}

// indexHandler responds to requests with our greeting.
func dialogFlowHandler(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	req := fullfillment{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.QueryResult.Intent.DisplayName == dialogFlowNextDepartureIntent {
		handleNextDepartureQuery(w, req)
	} else {
		http.Error(w, "Unkown Intent", http.StatusNotFound)
	}

}

func stopNameFromMap(m map[string]interface{}) (string, error) {

	key, hasKey := m["stop-name"].(string)
	if !hasKey {
		return "", errors.New("Entiity does not seem to be a stop")
	}

	return key, nil
}

func answer(w http.ResponseWriter, mesg string) {

	resp := fullFillementResponse{Text: mesg}
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	json.NewEncoder(w).Encode(&resp)
}

func handleNextDepartureQuery(w http.ResponseWriter, f fullfillment) {

	log.Printf("Next departure query...\n")
	parameters := f.QueryResult.Parameters

	// Get origin
	stopOriginMap, hasOrigin := parameters[StopOriginKey].(map[string]interface{})

	if !hasOrigin {
		log.Printf("The origin information has not been provided by the bot\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	stopOriginName, err := stopNameFromMap(stopOriginMap)
	if err != nil {
		log.Printf("The origin value has not been provided\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	originStop, err := store.GetStopByName(stopOriginName)
	if err != nil {
		log.Printf("The origin value has not been found in the index\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	// Get direction
	stopDirectionMap, hasDirection := parameters[StopDirectionKey].(map[string]interface{})
	if !hasDirection {
		log.Printf("The direction value has not been found in the index\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	stopDirectionName, err := stopNameFromMap(stopDirectionMap)
	if err != nil {
		log.Printf("The direction value has not been provided\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	lineName, hasLine := parameters[LineNameKey].(string)
	if !hasLine {
		log.Printf("The line value has not been found in the index\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	// We look for the line in the system
	line, err := store.GetLineByName(lineName)
	if err != nil {
		log.Printf("The line %s has not been found in the store: %v\n", lineName, err)
		answer(w, fmt.Sprintf("Je n'arrive pas à identifier la ligne correspondant à %s dans mon système.", lineName))
		return
	}

	// We ensure lines holds the start stop
	isStartStopInLine := false
	for _, name := range originStop.LinesShortName {

		if line.ShortName == name {
			isStartStopInLine = true
		}
	}

	if !isStartStopInLine {
		answer(w, fmt.Sprintf("La ligne %s ne semble pas s'arrêter à l'arrêt %s", line.ShortName, originStop.Name))
		return
	}
	// Then we try to find a routes for the corresponding destination
	routes, err := store.GetRoutesForLineID(line.ID)
	if err != nil {
		log.Printf("The routes has not been found in the store: %v\n", err)
		answer(w, fmt.Sprintf("Je ne peux fournir des informations concernant la ligne %s pour le moment.", line.Name))
		return
	}

	for _, route := range routes {

		if route.CityDestinationStopName == stopDirectionName {
			answerNextSchedule(w, originStop, route.ID, stopDirectionName, line)
			return
		}
	}

	answer(w, fmt.Sprintf("Aucune route en direction de %s n'a été trouvée pour la ligne %s", stopDirectionName, line.Name))

}

func getNextDeparture(stop tlgo.Stop, lineID string) ([]tlgo.Journey, error) {
	return tlClient.ListStopDepartures(stop.ID, lineID, time.Now(), false)
}

func answerNextSchedule(w http.ResponseWriter, stop tlgo.Stop, routeID, direction string, line tlgo.Line) {

	departures, err := getNextDeparture(stop, line.ID)

	if err != nil {
		log.Println("TLAPI get schedules error:", err)
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	if len(departures) < 1 {
		msg := fmt.Sprintf("Aucun départ n'a été trouvé sur la ligne %s en direction de %s", line.ShortName, direction)
		answer(w, msg)
	}

	msg := fmt.Sprintf("Le prochain bus %s en direction de %s partira ", line.ShortName, direction)

	departure := departures[0]

	var waiting string
	if departure.WaitingTime.Minutes() < 0 {
		waiting = "dans moins d'une minute."
	} else if departure.WaitingTime.Hours() < 1.0 {
		minutes := int(math.Ceil(departure.WaitingTime.Minutes()))
		if minutes > 1 {
			waiting = fmt.Sprintf("dans %d minutes environ", minutes)
		} else {
			waiting = fmt.Sprintf("dans %d minutes environ", minutes)
		}

	} else {
		waiting = "dans plus d'une heure."
	}

	msg += waiting

	answer(w, msg)
}
