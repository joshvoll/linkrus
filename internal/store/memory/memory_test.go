package memory

import (
	"testing"

	"github.com/joshvoll/linkrus/internal/graph/graphtest"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(new(InMemoryGraphTestSuite))

// InMemoryGraphTestSuite definition
type InMemoryGraphTestSuite struct {
	graphtest.SuiteBase
}

func Test(t *testing.T) {
	gc.TestingT(t)
}

func (s *InMemoryGraphTestSuite) SetUpTest(c *gc.C) {
	s.SetGraph(NewInMemoryGraph())
}
