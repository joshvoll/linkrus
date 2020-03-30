package index

import (
	"time"

	"github.com/google/uuid"
)

// Document describe a web-page who content have been indexe by linkgraph.
type Document struct {
	// The ID of the link graph entrey that point to this document.
	LinkID uuid.UUID
	// the url where the document was obteined.
	URL string
	// the document title (not mandatory)
	Title string
	// the document body
	Content string
	// the last time this document was index.
	IndexedAt time.Time
	// the PageRank score (need some help for this one).
	PageRank float64
}
