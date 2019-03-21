package livesync

import (
	cs "github.com/stratumn/go-chainscript"
)

// pollQuery is the query sent to fetch links from trace API.
const pollQuery = `query workflowLinks(
	$id: BigInt!
	$cursor: Cursor
	$limit: Int!
  ) {
	workflowByRowId(rowId: $id) {
	  id
	  name
	  links(after: $cursor, first: $limit) {
		edges {
			cursor
			node {
				raw
			}
		}
		pageInfo {
		  hasNextPage
		  endCursor
	    }
	  }
	}
  }`

type rspData struct {
	WorkflowByRowID struct {
		Name  string
		Links struct {
			PageInfo struct {
				EndCursor   string
				HasNextPage bool
			}
			Edges linkEdges
		}
	}
}

type linkEdges []struct {
	Cursor string
	Node   struct {
		Raw *cs.Link
	}
}

// Links returns the list of Links in the linkEdges object
func (edges linkEdges) Links() []*cs.Link {
	links := make([]*cs.Link, len(edges))
	for i, link := range edges {
		links[i] = link.Node.Raw
	}
	return links
}

// Slice returns the list of link for which the cursor is positioned after the provided one.
// It assumes the linkEdges are ordered by ascending cursor.
func (edges linkEdges) Slice(cursor string) []*cs.Link {
	links := make([]*cs.Link, 0, len(edges))
	for i := len(edges) - 1; i >= 0; i-- {
		if edges[i].Cursor == cursor {
			return links
		}
		links = append([]*cs.Link{edges[i].Node.Raw}, links...)
	}
	return links
}
