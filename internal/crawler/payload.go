package crawler

import (
	"bytes"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	payloadPool = sync.Pool{
		New: func() interface{} {
			return new(crawlerPayload)
		},
	}
)

// crawlerPayload definition for the model of the crawler
type crawlerPayload struct {
	LinkID      uuid.UUID
	URL         string
	RetrievedAt time.Time
	RawContent  bytes.Buffer

	// NoFollowLinks are still added to the graph but no outgoint edges
	// will be created from this link to them
	NoFollowLinks []string
	Links         []string
	Title         string
	TextContext   string
}
