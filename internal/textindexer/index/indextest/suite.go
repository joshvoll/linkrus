package indextest

import (
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
	gc "gopkg.in/check.v1"
)

// SuiteBase define a re-usable set of index-related test that can
// be executed againts any type of test
type SuiteBase struct {
	idx index.Indexer
}

// SetIndexer  configurate to run all test-suite
func (s *SuiteBase) SetIndexer(idx index.Indexer) {
	s.idx = idx
}

// TestIndexDocument verifies the indexings logic for existing documents.
// Insert new document
// Update new document
// Insert document without an id
func (s *SuiteBase) TestIndexDocument(c *gc.C) {
	doc := &index.Document{
		LinkID:    uuid.New(),
		URL:       "https://www.sandals.com/",
		Title:     "luxury included island",
		Content:   "lorem ipsum",
		IndexedAt: time.Now().Add(-12 * time.Hour).UTC(),
	}
	err := s.idx.Index(doc)
	c.Assert(err, gc.IsNil)
	updateDoc := &index.Document{
		LinkID:    doc.LinkID,
		URL:       "https://www.sandals.com/",
		Title:     "another one",
		Content:   "nothing about beaches for now",
		IndexedAt: time.Now().UTC(),
	}
	err = s.idx.Index(updateDoc)
	c.Assert(err, gc.IsNil)
	imconpliteDoc := &index.Document{
		URL: "http://wwww.sanservices.hn",
	}
	err = s.idx.Index(imconpliteDoc)
	c.Assert(err, gc.IsNil)

}
