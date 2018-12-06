package search

// queue is a basic FIFO queue based on a circular list that resizes as needed.
type queue struct {
	nodes []*bfsNode
	size  int
	head  int
	tail  int
	count int
}

// Newqueue returns a new queue with the given initial size.
func newQueue(size int) *queue {
	return &queue{
		nodes: make([]*bfsNode, size),
		size:  size,
	}
}

// Push adds a node to the queue.
func (q *queue) push(n *bfsNode) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*bfsNode, len(q.nodes)+q.size)
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
func (q *queue) pop() *bfsNode {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}
