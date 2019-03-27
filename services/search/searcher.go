package search

import (
	"context"
	"encoding/json"

	"github.com/blevesearch/bleve"
	"github.com/pkg/errors"
	cs "github.com/stratumn/go-chainscript"
)

// Searcher is the type sxposed by the search service.
type Searcher interface {
	Search(context.Context, string) ([]*cs.Link, error)
}

type searcher struct {
	idx bleve.Index
}

func newSearcher(idx bleve.Index) Searcher {
	return &searcher{
		idx: idx,
	}
}

// Search searches for all the links that contain a given
// string in either their data or meta.
func (s *searcher) Search(ctx context.Context, str string) ([]*cs.Link, error) {
	res := []*cs.Link{}

	query := bleve.NewMatchQuery(str)
	query.Fuzziness = 1
	search := bleve.NewSearchRequest(query)
	search.Fields = []string{"*"}
	searchResults, err := s.idx.Search(search)
	if err != nil {
		return nil, err
	}

	for _, l := range searchResults.Hits {
		var link cs.Link

		raw, ok := l.Fields["raw"].(string)
		if !ok {
			return nil, errors.New("raw link should be a string")
		}

		err = json.Unmarshal([]byte(raw), &link)
		if err != nil {
			return nil, err
		}
		res = append(res, &link)
	}

	return res, nil
}
