package graphtest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/linkgraph/graph"
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
		URL:         "https://www.beaches.com",
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

// TestUpsertEdge just going to update edges to the databae
func (s *SuiteBase) TestUpsertEdge(c *gc.C) {
	ctx := context.Background()
	// createing the link
	linkUUIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		link := &graph.Link{
			URL: fmt.Sprint(i),
		}
		c.Assert(s.g.UpsertLink(ctx, link), gc.IsNil)
		linkUUIDs[i] = link.ID

	}
	// creating the edge
	edge := &graph.Edge{
		Src: linkUUIDs[0],
		Dst: linkUUIDs[1],
	}
	err := s.g.UpsertEdge(ctx, edge)
	c.Assert(err, gc.IsNil)
	c.Assert(edge.ID, gc.Not(gc.Equals), uuid.Nil, gc.Commentf("expected a edgeID to be assing to the new edge"))
	c.Assert(edge.UpdateAt.IsZero(), gc.Equals, false, gc.Commentf("UpdateAt field not setup"))
}
