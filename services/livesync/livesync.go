package livesync

import (
	"context"
	"fmt"
	"strings"

	logging "github.com/ipfs/go-log"
	cs "github.com/stratumn/go-chainscript"

	"github.com/stratumn/go-connector/services/client"
)

//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mocksynchronizer.go github.com/stratumn/go-connector/services/livesync Synchronizer

// DefaultPagination is the number of links fetched by API call.
const DefaultPagination = 50

var log = logging.Logger("livesync")

// Synchronizer is the type exposed by the livesync service.
type Synchronizer interface {
	Register(WorkflowStates) <-chan []*cs.Segment
}

type synchronizer struct {
	client client.StratumnClient

	// The syncing state of the watched workflows.
	workflowStates WorkflowStates
	// Services subscribing to links updates.
	registeredServices []*listener
}

// WorkflowStates maps the ID of the workflow to the cursor of the last synced link.
type WorkflowStates map[string]string

type listener struct {
	states   WorkflowStates
	listener chan<- []*cs.Segment
}

// NewSycnhronizer returns a new Synchronizer.
// It takes a stratumn client and a list of workflows to sync with.
func NewSycnhronizer(client client.StratumnClient, watchedWorkflows []string) Synchronizer {
	states := make(WorkflowStates, len(watchedWorkflows))
	for _, wfID := range watchedWorkflows {
		states[wfID] = ""
	}

	return &synchronizer{
		client:         client,
		workflowStates: states,
	}
}

// Register subscribes a listener to future updates.
// The listener may pass a WorkflowStates object to specify which workflows it
// wants to receive updates from and from which cursor it should receive updates.
// If nil is passed, the listener will be notified of updates for all synced workflows.
func (s *synchronizer) Register(states WorkflowStates) <-chan []*cs.Segment {
	if states == nil {
		states = make(WorkflowStates, len(s.workflowStates))
		for wfID := range s.workflowStates {
			states[wfID] = ""
		}
	}

	newCh := make(chan []*cs.Segment)
	s.registeredServices = append(s.registeredServices, &listener{
		listener: newCh,
		states:   states,
	})
	return newCh
}

// pollAndNotify fetches all the missing links from the given workflows.
func (s *synchronizer) pollAndNotify(ctx context.Context) error {
	for id := range s.workflowStates {
		variables := map[string]interface{}{
			"id":    id,
			"limit": DefaultPagination,
		}

		rsp := rspData{}
		rsp.WorkflowByRowID.Links.PageInfo.HasNextPage = true
		for rsp.WorkflowByRowID.Links.PageInfo.HasNextPage {
			// the cursor acts as an offset to fetch links from.
			if s.workflowStates[id] != "" {
				variables["cursor"] = s.workflowStates[id]
			}

			err := s.client.CallTraceGql(ctx, pollQuery, variables, &rsp)
			if err != nil {
				log.Errorf("API returned error %s, shutting down..", err)
				s.closeListeners()
				return err
			}

			segments := rsp.WorkflowByRowID.Links.Edges.Segments()
			if len(segments) > 0 {
				fmt.Printf("Synced %d links\n", len(segments))
				s.workflowStates[id] = rsp.WorkflowByRowID.Links.PageInfo.EndCursor
				// send the synced segments to the registered services.
				// compare the current cursor to the cursor specified by each service:
				// - if the current cursor is anterior or equal, do not send any updates.
				// - else send all the segments starting from the service's cursor.
				for _, service := range s.registeredServices {
					// if the service does not subscribe to this workflow, skip it.
					if _, ok := service.states[id]; !ok {
						continue
					}
					switch strings.Compare(s.workflowStates[id], service.states[id]) {
					case -1, 0:
						break
					default:
						service.listener <- rsp.WorkflowByRowID.Links.Edges.Slice(service.states[id])
					}

				}
			}
		}
	}
	return nil
}

func (s *synchronizer) closeListeners() {
	for _, l := range s.registeredServices {
		close(l.listener)
	}
}
