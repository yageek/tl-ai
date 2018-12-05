package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/yageek/tl-ai/search"
)

const (
	FromToScheduleQuery     = "FromToScheduleQuery"
	SourceParameterKey      = "Origin"
	DestinationParameterKey = "Direction"
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

	if req.QueryResult.Intent.DisplayName == FromToScheduleQuery {

		sourceMap, hastSource := req.QueryResult.Parameters[SourceParameterKey].(map[string]interface{})
		destMap, hastDest := req.QueryResult.Parameters[DestinationParameterKey].(map[string]interface{})
		if !hastDest || !hastSource {
			http.Error(w, "Unkown Intent", http.StatusNotFound)
		}

		req := search.NewBFS(rawData.Stops, rawData.Lines, rawData.Routes)

		source := sourceMap["Stop"].(string)
		dest := destMap["Stop"].(string)

		steps, err := req.FindStopToStopPath(source, dest)
		if err == search.ErrNoPathFound {
			log.Println("No path found!")
			http.Error(w, "No path found", http.StatusNotFound)

		} else if err != nil {
			log.Println("Unknown error:", err)
			http.Error(w, "Unknown error!", http.StatusInternalServerError)
		} else {
			strBuf := bytes.NewBufferString("")
			for i, step := range steps {
				fmt.Fprintf(strBuf, "Step %d: %s\n", i, step)
			}
			io.Copy(w, strBuf)
		}

	} else {
		http.Error(w, "Unkown Intent", http.StatusNotFound)
	}

}
