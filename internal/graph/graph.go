package graph

import (
	"time"

	"github.com/google/uuid"
)

// Link encapsulate all the information data about a link
type Link struct {
	// a unique uuid for each link
	ID uuid.UUID
	// the url target for each linl
	URL string
	// timestamp when the link was retrieve it
	RetrievedAt time.Time
}

// LinkIterator describe the behavior for interactions from the linkd tabls
type LinkIterator interface {
	Iterator

	// Link return the current fetch link from an object
	Link() *Link
}

// Iterator is the main behavior of the interactions
type Iterator interface {
	// Next advance the search on the query, if there is no more item available return false
	Next() bool
	// Error return the last error encounter for the iterator
	Error() error
	// Close release any association with the current iterator
	Close() error
}

// EdgeIterator extend the interator interface to access next(), Error(), Close() to save the graph
type EdgeIterator interface {
	Iterator

	// Edge return the current edge fetch by the object
	Edge() error
}

// Edge describe the graph each that origin from src.
type Edge struct {
	// a unique uuid
	ID uuid.UUID
	// the origin of the link
	Src uuid.UUID
	// the destination of the link
	Dst uuid.UUID
	// the tiemestamp for the events
	UpdateAt time.Time
}

// Graph contain the core logic of the application
type Graph interface {
	// UpsertLink create a new link or update an existing one.
	UpsertLink(link *Link) error
	// FindLink look up and link base on the ide
	FindLink(id uuid.UUID) (*Link, error)
	// Links return all the link base on a iterator who ide belong to that particular link
	// [fromID, toID] range can be retrieved
	Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (LinkIterator, error)
	// UpsertEdge create a new edge or update a exsiting ne
	UpsertEdge(edge *Edge) error
	// Edges return all the edges that are belong for the particular edage
	// the source is on vertex id [fromID, toID]
	// rnage is update before provide a timestamp
	Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (EdgeIterator, error)
	// RemoveStaledges remove any edges from the origin specifications
	RemoveStaledges(fromID, toID uuid.UUID, udpatedBefore time.Time) error
}
