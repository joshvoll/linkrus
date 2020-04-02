package es

import (
	"github.com/elastic/go-elasticsearch"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
)

// esIterator implements the index.Iterator interface
type esIterator struct {
	es         *elasticsearch.Client
	searchReq  map[string]interface{}
	rs         *esSearchRes
	rsIdx      int
	cumIdx     uint64
	latchedDoc *index.Document
	lastErr    error
}

// Close add the Close iterator from the index.Iterator
func (it *esIterator) Close() error {
	it.es = nil
	it.searchReq = nil
	it.cumIdx = it.rs.Hits.Total.Count
	return nil
}

// Next implements next from index.Iterator and return the next element on the query
// it will return false if there is no more document available
func (it *esIterator) Next() bool {
	if it.lastErr != nil || it.rs == nil || it.cumIdx >= it.rs.Hits.Total.Count {
		return false
	}
	if it.rsIdx >= len(it.rs.Hits.HitList) {
		it.searchReq["from"] = it.searchReq["from"].(uint64) + batchSize
		if it.rs, it.lastErr = runSearch(it.es, it.searchReq); it.lastErr != nil {
			return false
		}
		it.rsIdx = 0
	}
	it.latchedDoc = mapEsDoc(&it.rs.Hits.HitList[it.rsIdx].DocSource)
	it.cumIdx++
	it.rsIdx++
	return true
}

// Error implements Error from index.Iterator
func (it *esIterator) Error() error {
	return it.lastErr
}

// Document return the current document from the result
func (it *esIterator) Document() *index.Document {
	return it.latchedDoc
}

// TotalCount return the aproximade number of document on the query
func (it *esIterator) TotalCount() uint64 {
	return it.rs.Hits.Total.Count
}
