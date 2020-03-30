package memory

import (
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
	"golang.org/x/xerrors"
)

const batchSize = 10

// bleveDoc struct definition
type bleveDoc struct {
	Title    string
	Content  string
	PageRank float64
}

// InMemoryBleveIndexer is the indexer defintion from the index
type InMemoryBleveIndexer struct {
	mu   sync.RWMutex
	docs map[string]*index.Document

	idx bleve.Index
}

// NewInMemoryBleveIndexer return the memory test
func NewInMemoryBleveIndexer() (*InMemoryBleveIndexer, error) {
	mapping := bleve.NewIndexMapping()
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}
	return &InMemoryBleveIndexer{
		idx:  idx,
		docs: make(map[string]*index.Document),
	}, nil
}

// Close the indexer and release all connection
func (i *InMemoryBleveIndexer) Close() error {
	return i.idx.Close()
}

// Index insert a new document index or update an existing one
func (i *InMemoryBleveIndexer) Index(doc *index.Document) error {
	if doc.LinkID == uuid.Nil {
		return xerrors.Errorf("index %w ", index.ErrMissingLinkID)
	}
	doc.IndexedAt = time.Now()
	dcopy := copyDoc(doc)
	key := dcopy.LinkID.String()
	i.mu.Lock()
	if orig, exists := i.docs[key]; exists {
		dcopy.PageRank = orig.PageRank
	}
	if err := i.idx.Index(key, makeBleveDoc(dcopy)); err != nil {
		return xerrors.Errorf("index: %w ", err)
	}
	i.docs[key] = dcopy
	i.mu.Unlock()
	return nil
}

// FindByID return the document base on the id from the link R us link project
func (i *InMemoryBleveIndexer) FindByID(linkID uuid.UUID) (*index.Document, error) {
	return i.findByID(linkID.String())
}

// findByID look a document by its linkID expressed as string.
func (i *InMemoryBleveIndexer) findByID(linkID string) (*index.Document, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if d, found := i.docs[linkID]; found {
		return copyDoc(d), nil
	}
	return nil, xerrors.Errorf("find by id: %w ", index.ErrNotFound)
}

// Search for a particular document return back an Iterator
func (i *InMemoryBleveIndexer) Search(q index.Query) (index.Iterator, error) {
	var bq query.Query
	switch q.Type {
	case index.QueryTypeFrase:
		bq = bleve.NewMatchPhraseQuery(q.Expression)
	default:
		bq = bleve.NewMatchQuery(q.Expression)
	}
	searchReq := bleve.NewSearchRequest(bq)
	searchReq.SortBy([]string{"-PageRank", "-_score"})
	searchReq.Size = batchSize
	searchReq.From = int(q.Offset)
	rs, err := i.idx.Search(searchReq)
	if err != nil {
		return nil, xerrors.Errorf("serach %w : ", err)
	}
	return &bleveIterator{
		idx:       i,
		searchReq: searchReq,
		rs:        rs,
		cumIdx:    q.Offset,
	}, nil
}

// UpdateScore it will udpate the score or existing one base on link id and score pass
func (i *InMemoryBleveIndexer) UpdateScore(linkID uuid.UUID, score float64) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	key := linkID.String()
	doc, found := i.docs[key]
	if !found {
		doc = &index.Document{
			LinkID: linkID,
		}
		i.docs[key] = doc
	}
	doc.PageRank = score
	if err := i.idx.Index(key, makeBleveDoc(doc)); err != nil {
		return xerrors.Errorf("update score: %w ", err)
	}
	return nil
}

func copyDoc(d *index.Document) *index.Document {
	dcopy := new(index.Document)
	*dcopy = *d
	return dcopy
}

// makeBleveDoc just copy the doc to the bleve memory
func makeBleveDoc(d *index.Document) bleveDoc {
	return bleveDoc{
		Title:    d.Title,
		Content:  d.Content,
		PageRank: d.PageRank,
	}
}
