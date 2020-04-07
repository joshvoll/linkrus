package es

import (
	"os"
	"strings"
	"testing"

	"github.com/joshvoll/linkrus/internal/textindexer/index/indextest"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(new(ElasticSearchTestSuite))

// ElasticSearchTestSuite model definition
type ElasticSearchTestSuite struct {
	indextest.SuiteBase
	idx *ElasticSearchIndexer
}

// Test contructor searching
func Test(t *testing.T) {
	gc.TestingT(t)
}

// SetUpSuite is the test for setting up de db
func (s *ElasticSearchTestSuite) SetUpSuite(c *gc.C) {
	nodeList := os.Getenv("ES_NODES")
	if nodeList == "" {
		c.Fatalf("Missing elastic search node, skipping elastic search backed index test suite : %v", nodeList)
	}
	idx, err := NewElasticSearchIndexer(strings.Split(nodeList, ","), true)
	c.Assert(err, gc.IsNil)
	s.SetIndexer(idx)
	s.idx = idx
}

// SetUpTest just testing setup
func (s *ElasticSearchTestSuite) SetUpTest(c *gc.C) {
	if s.idx.es != nil {
		_, err := s.idx.es.Indices.Delete([]string{indexName})
		c.Assert(err, gc.IsNil)
		err = ensureIndex(s.idx.es)
		c.Assert(err, gc.IsNil)
	}

}
