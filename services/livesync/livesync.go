package livesync

import (
	"context"
	"strings"

	"github.com/stratumn/go-connector/services/client"

	logging "github.com/ipfs/go-log"
	cs "github.com/stratumn/go-chainscript"
)

//go:generate mockgen -package mocksynchronizer -destination mocksynchronizer/mocksynchronizer.go github.com/stratumn/go-connector/services/livesync Synchronizer

// DefaultPagination is the number of links fetched by API call.
const DefaultPagination = 50

var log = logging.Logger("livesync")

// Synchronizer is the type exposed by the livesync service.
type Synchronizer interface {
	Register(WorkflowStates) <-chan []*cs.Link
}

type synchronizer struct {
	client client.StratumnClient

	// The syncing state of the watched workflows.
	workflowStates WorkflowStates
	// Services subscribing to links updates.
	registeredServices []*listener
}

// WorkflowStates maps the ID of the workflow to the cursor of the last synced link.
type WorkflowStates map[uint]string

type listener struct {
	states   WorkflowStates
	listener chan<- []*cs.Link
}

// NewSycnhronizer returns a new Synchronizer.
// It takes a stratumn client and a list of workflows to sync with.
func NewSycnhronizer(client client.StratumnClient, watchedWorkflows []uint) Synchronizer {
	states := WorkflowStates{}
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
func (s *synchronizer) Register(states WorkflowStates) <-chan []*cs.Link {
	if states == nil {
		states = WorkflowStates{}
		for wfID := range s.workflowStates {
			states[wfID] = ""
		}
	}

	newCh := make(chan []*cs.Link)
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
				return err
			}

			links := rsp.WorkflowByRowID.Links.Edges.Links()
			if len(links) > 0 {
				log.Infof("Synced %d links\n", len(links))
				s.workflowStates[id] = rsp.WorkflowByRowID.Links.PageInfo.EndCursor
				// send the synced links to the registered services.
				// compare the current cursor to the cursor specified by each service:
				// - if the current cursor is anterior or equal, do not send any updates.
				// - else send all the links starting from the service's cursor.
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
