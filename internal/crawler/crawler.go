package crawler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/linkgraph/graph"
	"github.com/joshvoll/linkrus/internal/pipeline"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
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

// Graph is implemented by the objects that can upsert the links and edges into a links
// graph intnace.
type Graph interface {
	// UpsertLink create a new link or updates an existing link.
	UpsertLink(ctx context.Context, link *graph.Link) error

	// UpsertEdge create a new edge or updates an existing edge.
	UpdsertEdge(ctx context.Context, edge *graph.Edge) error

	// RemoveStaleEdges remove any edge that originates from the specified
	// link ID and was updated before the espcifiet timestamp.
	RemoveStaleEdges(ctx context.Context, fromID uuid.UUID, updateBefore time.Time) error
}

// Indexer is the implemented by objects that can index the contents of web-pages
// retrieved by the crawler pipeline.
type Indexer interface {
	// INdex inserts a new document to the index or updates the index entrye
	// for and existing document.
	Index(ctx context.Context, doc *index.Document) error
}

// Config encapsulates the configuration options for creating new Crawler.
type Config struct {
	// PrivateNetworkDetector a instance
	PrivateNetworkDetector PrivateNetworkDetector

	// A URLGetter instance for fetching links.
	URLGetter URLGetter

	// A GraphUpdater instance for adding new links to the link graph.
	Graph Graph

	// A TextIndexer instance for indexing the content of each retrieved links.
	Indexer Indexer

	// The numbers of concurrent worker used for retrieving links.
	FetchWorkers int
}

// Crawler implements a web-page crawling pipeline consisting of the following
// stages:
//
// - Given a URL, retrieve the web-page contents from the remote server.
// - Extract and resolve absolute and relative links from the retrieved page.
// - Extract page title and text content from the retrieved page.
// - Update the link graph: add new links and create edges between the crawled
//   page and the links within it.
// - Index crawled page title and text content.
type Crawler struct {
	p *pipeline.Pipeline
}

// NewCrawler returns a new crawler instnace.
func NewCrawler(cfg Config) *Crawler {
	return &Crawler{
		p: assembleCrawlerPipeline(cfg),
	}
}

// assembleCrawlerPipeline creates the various stages of a crawler pipeline
// using the options in cfg and assembles them into a pipeline instance.
func assembleCrawlerPipeline(cfg Config) *pipeline.Pipeline {
	return pipeline.New(
		pipeline.FixedWorkerPool(
			newLinkFetcher(cfg.URLGetter, cfg.PrivateNetworkDetector),
			cfg.FetchWorkers,
		),
		pipeline.FIFO(newLinkExtractor(cfg.PrivateNetworkDetector)),
		pipeline.FIFO(newTextExtractor()),
		pipeline.Broadcast(
			newGraphUdater(cfg.Graph),
			newTextIndexer(cfg.Indexer),
		),
	)
}

// Crawl iterates linkIt and send each link through the crawler pipeline
// returning the total count of links that went through the pipeline. Call to
// Crawl block util the link iterator is exhausted, an error occurs or the
// context is cancelled
func (c *Crawler) Crawl(ctx context.Context, linkIt graph.LinkIterator) (int, error) {
	sink := new(countingSink)
	err := c.p.Process(ctx, &linkSource{linkIt: linkIt}, sink)
	return sink.getCount(), err
}

// LinkSource going to implement the graph.LinkIterator
type linkSource struct {
	linkIt graph.LinkIterator
}

// Error implemented by the iterator
func (l *linkSource) Error() error { return l.linkIt.Error() }

// Next implemented by the iterarot
func (l *linkSource) Next(context.Context) bool { return l.linkIt.Next() }

// Payload implemente the iterator
func (l *linkSource) Payload() pipeline.Payload {
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
