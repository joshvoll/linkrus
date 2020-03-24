package graphtest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/graph"
	gc "gopkg.in/check.v1"
)

// SuiteBase define a re-usable set of graph-related test that can be execute
// againts any type that implements graph.Graph.
type SuiteBase struct {
	g graph.Graph
}

// SetGraph configure the test-suite to run all test again g.
func (s *SuiteBase) SetGraph(g graph.Graph) {
	s.g = g
}

// TestUpsertLink verify the upset link
func (s *SuiteBase) TestUpsertLink(c *gc.C) {
	ctx := context.Background()
	originalLink := &graph.Link{
		URL:         "https://www.sandals.com",
		RetrievedAt: time.Now().Add(-10 * time.Hour),
	}
	err := s.g.UpsertLink(ctx, originalLink)
	c.Assert(err, gc.IsNil)
	c.Assert(originalLink.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a linkID to be assign to new link"))

}

// TestFindLink going to find the link base on the id
func (s *SuiteBase) TestFindLink(c *gc.C) {
	ctx := context.Background()
	original := &graph.Link{
		URL:         "https://www.sandals.com",
		RetrievedAt: time.Now().Add(-10 * time.Hour),
	}
	err := s.g.UpsertLink(ctx, original)
	c.Assert(err, gc.IsNil)
	c.Assert(original.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a linkID to be assign to new link"))

	link, err := s.g.FindLink(ctx, original.ID)
	c.Assert(err, gc.IsNil)
	c.Assert(link, gc.DeepEquals, link, gc.Commentf("look by link.ID returning the wrong id"))
	fmt.Println("LINK: ", link)

}
