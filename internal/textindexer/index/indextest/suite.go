package indextest

import "github.com/joshvoll/linkrus/internal/textindexer/index"

// SuiteBase define a re-usable set of index-related test that can
// be executed againts any type of test
type SuiteBase struct {
	idx index.Indexer
}

// SetIndexer  configurate to run all test-suite
func (s *SuiteBase) SetIndexer(idx index.Indexer) {
	s.idx = idx
}
