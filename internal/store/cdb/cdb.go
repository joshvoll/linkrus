package cdb

import (
	"context"
	"database/sql"

	"github.com/joshvoll/linkrus/internal/graph"
	"golang.org/x/xerrors"
)

var (
	upsertLinkQuery = `
	    INSERT INTO links (url, retrieved_at) VALUES ($1,$2)
	    ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2)
	    RETURNING id, retrieved_at`
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
