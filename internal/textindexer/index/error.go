package index

import "golang.org/x/xerrors"

var (
	// ErrNotFound is return by the indexer when there is no result
	ErrNotFound = xerrors.New("not found")
	// ErrMissingLinkID is return when atending to the indexer an there is no linkID
	ErrMissingLinkID = xerrors.New("document do not provide a valid linkID")
)
