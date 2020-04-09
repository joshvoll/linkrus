package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/google/uuid"
	"github.com/joshvoll/linkrus/internal/textindexer/index"
	"golang.org/x/xerrors"
)

// indexName is the name of the elasticsearch index to use
const indexName = "textindexer"

// batchSize is the size result for each query cached locally
const batchSize = 10

/*
var esMappings = `
{
  "mappings" : {
    "data":{
	"properties": {
	    "LinkID": {"type": "keyword"},
	    "URL": {"type": "keyword"},
            "Content": {"type": "text"},
            "Title": {"type": "text"},
	    "IndexedAt": {"type": "date"},
            "PageRank": {"type": "double"}
	}
    }
  }
}`
*/

var esMappings = `
{
  "mappings" : {
    "properties": {
      "LinkID": {"type": "keyword"},
      "URL": {"type": "keyword"},
      "Content": {"type": "text"},
      "Title": {"type": "text"},
      "IndexedAt": {"type": "date"},
      "PageRank": {"type": "double"}
    }
  }
}`

// esSearchRes search query document definition
type esSearchRes struct {
	Hits esSearchResHits `json:"hits"`
}

// esSearchResHits define total and hit list
type esSearchResHits struct {
	Total   esTotal        `json:"total"`
	HitList []esHitWrapper `json:"hits"`
}

// esTotal count the total of the hits
type esTotal struct {
	Count uint64 `json:"count"`
}

// HitList gets the total list
type esHitWrapper struct {
	DocSource esDoc `json:"_source"`
}

// esDoc are the documentation definition base on index.Document
type esDoc struct {
	LinkID    string    `json:"LinkID"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IndexedAt time.Time `json:"IndexedAt"`
	PageRank  float64   `json:"PageRank"`
}

// esUpdateRes define the update for the index method
type esUpdateRes struct {
	Result string `json:"result"`
}

// esErrorRes define the erros for the response unmarshal
type esErrorRes struct {
	Error esError `json:"error"`
}

// esError is the type of elastic serach error
type esError struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// Error() return the error from the esError method
func (e esError) Error() string {
	return fmt.Sprintf("%s : %s ", e.Type, e.Reason)
}

// ElasticSearchIndexer is an indexer implementation using elastic search.
// instance for the search query
type ElasticSearchIndexer struct {
	es         *elasticsearch.Client
	refreshOpt func(*esapi.UpdateRequest)
}

// NewElasticSearchIndexer create a new instance of the elastic search engine
func NewElasticSearchIndexer(esNode []string, syncUpdates bool) (*ElasticSearchIndexer, error) {
	cfg := elasticsearch.Config{
		Addresses: esNode,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if err = ensureIndex(es); err != nil {
		return nil, err
	}
	refreshOpt := es.Update.WithRefresh("false")
	if syncUpdates {
		refreshOpt = es.Update.WithRefresh("true")
	}
	return &ElasticSearchIndexer{
		es:         es,
		refreshOpt: refreshOpt,
	}, nil
}

// Index insert a new document to the indexer elastict search or update a existing one
func (i *ElasticSearchIndexer) Index(doc *index.Document) error {
	if doc.LinkID == uuid.Nil {
		return xerrors.Errorf("index: %w ", index.ErrMissingLinkID)
	}
	var (
		buf   bytes.Buffer
		esDoc = makeEsDoc(doc)
	)
	update := map[string]interface{}{
		"doc":           esDoc,
		"doc_as_upsert": true,
	}
	if err := json.NewEncoder(&buf).Encode(&update); err != nil {
		return xerrors.Errorf("index: %w ", err)
	}
	res, err := i.es.Update(indexName, esDoc.LinkID, &buf, i.refreshOpt)
	if err != nil {
		return xerrors.Errorf("index update: %w ", err)
	}
	var updateRes esUpdateRes
	if err = unmarshalResponse(res, &updateRes); err != nil {
		return xerrors.Errorf("index unmarshal: %w ", err)
	}
	return nil
}

// FindByID look up a document base on its Linkd ID. and return a document
func (i *ElasticSearchIndexer) FindByID(linkID uuid.UUID) (*index.Document, error) {
	if linkID == uuid.Nil {
		return nil, xerrors.Errorf("FindByID: %w ", index.ErrMissingLinkID)
	}
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"LinkID": linkID.String(),
			},
		},
		"from": 0,
		"size": 1,
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, xerrors.Errorf("find by id  query: %w ", err)
	}
	searchRes, err := runSearch(i.es, query)
	if err != nil {
		return nil, xerrors.Errorf("run search: %w ", err)
	}
	if len(searchRes.Hits.HitList) != 1 {
		return nil, xerrors.Errorf("search hits : %w ", err)
	}
	return mapEsDoc(&searchRes.Hits.HitList[0].DocSource), nil
}

// Search the index for a particular query and return back the results
// these result can be multiple queries
func (i *ElasticSearchIndexer) Search(q index.Query) (index.Iterator, error) {
	var querytype string
	switch q.Type {
	case index.QueryTypeFrase:
		querytype = "phrase"
	default:
		querytype = "best_fields"
	}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"type":   querytype,
						"query":  q.Expression,
						"fields": []string{"Title", "Content"},
					},
				},
			},
		},
		"from": q.Offset,
		"to":   batchSize,
	}
	searchRes, err := runSearch(i.es, query)
	if err != nil {
		return nil, xerrors.Errorf("search run search %w ", err)
	}
	return &esIterator{
		es:        i.es,
		searchReq: query,
		rs:        searchRes,
		cumIdx:    q.Offset,
	}, nil
}

// UpdateScore updates the PageRank score from a document with specific link
// if the linkid not exists with put a place holder with a new score
func (i *ElasticSearchIndexer) UpdateScore(linkID uuid.UUID, score float64) error {
	var buf bytes.Buffer
	update := map[string]interface{}{
		"doc": map[string]interface{}{
			"LinkID":   linkID.String(),
			"PageRank": score,
		},
		"doc_as_upsert": true,
	}
	if err := json.NewEncoder(&buf).Encode(update); err != nil {
		return xerrors.Errorf("UpdateScore encode update: %w ", err)
	}
	res, err := i.es.Update(indexName, linkID.String(), &buf, i.refreshOpt)
	if err != nil {
		return xerrors.Errorf("update score updating %w ", err)
	}
	var updateRes esUpdateRes
	if err = unmarshalResponse(res, &updateRes); err != nil {
		return xerrors.Errorf("update score unmarshal: %w ", err)
	}
	return nil
}

// mapEsDoc helper function return the index.Document ready for work with elastic search
func mapEsDoc(d *esDoc) *index.Document {
	return &index.Document{
		LinkID:    uuid.MustParse(d.LinkID),
		URL:       d.URL,
		Title:     d.Title,
		Content:   d.Content,
		IndexedAt: d.IndexedAt.UTC(),
		PageRank:  d.PageRank,
	}
}

// runSearch going to run the search on the elastic search db and return the struct with the findings
// decode the search for bytes
// perform the search query and check the error
func runSearch(es *elasticsearch.Client, searchQuery map[string]interface{}) (*esSearchRes, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, xerrors.Errorf("run search: %w ", err)
	}
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(indexName),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	var esRes esSearchRes
	if err = unmarshalResponse(res, &esRes); err != nil {
		return nil, err
	}
	return &esRes, nil
}

// makeEsDoc make the document ready for elastic search and prepper to the inserted
// Note: PageRank is intentionally skip it, still have no idea how to do this, and don't want to overwrite existing PageRank values.
// for update PageRank we will use another method with an id.
func makeEsDoc(doc *index.Document) esDoc {
	return esDoc{
		LinkID:    doc.LinkID.String(),
		URL:       doc.URL,
		Title:     doc.Title,
		Content:   doc.Content,
		IndexedAt: doc.IndexedAt.UTC(),
	}
}

// ensureIndex helper function to create the instance
func ensureIndex(es *elasticsearch.Client) error {
	mappingsReader := strings.NewReader(esMappings)
	res, err := es.Indices.Create(indexName, es.Indices.Create.WithBody(mappingsReader))
	if err != nil {
		return fmt.Errorf("error creatint the indexer instance : %v ", err)
	} else if res.IsError() {
		err := unmarshalError(res)
		if esErr, valid := err.(esError); valid && esErr.Type == "resource_already_exists_exception" {
			return nil
		}
		return xerrors.Errorf("could not create es instance: %w ", err)
	}
	return nil
}

// unmarshalError is just a helpeing for errors
func unmarshalError(res *esapi.Response) error {
	return unmarshalResponse(res, nil)
}

// unmarshalResponse is a valid response format and an unmarshal function tool
func unmarshalResponse(res *esapi.Response, to interface{}) error {
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		var errRes esErrorRes
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return err
		}
		return errRes.Error
	}
	return json.NewDecoder(res.Body).Decode(to)
}
