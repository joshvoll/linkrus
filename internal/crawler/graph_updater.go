package crawler

import (
	"context"
	"time"

	"github.com/joshvoll/linkrus/internal/linkgraph/graph"
	"github.com/joshvoll/linkrus/internal/pipeline"
)

// graphUpdater definition
type graphUpdater struct {
	updater Graph
}

func newGraphUdater(updater Graph) *graphUpdater {
	return &graphUpdater{
		updater: updater,
	}
}

// Process of the graphUpdater using the payload interface
// Upsert discovered links and create edges for them. Keep track of
// the current time so we can drop stale edges that have not been
// updated after this loop.
func (u *graphUpdater) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*crawlerPayload)
	src := &graph.Link{
		ID:          payload.LinkID,
		URL:         payload.URL,
		RetrievedAt: time.Now(),
	}
	if err := u.updater.UpsertLink(ctx, src); err != nil {
		return nil, err
	}
	for _, dstLink := range payload.NoFollowLinks {
		dst := &graph.Link{
			URL: dstLink,
		}
		if err := u.updater.UpsertLink(ctx, dst); err != nil {
			return nil, err
		}
	}
	removeEdgeOlderThan := time.Now()
	for _, dstLink := range payload.Links {
		dst := &graph.Link{
			URL: dstLink,
		}
		if err := u.updater.UpsertLink(ctx, dst); err != nil {
			return nil, err
		}
		if err := u.updater.UpsertEdge(ctx, &graph.Edge{Src: src.ID, Dst: dst.ID}); err != nil {
			return nil, err
		}
	}
	if err := u.updater.RemoveStalEdges(ctx, src.ID, removeEdgeOlderThan); err != nil {
		return nil, err
	}
	return payload, nil
}
