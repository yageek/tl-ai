package storage

import (
	"github.com/gophersch/tlgo"
	"github.com/yageek/tl-ai/dataprovider"
)

type LocalStorage struct {
	stops                  []tlgo.Stop
	lines                  []tlgo.Line
	routesByLineID         map[string][]tlgo.Route
	stopsByStopName        map[string]tlgo.Stop
	linesByLineID          map[string]tlgo.Line
	linesByRouteID         map[string]tlgo.Line
	linesByName            map[string]tlgo.Line
	routesDetailsByRouteID map[string]tlgo.RouteDetails
	routesByRouteID        map[string]tlgo.Route
}

func NewLocalStorage(data dataprovider.APIRawData) *LocalStorage {

	st := &LocalStorage{
		stops:                  data.Stops,
		lines:                  data.Lines,
		stopsByStopName:        map[string]tlgo.Stop{},
		routesByLineID:         data.RoutesByLineID,
		routesDetailsByRouteID: data.RoutesDetailsByRouteID,
		linesByLineID:          map[string]tlgo.Line{},
		linesByRouteID:         map[string]tlgo.Line{},
		linesByName:            map[string]tlgo.Line{},
		routesByRouteID:        map[string]tlgo.Route{},
	}

	// Build stop index
	for _, stop := range data.Stops {
		st.stopsByStopName[stop.Name] = stop
	}

	for _, line := range data.Lines {
		st.linesByLineID[line.ID] = line
		st.linesByName[line.Name] = line
	}
	// build lineRoute index
	for lineID, routes := range data.RoutesByLineID {

		for _, route := range routes {
			st.linesByRouteID[route.ID] = st.linesByLineID[lineID]
			st.routesByRouteID[route.ID] = route
		}
	}

	return st
}

func (s *LocalStorage) GetStops() ([]tlgo.Stop, error) {

	return s.stops, nil
}

func (s *LocalStorage) GetLines() ([]tlgo.Line, error) {
	return s.lines, nil
}

func (s *LocalStorage) GetRoutesForLineID(lineID string) ([]tlgo.Route, error) {

	routes, hasFound := s.routesByLineID[lineID]
	if hasFound {
		return routes, nil
	}

	return []tlgo.Route{}, ErrNotFound
}

func (s *LocalStorage) GetRoutesDetailsForRouteID(routeID string) (tlgo.RouteDetails, error) {

	details, hasDetails := s.routesDetailsByRouteID[routeID]
	if hasDetails {
		return details, nil
	}

	return tlgo.RouteDetails{}, ErrNotFound
}

func (s *LocalStorage) GetLineForRouteID(routeID string) (tlgo.Line, error) {

	line, hasLine := s.linesByRouteID[routeID]
	if hasLine {
		return line, nil
	}
	return tlgo.Line{}, ErrNotFound
}

func (s *LocalStorage) GetStopByName(name string) (tlgo.Stop, error) {
	stop, hasFound := s.stopsByStopName[name]
	if hasFound {
		return stop, nil
	}

	return tlgo.Stop{}, ErrNotFound
}

func (s *LocalStorage) GetLineByName(name string) (tlgo.Line, error) {

	line, hasLine := s.linesByName[name]
	if hasLine {
		return line, nil
	}
	return tlgo.Line{}, ErrNotFound
}
