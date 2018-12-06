package dataprovider

import (
	"github.com/gophersch/tlgo"
)

type APIRawData struct {
	Stops                  []tlgo.Stop
	Lines                  []tlgo.Line
	RoutesByLineID         map[string][]tlgo.Route
	RoutesDetailsByRouteID map[string]tlgo.RouteDetails
}

func GetAPIData() (APIRawData, error) {

	client := tlgo.NewClient()
	stops, err := client.ListStops()
	if err != nil {
		return APIRawData{}, err
	}

	lines, err := client.ListLines()
	if err != nil {
		return APIRawData{}, err
	}

	linesByRouteID := make(map[string]tlgo.Line)
	routesByLineID := make(map[string][]tlgo.Route)
	routeDetailsByRouteID := make(map[string]tlgo.RouteDetails)

	for _, line := range lines {
		routes, err := client.ListRoutes(line)
		if err != nil {
			return APIRawData{}, err
		}

		routesByLineID[line.ID] = []tlgo.Route{}

		for _, route := range routes {

			routesByLineID[line.ID] = append(routesByLineID[line.ID], route)
			linesByRouteID[route.ID] = line

			details, err := client.GetRouteDetails(route)
			if err != nil {
				return APIRawData{}, err
			}
			routeDetailsByRouteID[route.ID] = details
		}

	}

	return APIRawData{
		Stops:                  stops,
		Lines:                  lines,
		RoutesByLineID:         routesByLineID,
		RoutesDetailsByRouteID: routeDetailsByRouteID,
	}, nil

}
