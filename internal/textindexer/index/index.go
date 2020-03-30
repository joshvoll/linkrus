package index

import "github.com/google/uuid"

// Indexer implement by objects that can index and search the documents.
// all data is provider by the linkrus crawler
type Indexer interface {
	// Index insert a new document to the index or update a existing one, it will return an error if there is a problem
	Index(doc *Document) error

	// FindByID look up a document base on its Linkd ID. and return a document
	FindByID(linkID uuid.UUID) (*Document, error)

	// Searh the index for a particular query and return back the result
	// Iterator
	Search(query Query) (Iterator, error)

	// UpdateScore updates the pagerank socre for a document with a specific link.
	// if no exists , a place holder document with the provided score will be created
	UpdateScore(linkID uuid.UUID, score float64) error
}

// Iterator is implemented by an object that can paginate the search
type Iterator interface {
	// Close the iterator and realease any allocated
	Close() error

	// Next load the next document matching the query.
	// it will rerturn false it there is no more document to pull.
	Next() bool

	// Error return the last error encountered on the Iterator
	Error() error

	// Document returns the current document from the results set.
	Document() *Document

	// TotalCount return the total rturn document.
	TotalCount() uint64
}

const (
	// QueryTypeMatch request the match of each expression term.
	QueryTypeMatch QueryType = iota

	// QueryTypeFrase searche for a expected search expression
	QueryTypeFrase
)

// QueryType describe the type of query that the indexer support
type QueryType uint8

// Query encapsulate the query and set of parameters to use when search indexed
type Query struct {
	// The way the indexer should interpreter base on input values, this is for flexibility
	Type QueryType

	// The Search expression, store the search query of the app
	// Search for query list keywords in order
	// search for the exact phrase match
	Expression string

	// the number of search
	Offset uint64
}
