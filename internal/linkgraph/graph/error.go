package graph

import "golang.org/x/xerrors"

var (
	// ErrUnknownEdgeLinks is return when attend to create an edge with a invalid source and/or destination
	ErrUnknownEdgeLinks = xerrors.New("Unknow source and/or destination for edge")
	// ErrNotFound return an error if the link is not found
	ErrNotFound = xerrors.New("Unkown link for specific ID")
)
