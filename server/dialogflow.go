package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gophersch/tlgo"
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
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	json.NewEncoder(w).Encode(&resp)
}

func handleNextDepartureQuery(w http.ResponseWriter, f fullfillment) {

	log.Printf("Next departure query...\n")
	parameters := f.QueryResult.Parameters

	stopOrigin, hasOrigin := parameters[StopOriginKey].(map[string]interface{})

	if !hasOrigin {
		log.Printf("The origin information has not been provided by the bot\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	originValue, err := stopNameFromMap(stopOrigin)
	if err != nil {
		log.Printf("The origin value has not been provided\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	originStop, err := store.GetStopByName(originValue)
	if err != nil {
		log.Printf("The origin value has not been found in the index\n")
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	stopDestination, hasDestination := parameters[StopDestinationKey].(map[string]interface{})
	line, hasLine := parameters[LineNameKey].(string)
	hasLine = hasLine && line != ""

	// Now depending on the situation, we will be able to answer the user or not

	var destinationValue string
	// But first, we check that destination is not the same as start
	if hasDestination {
		destinationValue, err = stopNameFromMap(stopDestination)
		if err != nil {
			log.Printf("The destination value has not been provided\n")
			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
			return
		}

		if destinationValue == originValue {
			answer(w, "Il me semble que vous soyez déjà arriver.")
			return
		}
	}

	var lineValue tlgo.Line
	if hasLine {
		// Look for line

		line, err := store.GetLineByName(line)
		if err != nil {
			log.Printf("Line node not found\n")
			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
			return
		}

		lineValue = line
	}

	// Simple case, we perform a request with the correct values
	if hasDestination && hasLine {
		fmt.Printf("Get schedule from start, destination and line\n")
		// // Look for line
		// var line *tlgo.Line
		// for _, v := range rawData.LineRoutesIndex {
		// 	if v.Name == line.Name {
		// 		line = v
		// 		break
		// 	}
		// }

		// if line == nil {
		// 	answer(w, "Une erreur est survenue sur nos serveurs. Veuillez-nous excuser pour ce contre-temps")
		// 	return
		// }

		// // Look for correct sense

		// answerNextSchedule(w, originStop, nil, line)
		answer(w, "Cette fonctionnalité n'est pas encore implémentée")
		return
	}

	// If line is not provided, we do a graph search to find the direction
	if hasDestination && !hasLine {
		fmt.Printf("Get schedule from start and destination\n")

		bfs, err := search.NewBFS(store)
		log.Printf("bfs created")

		if err != nil {
			log.Printf("Error during query creation: %s", err)
			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
			return
		}

		steps, err := bfs.FindStopToStopPath(originValue, destinationValue)

		if err == search.ErrNoPathFound {
			log.Println("No path found!")
			answer(w, "Aucun bus partant dans cette direction n'a été trouvé.")
			return
		}

		log.Printf("Travel will requires %d step(s)\n", len(steps))
		if len(steps) < 1 || err != nil {

			if err != nil {
				log.Println("Unknown error:", err)
			} else {
				log.Println("No journeys has been returned.")
			}

			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
			return
		}

		// We ensure that only one line is used. If not, we prefer to not determine the next stop
		lineName := steps[0].Line.Name
		for _, iter := range steps {
			if iter.Line.Name != lineName {
				msg := fmt.Sprintf("Les arrêts %s et %s ne se trouvent pas sur la même route. Je ne peux déterminer le chemin optimal pour le moment. Désolé.", originValue, destinationValue)
				answer(w, msg)
				return
			}
		}

		answerNextSchedule(w, steps[0].Stop.Name, steps[0].RouteID, steps[0].Line)
		return

	}

	// If no destination, we enumerate all the departures from the line at the given stop
	if !hasDestination && hasLine {
		fmt.Printf("Get schedule from start and line\n")
		routes, err := store.GetRoutesForLineID(lineValue.ID)
		if err != nil {
			log.Printf("Can not get routes from line: %v", err)
			answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
			return
		}

		allDepartures := make([]departure, 1)
		for _, r := range routes {
			departures, err := getNextDeparture(originStop.Name, r.ID, lineValue.ID)

			if err != nil {
				log.Printf("Impossible to get next departures: %s\n", err)
				continue
			}
			allDepartures = append(allDepartures, departures...)
		}

		msg := fmt.Sprintf("Les prochains départs pour la ligne %s sont:\n", line)
		for _, dpt := range allDepartures {
			msg += fmt.Sprintf("%s en direction de %s", dpt.displaytext, dpt.routeID)
		}
		answer(w, msg)
	}

}

type departure struct {
	routeID     string
	lineID      string
	displaytext string
	date        time.Time
}

func getNextDeparture(stopName, routeID, lineID string) ([]departure, error) {
	journeys, err := tlClient.ListStopDepartures(routeID, lineID, time.Now(), false)
	if err != nil {
		return []departure{}, err
	}

	if len(journeys) < 1 {
		return []departure{}, nil
	}

	// Search all stops from origin
	departures := make([]departure, 1)

	for _, j := range journeys {
		for _, s := range j.Stops {
			if s.Name == stopName {
				d := departure{routeID, lineID, j.DisplayTime, j.Time}
				departures = append(departures, d)
			}
		}
	}
	return departures, nil
}

func answerNextSchedule(w http.ResponseWriter, stopName, routeID string, line tlgo.Line) {

	departures, err := getNextDeparture(stopName, routeID, line.ID)

	if err != nil {
		log.Println("TLAPI get schedules error:", err)
		answer(w, "Une erreur est survenue sur nos serveurs. Veuillez nous excuser pour ce contre-temps.")
		return
	}

	if len(departures) < 1 {
		msg := fmt.Sprintf("Aucun départ n'a été trouvé sur la ligne %s en direction de %s", line.Name, stopName)
		answer(w, msg)
	}

	msg := fmt.Sprintf("Les prochains départs pour le bus %s sont:\n", line.Name)
	for _, departure := range departures {
		msg += fmt.Sprintf("%s", departure.displaytext)
	}

	answer(w, msg)
}
