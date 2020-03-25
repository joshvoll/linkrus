package cdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/linkgraph/graph"
	"github.com/lib/pq"
	"golang.org/x/xerrors"
)

var (
	upsertLinkQuery = `
	    INSERT INTO links (url, retrieved_at) VALUES ($1,$2)
	    ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2)
	    RETURNING id, retrieved_at`
	findLinkQuery        = "SELECT url, retrieved_at FROM links WHERE id = $1"
	linkInPartitionQuery = "SELECT id, url, retrieved_at FROM links WHERE id >= $1 AND id < $2 AND retrieved_at < $3"

	upsertEdgeQuery = `
	    INSERT INTO edges (src, dst, updated_at) VALUES ($1, $2, NOW())
	    ON CONFLICT (src, dst) DO UPDATE SET updated_at=NOW()
	    RETURNING id, updated_at`
	edgesInPartitionQuery = "SELECT id, src, dst, updated_at FROM edges WHERE src >= $1 AND src < $2 AND updated_at < $3"
	removeStaleEdgesQuery = "DELETE FROM edges WHERE src=$1 AND updated_at < $2"
)

// CockroachDBGraph struct definition implemente the graph persistence layer
type CockroachDBGraph struct {
	db *sql.DB
}

// NewCockroachDBGraph return the cockroach db instance client with the dns specific
func NewCockroachDBGraph(dns string) (*CockroachDBGraph, error) {
	db, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, err
	}
	return &CockroachDBGraph{
		db: db,
	}, nil
}

// Close just close the cockroach db instance
func (s *CockroachDBGraph) Close() error {
	return s.db.Close()
}

// UpsertLink create a new link or update an existing one base on the graph.Graph interface using cockroach db
func (s *CockroachDBGraph) UpsertLink(ctx context.Context, link *graph.Link) error {
	if err := s.db.QueryRowContext(ctx, upsertLinkQuery, link.URL, link.RetrievedAt).Scan(&link.ID, &link.RetrievedAt); err != nil {
		return xerrors.Errorf("upsert link: %w ", err)
	}
	link.RetrievedAt = link.RetrievedAt.UTC()
	return nil
}

// FindLink look up a link base on the ID. and return the struct of the Link.
func (s *CockroachDBGraph) FindLink(ctx context.Context, id uuid.UUID) (*graph.Link, error) {
	link := &graph.Link{
		ID: id,
	}
	if err := s.db.QueryRowContext(ctx, findLinkQuery, id).Scan(&link.URL, &link.RetrievedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, xerrors.Errorf("find link : %w ", graph.ErrNotFound)
		}
		return nil, xerrors.Errorf("find link: %w ", err)
	}
	link.RetrievedAt = link.RetrievedAt.UTC()
	return link, nil
}

// Links return all the link base on an Iterator, who belong to a particular link
// [fromID, toID] range can be retrieved
func (s *CockroachDBGraph) Links(ctx context.Context, fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	rows, err := s.db.QueryContext(ctx, linkInPartitionQuery, fromID, toID, retrievedBefore.UTC())
	if err != nil {
		return nil, xerrors.Errorf("Error Links: %w ", err)
	}
	return &linkIterator{
		rows: rows,
	}, nil
}

// UpsertEdge create a new edge or update an existing one.
func (s *CockroachDBGraph) UpsertEdge(ctx context.Context, edge *graph.Edge) error {
	if err := s.db.QueryRowContext(ctx, upsertEdgeQuery, edge.Src, edge.Dst).Scan(&edge.ID, &edge.UpdateAt); err != nil {
		if isForeignKeyViolation(err) {
			err = graph.ErrUnknownEdgeLinks
		}
		return xerrors.Errorf("Upsert Error: %w ", err)
	}
	edge.UpdateAt = edge.UpdateAt.UTC()
	return nil
}

// Edges return all the edges that are belong to a particular eedge.
// the source is on a vertex id [fromID, toID]
// range is update before provide timestamp
func (s *CockroachDBGraph) Edges(ctx context.Context, fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	rows, err := s.db.QueryContext(ctx, edgesInPartitionQuery, fromID, toID, updatedBefore.UTC())
	if err != nil {
		return nil, xerrors.Errorf("Error edges: %w ", err)
	}
	return &edgeIterator{
		rows: rows,
	}, nil
}

// RemoveStalEdges remove any edges from the origin specification
func (s *CockroachDBGraph) RemoveStalEdges(ctx context.Context, fromID uuid.UUID, updatedBefore time.Time) error {
	_, err := s.db.Exec(removeStaleEdgesQuery, fromID, updatedBefore.UTC())
	if err != nil {
		return xerrors.Errorf("remove stale edges: %w ", err)
	}
	return nil
}

// isForeignKeyViolatione returns true if there is a foreign key violation constrain violation
func isForeignKeyViolation(err error) bool {
	pqErr, valid := err.(*pq.Error)
	if !valid {
		return false
	}
	return pqErr.Code.Name() == "foreign_key_violation"
}
