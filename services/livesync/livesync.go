package livesync

import (
	"context"

	"go-connector/services/client"

	logging "github.com/ipfs/go-log"
	cs "github.com/stratumn/go-chainscript"
)

//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mocksynchronizer.go github.com/stratumn/go-connector/services/livesync Synchronizer

// DefaultPagination is the number of links fetched by API call.
const DefaultPagination = 50

var log = logging.Logger("livesync")

// Synchronizer is the type exposed by the livesync service.
type Synchronizer interface {
	Register() <-chan []*cs.Link
}

type synchronizer struct {
	client client.StratumnClient

	// The syncing state of the watched workflows.
	workflowStates []*workflowState
	// Services subscribing to links updates.
	registeredServices []chan<- []*cs.Link
}

type workflowState struct {
	// The ID of the workflow to pull links from.
	id uint
	// Cursor of the last synced link.
	endCursor string
}

// NewSycnhronizer returns a new Synchronizer
func NewSycnhronizer(client client.StratumnClient, watchedWorkflows []uint) Synchronizer {
	states := make([]*workflowState, len(watchedWorkflows))
	for i, w := range watchedWorkflows {
		states[i] = &workflowState{
			id: w,
		}
	}
	return &synchronizer{
		client:         client,
		workflowStates: states,
	}
}

type rspData struct {
	WorkflowByRowID struct {
		Name  string
		Links struct {
			PageInfo struct {
				EndCursor   string
				HasNextPage bool
			}
			Nodes []struct {
				Raw *cs.Link
			}
		}
	}
}

const pollQuery = `query workflowLinks(
	$id: BigInt!
	$cursor: Cursor
	$limit: Int!
  ) {
	workflowByRowId(rowId: $id) {
	  id
	  name
	  links(after: $cursor, first: $pagination) {
		nodes {
			raw
		}
		pageInfo {
		  hasNextPage
		  endCursor
	    }
	  }
	}
  }`

func (s *synchronizer) Register() <-chan []*cs.Link {
	newCh := make(chan []*cs.Link)
	s.registeredServices = append(s.registeredServices, newCh)
	return newCh
}

// pollAndNotify fetches all the missing links from the given workflows.
func (s *synchronizer) pollAndNotify(ctx context.Context) error {
	for _, w := range s.workflowStates {
		variables := map[string]interface{}{
			"id":    w.id,
			"limit": DefaultPagination,
		}

		rsp := rspData{}
		rsp.WorkflowByRowID.Links.PageInfo.HasNextPage = true
		for rsp.WorkflowByRowID.Links.PageInfo.HasNextPage {
			// the cursor acts as an offset to fetch links from.
			if w.endCursor != "" {
				variables["cursor"] = w.endCursor
			}

			err := s.client.CallTraceGql(ctx, pollQuery, variables, &rsp)
			if err != nil {
				return err
			}

			links := make([]*cs.Link, len(rsp.WorkflowByRowID.Links.Nodes))
			for i, link := range rsp.WorkflowByRowID.Links.Nodes {
				links[i] = link.Raw
			}

			if len(links) > 0 {
				log.Infof("Synced %d links\n", len(links))
				// remember the cursor of the last returned link.
				w.endCursor = rsp.WorkflowByRowID.Links.PageInfo.EndCursor
				// send the synced links to the registered services.
				for _, service := range s.registeredServices {
					service <- links
				}
			}
		}
	}
	return nil
}
