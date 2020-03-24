package cdb

import (
	"database/sql"

	"github.com/joshvoll/linkrus/internal/graph"
)

// linkIterator implements the graph.LinkIterator interface
type linkIterator struct {
	rows        *sql.Rows
	lastErr     error
	latchedLink *graph.Link
}

// Next implements Next() from graph.LinkIterator
func (i *linkIterator) Next() bool {
	if i.lastErr != nil || !i.rows.Next() {
		return false
	}
	link := new(graph.Link)
	if i.lastErr = i.rows.Scan(&link.ID, &link.URL, &link.RetrievedAt); i.lastErr != nil {
		return false
	}
	link.RetrievedAt = link.RetrievedAt.UTC()
	i.latchedLink = link
	return true
}

// Close implements Close from graph.LinkIterator
func (i *linkIterator) Close() error {
	if err := i.rows.Close(); err != nil {
		return err
	}
	return nil
}

// Error implements Error from graph.LinkIterator
func (i *linkIterator) Error() error {
	return i.lastErr
}

// Link() return the current fetch link, implementation from graph.LinkIterator
func (i *linkIterator) Link() *graph.Link {
	return i.latchedLink
}

// edgeIterator implementes the graph.EdgeIterator from interface
type edgeIterator struct {
	rows        *sql.Rows
	lastErr     error
	latchedEdge *graph.Edge
}

// Next implements Next() from graph.EdgeIterator
func (e *edgeIterator) Next() bool {
	if e.lastErr != nil || !e.rows.Next() {
		return false
	}
	edge := new(graph.Edge)
	if e.lastErr = e.rows.Scan(&edge.ID, &edge.Src, &edge.Dst, &edge.UpdateAt); e.lastErr != nil {
		return false
	}
	edge.UpdateAt = edge.UpdateAt.UTC()
	e.latchedEdge = edge
	return true

}

// Close implementes close from graph.EdgeIterator
func (e *edgeIterator) Close() error {
	if err := e.rows.Close(); err != nil {
		return err
	}
	return nil
}

// Error implements Error from graph.EdgeIterator
func (e *edgeIterator) Error() error {
	return e.lastErr
}

// Edge Implemente grahp.EdgeIterator
func (e *edgeIterator) Edge() *graph.Edge {
	return e.latchedEdge
}
