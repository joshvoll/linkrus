package memory

import "github.com/joshvoll/linkrus/internal/graph"

// linkIterator  struct definition
type linkIterator struct {
	s *InMemoryGraph

	links        []*graph.Link
	currentIndex int
}

// Next() implemeantion fro the graph.Iterator interface
func (l *linkIterator) Next() bool {
	if l.currentIndex >= len(l.links) {
		return false
	}
	l.currentIndex++
	return true
}

// Close implemente Close method from graph.Iterator interface
func (l *linkIterator) Close() error {
	return nil
}

// Error implemente Error() method from graph.Iterator interface
func (l *linkIterator) Error() error {
	return nil
}

// Link implementes the link.Iterator from graph
func (l *linkIterator) Link() *graph.Link {
	l.s.mu.RLock()
	link := new(graph.Link)
	*link = *l.links[l.currentIndex-1]
	l.s.mu.RUnlock()
	return link
}

// edgeIterator struct definition
type edgeIterator struct {
	s *InMemoryGraph

	edges        []*graph.Edge
	currentIndex int
}

// Next() implemente the Next() method from the graph.Iterator interfaceP
func (e *edgeIterator) Next() bool {
	if e.currentIndex >= len(e.edges) {
		return false
	}
	e.currentIndex++
	return true
}

// Error implements Error() from graph.Iterator
func (e *edgeIterator) Error() error {
	return nil
}

// Close implements Close() from graph.Iterator
func (e *edgeIterator) Close() error {
	return nil
}

// Edge return the current edge fetch by the iterator
func (e *edgeIterator) Edge() *graph.Edge {
	e.s.mu.RLock()
	edge := new(graph.Edge)
	*edge = *e.edges[e.currentIndex-1]
	e.s.mu.RUnlock()
	return edge
}
