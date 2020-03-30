package memory

import (
	"github.com/blevesearch/bleve"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
)

// bleveIterator implements index.Iterator
type bleveIterator struct {
	idx        *InMemoryBleveIndexer
	searchReq  *bleve.SearchRequest
	cumIdx     uint64
	rsIdx      int
	rs         *bleve.SearchResult
	latchedDoc *index.Document
	lastErr    error
}

// Close implements Close from index.Iterator
func (b *bleveIterator) Close() error {
	b.idx = nil
	b.searchReq = nil
	if b.rs != nil {
		b.cumIdx = b.rs.Total
	}
	return nil
}

// Error implements the Error from index.Iterator
func (b *bleveIterator) Error() error {
	return b.lastErr
}

// Next implements the Next() from index.Iterator
// load the next document matching the search if is not available return false
func (b *bleveIterator) Next() bool {
	if b.lastErr != nil || b.rs == nil || b.cumIdx >= b.rs.Total {
		return false
	}
	if b.rsIdx >= b.rs.Hits.Len() {
		b.searchReq.From += b.searchReq.Size
		if b.rs, b.lastErr = b.idx.idx.Search(b.searchReq); b.lastErr != nil {
			return false
		}
		b.rsIdx = 0
	}
	nextID := b.rs.Hits[b.rsIdx].ID
	if b.latchedDoc, b.lastErr = b.idx.findByID(nextID); b.lastErr != nil {
		return false
	}
	b.cumIdx++
	b.rsIdx++
	return true
}

// Document return the current document from the resutl set
func (b *bleveIterator) Document() *index.Document {
	return b.latchedDoc
}

// TotalCount return the total count of document
func (b *bleveIterator) TotalCount() uint64 {
	if b.rs == nil {
		return 0
	}
	return b.rs.Total
}
