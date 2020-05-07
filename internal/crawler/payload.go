package crawler

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/pipeline"
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

// Clone implements the pipeline.Payload
func (p *crawlerPayload) Clone() pipeline.Payload {
	newP := payloadPool.Get().(*crawlerPayload)
	newP.LinkID = p.LinkID
	newP.URL = p.URL
	newP.RetrievedAt = p.RetrievedAt
	newP.NoFollowLinks = append([]string(nil), p.NoFollowLinks...)
	newP.Links = append([]string(nil), p.Links...)
	newP.Title = p.Title
	newP.TextContext = p.TextContext
	_, err := io.Copy(&newP.RawContent, &p.RawContent)
	if err != nil {
		panic(fmt.Sprintf("[Bug] error cloing payload raw content: %v ", err))
	}
	return newP
}

// MarkAsProcessed implementes the pipeline.Payload
func (p *crawlerPayload) MarkAsProcessed() {
	p.URL = p.URL[:0]
	p.RawContent.Reset()
	p.NoFollowLinks = p.NoFollowLinks[:0]
	p.Links = p.Links[:0]
	p.Title = p.Title[:0]
	p.TextContext = p.TextContext[:0]
	payloadPool.Put(p)
}
