package crawler

import (
	"context"
	"time"

	"github.com/joshvoll/linkrus/internal/pipeline"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
)

// textIndexer definition
type textIndexer struct {
	indexer Indexer
}

// newIndexer constructor function
func newTextIndexer(indexer Indexer) *textIndexer {
	return &textIndexer{
		indexer: indexer,
	}
}

// Process method implementation for the textIndexer type
func (i *textIndexer) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*crawlerPayload)
	doc := &index.Document{
		LinkID:    payload.LinkID,
		URL:       payload.URL,
		Title:     payload.Title,
		Content:   payload.TextContext,
		IndexedAt: time.Now(),
	}
	if err := i.indexer.Index(ctx, doc); err != nil {
		return nil, err
	}
	return p, nil
}
