package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/graph"
	"golang.org/x/xerrors"
)

// edgeList is the type of each link and edge
type edgeList []uuid.UUID

// InMemoryGraph implements a in memory link graph base on the requirements
type InMemoryGraph struct {
	mu sync.RWMutex

	links map[uuid.UUID]*graph.Link
	edges map[uuid.UUID]*graph.Edge

	linkURLIndex map[string]*graph.Link
	linkEdgeMap  map[uuid.UUID]edgeList
}

// NewInMemoryGraph  create a new in-memory link graph
func NewInMemoryGraph() *InMemoryGraph {
	return &InMemoryGraph{
		links:        make(map[uuid.UUID]*graph.Link),
		edges:        make(map[uuid.UUID]*graph.Edge),
		linkURLIndex: make(map[string]*graph.Link),
		linkEdgeMap:  make(map[uuid.UUID]edgeList),
	}
}

// UpsertLink creates a new link or updates a existing one.
// check if the link with the same url already exits. if so convert this as update and point that to an existing link
// Assign new ID and insert the link
func (s *InMemoryGraph) UpsertLink(ctx context.Context, link *graph.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.linkURLIndex[link.URL]; existing != nil {
		link.ID = existing.ID
		origTs := existing.RetrievedAt
		*existing = *link
		if origTs.After(existing.RetrievedAt) {
			existing.RetrievedAt = origTs
		}
	}
	for {
		link.ID = uuid.New()
		if s.links[link.ID] == nil {
			break
		}
	}
	lCopy := new(graph.Link)
	*lCopy = *link
	s.linkURLIndex[lCopy.URL] = lCopy
	s.links[lCopy.ID] = lCopy
	return nil
}

// FindLink find the links for specific id.
func (s *InMemoryGraph) FindLink(ctx context.Context, id uuid.UUID) (*graph.Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	link := s.links[id]
	if link == nil {
		return nil, xerrors.Errorf("find link: %w ", graph.ErrNotFound)
	}
	lCopy := new(graph.Link)
	*lCopy = *link
	return lCopy, nil

}

// Links return an iterator for the set of links whose IDs belong to.
// [fromID, toID] range is retrieved
func (s *InMemoryGraph) Links(ctx context.Context, fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	from, to := fromID.String(), toID.String()
	s.mu.RLock()
	var list []*graph.Link
	for linkID, link := range s.links {
		if id := linkID.String(); id >= from && id < to && link.RetrievedAt.Before(retrievedBefore) {
			list = append(list, link)
		}
	}
	s.mu.RUnlock()
	return &linkIterator{
		s:     s,
		links: list,
	}, nil

}

// Edges return an iteractor from the edge vertex id
func (s *InMemoryGraph) Edges(ctx context.Context, fromID, toID uuid.UUID, updateBefore time.Time) (graph.EdgeIterator, error) {
	from, to := fromID.String(), toID.String()
	s.mu.RLock()
	var list []*graph.Edge
	for linkID := range s.links {
		if id := linkID.String(); id < from || id >= to {
			continue
		}
		for _, edgeID := range s.linkEdgeMap[linkID] {
			if edge := s.edges[edgeID]; edge.UpdateAt.Before(updateBefore) {
				list = append(list, edge)
			}
		}
	}

	s.mu.RUnlock()
	return &edgeIterator{
		s:     s,
		edges: list,
	}, nil

}

// UpsertEdge create a new edge or update for the existing one
// insert new edge to the memory
func (s *InMemoryGraph) UpsertEdge(ctx context.Context, edge *graph.Edge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, srcExists := s.links[edge.Src]
	_, dstExists := s.links[edge.Dst]
	if !srcExists || !dstExists {
		return xerrors.Errorf("upsert edge: %w ", graph.ErrUnknownEdgeLinks)
	}
	for _, edgeID := range s.linkEdgeMap[edge.Src] {
		existingEdge := s.edges[edgeID]
		if existingEdge.Src == edge.Src && existingEdge.Dst == edge.Dst {
			existingEdge.UpdateAt = time.Now()
			*edge = *existingEdge
			return nil
		}

	}
	for {
		edge.ID = uuid.New()
		if s.edges[edge.ID] == nil {
			break
		}
	}
	edge.UpdateAt = time.Now()
	eCopy := new(graph.Edge)
	*eCopy = *edge
	s.edges[eCopy.ID] = eCopy
	s.linkEdgeMap[edge.Src] = append(s.linkEdgeMap[edge.Src], eCopy.ID)
	return nil
}

// RemoveStalEdges remove any edges specific from the link ID
// update before specific update
func (s *InMemoryGraph) RemoveStalEdges(ctx context.Context, fromID, toID uuid.UUID, updatedBefore time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var newEdgelist edgeList
	for _, edgeID := range s.linkEdgeMap[fromID] {
		edge := s.edges[edgeID]
		if edge.UpdateAt.Before(updatedBefore) {
			delete(s.edges, edgeID)
			continue
		}
		newEdgelist = append(newEdgelist, edgeID)
	}
	s.linkEdgeMap[fromID] = newEdgelist
	return nil
}
