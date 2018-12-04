package search

import (
	"errors"
	"fmt"
	"log"

	"github.com/gophersch/tlgo"
)

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
func NewBFS(stops []tlgo.Stop, lines []tlgo.Line, routes map[string]tlgo.RouteDetails) *BFS {

	stopsNode := make([]*bfsNode, len(stops))
	nameIndex := make(map[string]*bfsNode, len(stops))
	lineMap := make(map[string]*tlgo.Line)

	for _, line := range lines {
		lineMap[line.ID] = &line
	}

	for i := range stops {

		// Create the node of the stops
		node := &bfsNode{
			stop:    &stops[i],
			visited: false,
			links:   []*bfsNodeLink{},
		}
		stopsNode[i] = node
		nameIndex[stops[i].Name] = node
	}

	for lineID, details := range routes {
		var previous *bfsNode

		for _, stopDetails := range details.Stops {
			current, hasFound := nameIndex[stopDetails.StopAreaName]
			if !hasFound {
				// log.Printf("Can not find stop %s in index\n", stopDetails.ID)
				continue
			}
			// log.Printf("Found stop %s in index\n", current.Name)
			if previous != nil {
				previous.addLinkToNode(current, lineMap[lineID], details.Wayback)

				if details.Wayback {
					current.addLinkToNode(previous, lineMap[lineID], details.Wayback)
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

type Step struct {
	FromStop *tlgo.Stop
	Stop     *tlgo.Stop
	ByLine   *tlgo.Line
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

	path, err := bfsSearchStopToStop(start, end)
	if err != nil {
		return []Step{}, err
	}

	outs := make([]Step, len(path)+1)

	outs[0] = Step{
		FromStop: nil,
		Stop:     start.stop,
		ByLine:   nil,
	}

	cursor := Step{
		FromStop: start.stop,
		Stop:     nil,
		ByLine:   nil,
	}

	for i := range path {
		step := path[len(path)-i-1]
		current := step.followedLink.node.stop
		cursor.Stop = current
		cursor.ByLine = step.followedLink.line

		outs[i+1] = cursor

		cursor = Step{
			FromStop: current,
			Stop:     nil,
			ByLine:   nil,
		}
	}

	return outs, nil
}

func bfsSearchStopToStop(start *bfsNode, target *bfsNode) ([]stepNode, error) {
	queue := newQueue(1)

	queue.push(start)
	start.mark()
	for queue.count != 0 {
		n := queue.pop()

		if n == target {

			path := []stepNode{}
			cursor := n
			for cursor.stepNode.followedLink != nil {
				path = append(path, cursor.stepNode)
				cursor = cursor.stepNode.parentNode
			}
			return path, nil
		}

		for _, c := range n.links {
			if !c.node.visited {
				c.node.stepNode = stepNode{n, c}
				queue.push(c.node)
				c.node.mark()
			}
		}
	}
	return nil, errors.New("No routes find :(")
}
