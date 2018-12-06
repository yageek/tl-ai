package storage

import (
	"errors"

	"github.com/gophersch/tlgo"
)

var (
	ErrNotFound = errors.New("Element not found")
)

// Store represents a general store interface
type Store interface {
	GetStops() ([]tlgo.Stop, error)
	GetStopByName(name string) (tlgo.Stop, error)
	GetLines() ([]tlgo.Line, error)
	GetLineByName(name string) (tlgo.Line, error)
	GetRoutesForLineID(lineID string) ([]tlgo.Route, error)
	GetRoutesDetailsForRouteID(routeID string) (tlgo.RouteDetails, error)
	GetRoutesDetailsByRouteID() (map[string]tlgo.RouteDetails, error)
	GetLineForRouteID(routeID string) (tlgo.Line, error)
}
