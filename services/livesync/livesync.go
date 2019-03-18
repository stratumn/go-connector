package livesync

import (
	"context"
	"go-connector/services/client"
	"log"

	cs "github.com/stratumn/go-chainscript"
)

//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mocksynchronizer.go github.com/stratumn/go-connector/services/livesync Synchronizer

// Synchronizer is the type exposed by the livesync service.
type Synchronizer interface {
	Register() <-chan []*cs.Link
}

type synchronizer struct {
	client client.StratumnClient

	// The syncing state of the watched workflows.
	workflowStates []*workflowState
	// Services subscribing to links updates.
	registeredServices []chan []*cs.Link
}

type workflowState struct {
	// The IDs of the workflows to pull links from.
	id uint
	// Cursor of the last synced link.
	endCursor string
}

func newSycnhronizer(client client.StratumnClient, watchedWorkflows []uint) *synchronizer {
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
				EndCursor string
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
  ) {
	workflowByRowId(rowId: $id) {
	  id
	  name
	  links(after: $cursor) {
		nodes {
			raw
		}
		pageInfo {
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

func (s *synchronizer) pollAndNotify(ctx context.Context) error {
	if len(s.workflowStates) == 0 {
		return nil
	}

	for _, w := range s.workflowStates {
		variables := map[string]interface{}{
			"id": w.id,
		}
		if w.endCursor != "" {
			variables["cursor"] = w.endCursor
		}
		rsp := rspData{}
		err := s.client.CallTraceGql(ctx, pollQuery, variables, &rsp)

		if err != nil {
			return err
		}

		links := make([]*cs.Link, 0)
		for _, link := range rsp.WorkflowByRowID.Links.Nodes {
			links = append(links, link.Raw)
		}
		if len(links) > 0 {
			log.Printf("Synced %d links\n", len(links))
			w.endCursor = rsp.WorkflowByRowID.Links.PageInfo.EndCursor
		}

		for _, service := range s.registeredServices {
			service <- links
		}

	}
	return nil
}
