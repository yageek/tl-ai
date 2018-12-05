package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yageek/tl-ai/search"
)

const (
	dialogFlowNextDepartureIntent = "NextDepartureQuery"
	LineNameKey                   = "line-name"
	StopOriginKey                 = "stop-origin"
	StopDestinationKey            = "stop-destination"
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resp)
}

func handleNextDepartureQuery(w http.ResponseWriter, f fullfillment) {

	parameters := f.QueryResult.Parameters

	stopOrigin, hasOrigin := parameters[StopOriginKey].(map[string]interface{})

	if !hasOrigin {
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	originValue, err := stopNameFromMap(stopOrigin)
	if err != nil {
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	stopDestination, hasDestination := parameters[StopDestinationKey].(map[string]interface{})
	line := parameters[LineNameKey]
	hasLine := line != ""

	// Now depending on the situation, we will be able to answer the user or not

	var destinationValue string
	// But first, we check that destination is not the same as start
	if hasDestination {
		destinationValue, err = stopNameFromMap(stopDestination)
		if err != nil {
			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
			return
		}

		if destinationValue == originValue {
			answer(w, "Il me semble que vous soyez déjà arriver.")
			return
		}
	}

	// Simple case, we perform a request with the correct values
	if hasDestination && hasLine {
		answer(w, fmt.Sprintf("Le prochain départ pour %s est 18:00", destinationValue))
	}

	// If line is not provided, we do a graph search to find the direction
	if hasDestination && !hasLine {

		steps, err := bfs.FindStopToStopPath(originValue, destinationValue)
		if err == search.ErrNoPathFound {
			log.Println("No path found!")
			answer(w, "Aucun bus partant dans cette direction n'a été trouvé.")

		} else if len(steps) < 1 || err != nil {
			log.Println("Unknown error:", err)
			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		} else {
			// We ensure that only one line is used. If not, we prefer to not determine the next stop
			startLine := steps[0].ByLine
			for _, iter := range steps {
				if iter.ByLine != startLine {
					msg := fmt.Sprintf("Les arrêts %s et %s ne se trouvent pas sur la même ligne. Je ne peux déterminer le chemin optimal pour le moment. Désolé.", originValue, destinationValue)
					answer(w, msg)
					return
				}
			}

			answer(w, fmt.Sprintf("Le prochain départ pour %s est 18:00", destinationValue))
			return
		}

	}

	// If no destination, we enumerate all the departures from the line at the given stop
	if !hasDestination && hasLine {
		answer(w, fmt.Sprintf("Les prochains départs pour %s sont à 18:00, 19:00", destinationValue))
	}

}
