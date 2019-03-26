package livesync

import (
	"encoding/hex"

	cs "github.com/stratumn/go-chainscript"
)

// pollQuery is the query sent to fetch segments from trace API.
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
				linkHash
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
		Raw      *cs.Link
		LinkHash string
	}
}

// Segments returns the list of Segments from the linkEdges object
func (edges linkEdges) Segments() []*cs.Segment {
	segments := make([]*cs.Segment, len(edges))
	for i, link := range edges {
		lh, _ := hex.DecodeString(link.Node.LinkHash)
		segments[i] = &cs.Segment{Link: link.Node.Raw, Meta: &cs.SegmentMeta{LinkHash: lh}}
	}
	return segments
}

// Slice returns the list of link for which the cursor is positioned after the provided one.
// It assumes the linkEdges are ordered by ascending cursor.
func (edges linkEdges) Slice(cursor string) []*cs.Segment {
	segments := make([]*cs.Segment, 0, len(edges))
	for i := len(edges) - 1; i >= 0; i-- {
		if edges[i].Cursor == cursor {
			return segments
		}
		lh, _ := hex.DecodeString(edges[i].Node.LinkHash)
		segments = append([]*cs.Segment{
			&cs.Segment{
				Link: edges[i].Node.Raw,
				Meta: &cs.SegmentMeta{LinkHash: lh},
			},
		}, segments...)
	}
	return segments
}
