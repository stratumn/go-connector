package search_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blevesearch/bleve"
	bs "github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	chainscript "github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-connector/services/blevestore/mockblevestore"
	"github.com/stratumn/go-connector/services/search"
)

func TestSearchService(t *testing.T) {
	// mock service dependencies
	ctrl := gomock.NewController(t)
	mockStore := mockblevestore.NewMockIndex(ctrl)

	s := &search.Service{}
	s.SetConfig(search.Config{
		Store: "blevestore",
	})
	s.Plug(map[string]interface{}{
		"blevestore": mockStore,
	})

	ctx := context.Background()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	searcher := s.Expose().(search.Searcher)

	// Create link data and encrypt it.
	d1 := map[string]interface{}{"life": "42"}
	l1, _ := chainscript.NewLinkBuilder("p", "m").WithData(d1).Build()
	lb1, _ := json.Marshal(l1)
	d2 := map[string]interface{}{"pizza": "yolo"}
	l2, _ := chainscript.NewLinkBuilder("p", "m").WithData(d2).Build()
	lb2, _ := json.Marshal(l2)

	str := "https://bit.ly/1a4vHym"
	mockStore.EXPECT().Search(gomock.Any()).DoAndReturn(func(r *bleve.SearchRequest) (*bleve.SearchResult, error) {
		assert.Equal(t, []string{"*"}, r.Fields)

		q, ok := r.Query.(*query.MatchQuery)
		require.True(t, ok, "the query should be a match query")
		assert.Equal(t, str, q.Match)

		return &bleve.SearchResult{
			Hits: []*bs.DocumentMatch{
				&bs.DocumentMatch{Fields: map[string]interface{}{
					"raw": string(lb1),
				}},
				&bs.DocumentMatch{Fields: map[string]interface{}{
					"raw": string(lb2),
				}},
			},
		}, nil
	})

	links, err := searcher.Search(ctx, str)
	assert.NoError(t, err)

	assert.Len(t, links, 2)

	// links[0].data = d1
	data := map[string]interface{}{}
	err = json.Unmarshal(links[0].Data, &data)
	require.NoError(t, err)
	assert.Equal(t, d1, data)

	// links[1].data = d2
	data = map[string]interface{}{}
	err = json.Unmarshal(links[1].Data, &data)
	require.NoError(t, err)
	assert.Equal(t, d2, data)
}
