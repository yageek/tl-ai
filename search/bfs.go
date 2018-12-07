package search

import (
	"errors"
	"fmt"
	"log"

	"github.com/gophersch/tlgo"
	"github.com/yageek/tl-ai/storage"
)

type bsfLink struct {
	routeID string
	details tlgo.RouteDetails
	node    *bfsNode
	line    tlgo.Line
}

type bsfMove struct {
	fromNode *bfsNode
	viaLink  *bsfLink
}
type bfsNode struct {
	in      bsfMove
	links   []*bsfLink
	visited bool
	stop    tlgo.Stop
}

func (n *bfsNode) linkToNode(o *bfsNode, routeID string, line tlgo.Line, details tlgo.RouteDetails) {
	link := &bsfLink{
		routeID: routeID,
		details: details,
		node:    o,
		line:    line,
	}
	n.links = append(n.links, link)
}

func (n *bfsNode) mark() {
	n.visited = true
}

// BFS represents a bread first search pass
type BFS struct {
	graph           []*bfsNode
	nodesByStopName map[string]*bfsNode
}

var (
	// ErrNoPathFound is returned when no path between are found
	ErrNoPathFound = errors.New("No path found")
)

// NewBFS Create a new BFS session
func NewBFS(store storage.Store) (*BFS, error) {

	stops, err := store.GetStops()
	if err != nil {
		return nil, err
	}

	routesDetails, err := store.GetRoutesDetailsByRouteID()
	if err != nil {
		return nil, err
	}

	stopsNode := make([]*bfsNode, len(stops))
	nameIndex := make(map[string]*bfsNode, len(stops))

	for k := range stops {

		// Create the node of the stops
		node := &bfsNode{
			stop:    stops[k],
			visited: false,
		}
		stopsNode[k] = node
		nameIndex[stops[k].Name] = node
	}

	for routeID, details := range routesDetails {
		var previous *bfsNode

		for _, stopDetails := range details.Stops {
			current, hasFound := nameIndex[stopDetails.StopAreaName]
			if !hasFound {
				continue
			}

			line, err := store.GetLineForRouteID(routeID)
			if err != nil {
				fmt.Printf("Line not found for route ID: %s\n", err)
				continue
			}

			if previous != nil {
				previous.linkToNode(current, routeID, line, details)

				if details.Wayback {
					current.linkToNode(previous, routeID, line, details)
				}
			}
			previous = current
		}
	}

	return &BFS{
		graph:           stopsNode,
		nodesByStopName: nameIndex,
	}, nil
}

// Step represents a bus stop travel step
type Step struct {
	Stop         tlgo.Stop
	RouteDetails tlgo.RouteDetails
	RouteID      string
	Line         tlgo.Line
}

// FindStopToStopPath finds the path between two stops if it exists
func (s *BFS) FindStopToStopPath(source string, target string) ([]Step, error) {

	log.Printf("Starting search from %s -> %s\n", source, target)

	start, hasStart := s.nodesByStopName[source]
	if !hasStart {
		return []Step{}, fmt.Errorf("Starting stop %s was not found", source)
	}
	end, hastarget := s.nodesByStopName[target]
	if !hastarget {
		return []Step{}, fmt.Errorf("Target stop %s was not found", target)

	}
	return bfsSearchStopToStop(start, end)
}

func bfsSearchStopToStop(start *bfsNode, target *bfsNode) ([]Step, error) {
	queue := newQueue(1)

	queue.push(start)
	start.mark()
	for queue.count != 0 {
		n := queue.pop()

		if n == target {

			path := []Step{}
			nodeCursor := n

			for nodeCursor.in.fromNode != nil {

				step := Step{
					RouteID:      nodeCursor.in.viaLink.routeID,
					RouteDetails: nodeCursor.in.viaLink.details,
					Line:         nodeCursor.in.viaLink.line,
					Stop:         nodeCursor.stop,
				}

				path = append(path, step)
				fmt.Printf("Stop: %s | Step Route: %s | Step Line: %s\n", step.Stop.Name, step.RouteID, step.Line)
				nodeCursor = nodeCursor.in.fromNode
			}

			return path, nil
		}

		for _, c := range n.links {
			if !c.node.visited {
				c.node.in = bsfMove{n, c}
				queue.push(c.node)
				c.node.mark()
			}
		}
	}
	return nil, ErrNoPathFound
}
