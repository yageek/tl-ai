package search

import (
	"errors"
	"fmt"
	"log"

	"github.com/gophersch/tlgo"
)

type NodeKind int

// SearchBFS represents a bread first search pass
type SearchBFS struct {
	graph         []*Node
	stopNameIndex map[string]*Node
}

// Create a new BFS search
func NewBFS(stops []tlgo.Stop, routes map[string]tlgo.RouteDetails) *SearchBFS {

	stopsNode := make([]*Node, len(stops))
	nameIndex := make(map[string]*Node, len(stops))
	for i := range stops {

		// Create the node of the stops
		node := &Node{
			nodeKind: nodeKindStop,
			ID:       stops[i].ID,
			Name:     stops[i].Name,
			visited:  false,
			childs:   []*Node{},
		}
		stopsNode[i] = node
		nameIndex[stops[i].Name] = node
	}

	for _, details := range routes {
		var previous *Node

		for _, stopDetails := range details.Stops {
			current, hasFound := nameIndex[stopDetails.StopAreaName]
			if !hasFound {
				// log.Printf("Can not find stop %s in index\n", stopDetails.ID)
				continue
			}
			// log.Printf("Found stop %s in index\n", current.Name)
			if previous != nil {
				previous.childs = append(previous.childs, current)

				if details.Wayback {
					current.childs = append(current.childs, previous)
				}
			}
			previous = current
		}
	}

	return &SearchBFS{
		graph:         stopsNode,
		stopNameIndex: nameIndex,
	}
}

func (s *SearchBFS) FindStopToStopPath(source string, target string) ([]*Node, error) {

	log.Printf("Starting search from %s -> %s\n", source, target)

	start, hasSourceStop := s.stopNameIndex[source]
	if !hasSourceStop {
		return nil, fmt.Errorf("Starting stop %s was not found", source)
	}
	end, hasTargetStop := s.stopNameIndex[target]
	if !hasTargetStop {
		return nil, fmt.Errorf("Target stop %s was not found", target)

	}

	fmt.Printf("Debug start: %+v -> %+v\n", start, end)
	return bfsSearchStopToStop(start, end)

}

const (
	nodeKindStop NodeKind = iota
	nodeKindLine
)

type Node struct {
	nodeKind NodeKind
	ID       string
	Name     string
	childs   []*Node
	visited  bool
	parent   *Node
}

func (n *Node) mark() {
	n.visited = true
}

// queue is a basic FIFO queue based on a circular list that resizes as needed.
type queue struct {
	nodes []*Node
	size  int
	head  int
	tail  int
	count int
}

// Newqueue returns a new queue with the given initial size.
func newQueue(size int) *queue {
	return &queue{
		nodes: make([]*Node, size),
		size:  size,
	}
}

// Push adds a node to the queue.
func (q *queue) push(n *Node) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*Node, len(q.nodes)+q.size)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *queue) pop() *Node {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}

func bfsSearchStopToStop(start *Node, target *Node) ([]*Node, error) {
	queue := newQueue(1)

	queue.push(start)
	start.mark()
	for queue.count != 0 {
		n := queue.pop()

		if n == target {

			path := []*Node{}
			cursor := n
			for cursor != nil {
				path = append(path, cursor)
				cursor = cursor.parent
			}
			return path, nil
		}

		for _, c := range n.childs {
			if !c.visited {
				c.parent = n
				queue.push(c)
				c.mark()
			}
		}
	}
	return nil, errors.New("No routes find :(")
}
