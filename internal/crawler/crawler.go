package crawler

import (
	"context"
	"net/http"

	"github.com/joshvoll/linkrus/internal/linkgraph/graph"
	"github.com/joshvoll/linkrus/internal/pipeline"
)

// URLGetter is implemented by object that can performs HTTP GET request.
type URLGetter interface {
	Get(url string) (*http.Response, error)
}

// PrivateNetworkDetector is implemented by the obeject that can detect wheater
// a host resolves to a private network address
type PrivateNetworkDetector interface {
	IsPrivate(host string) (bool, error)
}

// LinkSource going to implement the graph.LinkIterator
type LinkSource struct {
	linkIt graph.LinkIterator
}

// Error implemented by the iterator
func (l *LinkSource) Error() error { return l.linkIt.Error() }

// Next implemented by the iterarot
func (l *LinkSource) Next(context.Context) bool { return l.linkIt.Next() }

// Payload implemente the iterator
func (l *LinkSource) Payload() pipeline.Payload {
	link := l.linkIt.Link()
	p := payloadPool.Get().(*crawlerPayload)
	p.LinkID = link.ID
	p.URL = link.URL
	p.RetrievedAt = link.RetrievedAt
	return p
}

// countingSink going to coint all sink implementeed on the pipeline
type countingSink struct {
	count int
}

// Consume implements the pipeline.Sink interfaceo
func (s *countingSink) Consume(_ context.Context, p pipeline.Payload) error {
	s.count++
	return nil
}

// getCount just return the number of sink
// the broadcast split-stage send out to payloads for each incoming link
// so we need to devide to total by 2.
func (s *countingSink) getCount() int {
	return s.count / 2
}
