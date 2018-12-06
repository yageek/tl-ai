package search

import (
	"errors"
	"fmt"
	"log"

	"github.com/gophersch/tlgo"
)

type bsfLink struct {
	line    *tlgo.Line
	route   *tlgo.Route
	details *tlgo.RouteDetails
	node    *bfsNode
}

type bsfMove struct {
	fromNode *bfsNode
	viaLink  *bsfLink
}
type bfsNode struct {
	in      bsfMove
	links   []*bsfLink
	visited bool
	stop    *tlgo.Stop
}

func (n *bfsNode) linkToNode(o *bfsNode, line *tlgo.Line, route *tlgo.Route, details *tlgo.RouteDetails) {
	link := &bsfLink{
		line:    line,
		route:   route,
		details: details,
		node:    o,
	}
	n.links = append(n.links, link)
}

func (n *bfsNode) mark() {
	n.visited = true
}

// BFS represents a bread first search pass
type BFS struct {
	graph         []*bfsNode
	stopNameIndex map[string]*bfsNode
}

var (
	// ErrNoPathFound is returned when no path between are found
	ErrNoPathFound = errors.New("No path found")
)

// NewBFS Create a new BFS session
func NewBFS(stops map[string]*tlgo.Stop, lineRouteIndex map[*tlgo.Route]*tlgo.Line, routesDetails map[*tlgo.Route]tlgo.RouteDetails) *BFS {

	stopsNode := make([]*bfsNode, len(stops))
	nameIndex := make(map[string]*bfsNode, len(stops))

	i := 0
	for k := range stops {

		// Create the node of the stops
		node := &bfsNode{
			stop:    stops[k],
			visited: false,
		}
		stopsNode[i] = node
		nameIndex[stops[k].Name] = node
		i++
	}

	for route, details := range routesDetails {
		var previous *bfsNode

		for _, stopDetails := range details.Stops {
			current, hasFound := nameIndex[stopDetails.StopAreaName]
			if !hasFound {
				continue
			}

			line, hasLine := lineRouteIndex[route]
			if !hasLine {
				log.Printf("WARN: no line linked to route: %s\n", route.Name)
				continue
			}
			if previous != nil {
				previous.linkToNode(current, line, route, &details)

				if details.Wayback {
					current.linkToNode(previous, line, route, &details)
				}
			}
			previous = current
		}
	}

	return &BFS{
		graph:         stopsNode,
		stopNameIndex: nameIndex,
	}
}

// Step represents a bus stop travel step
type Step struct {
	FromStop     *tlgo.Stop
	ToStop       *tlgo.Stop
	Line         *tlgo.Line
	RouteDetails *tlgo.RouteDetails
	Route        *tlgo.Route
}

// FindStopToStopPath finds the path between two stops if it exists
func (s *BFS) FindStopToStopPath(source string, target string) ([]Step, error) {

	log.Printf("Starting search from %s -> %s\n", source, target)

	start, hasSourceStop := s.stopNameIndex[source]
	if !hasSourceStop {
		return []Step{}, fmt.Errorf("Starting stop %s was not found", source)
	}
	end, hasTargetStop := s.stopNameIndex[target]
	if !hasTargetStop {
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

			path := make([]Step, 1)
			nodeCursor := n
			step := Step{
				ToStop: n.stop,
			}
			for nodeCursor != nil {

				step.FromStop = n.in.fromNode.stop
				step.Line = n.in.viaLink.line
				step.Route = n.in.viaLink.route
				step.RouteDetails = n.in.viaLink.details

				path = append(path, step)
				step = Step{
					ToStop:   nodeCursor.stop,
					FromStop: nil,
					Line:     nil,
					Route:    nil,
				}
				nodeCursor = n.in.fromNode
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
