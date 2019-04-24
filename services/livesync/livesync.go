package livesync

import (
	"context"
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
	Register(WorkflowStates) (<-chan []*cs.Segment, error)
}

type synchronizer struct {
	client client.StratumnClient

	// The syncing state of the watched workflows.
	workflowStates WorkflowStates
	// Services subscribing to links updates.
	registeredServices []*listener
}

// WorkflowState maps the ID of the workflow to the cursor of the last synced link.
type WorkflowState struct {
	ID     string
	Cursor string
}

// WorkflowStates is a list of WorkflowState
type WorkflowStates []*WorkflowState

// Get finds a WorkflowState in the list given its ID.
func (wfStates WorkflowStates) Get(id string) (*WorkflowState, bool) {
	for _, w := range wfStates {
		if w.ID == id {
			return w, true
		}
	}
	return nil, false
}

type listener struct {
	states   WorkflowStates
	listener chan<- []*cs.Segment
}

// NewSycnhronizer returns a new Synchronizer.
// It takes a stratumn client and a list of workflows to sync with.
func NewSycnhronizer(client client.StratumnClient, watchedWorkflows []string) Synchronizer {
	states := make(WorkflowStates, len(watchedWorkflows))
	for i, wfID := range watchedWorkflows {
		states[i] = &WorkflowState{ID: wfID, Cursor: ""}
	}

	return &synchronizer{
		client:         client,
		workflowStates: states,
	}
}

// Register subscribes a listener to future updates.
// The listener may pass a WorkflowStates object to specify which workflows it
// wants to receive updates from and from which cursor it should receive updates.
// The livesync automatically subscribe to the workflow if it is not already the case.
// If nil is passed, the listener will be notified of updates for all synced workflows.
func (s *synchronizer) Register(states WorkflowStates) (<-chan []*cs.Segment, error) {
	for _, w := range states {
		if livesyncState, ok := s.workflowStates.Get(w.ID); !ok {
			s.workflowStates = append(s.workflowStates, &WorkflowState{ID: w.ID, Cursor: w.Cursor})
		} else if ok && strings.Compare(w.Cursor, livesyncState.Cursor) == -1 {
			gap, err := CompareCursors(w.Cursor, livesyncState.Cursor)
			if err != nil {
				return nil, err
			}
			// if a service register for updates in the past, lower the current end cursor.
			if gap < 0 {
				livesyncState.Cursor = w.Cursor
			}
		}
	}
	if states == nil {
		states = make(WorkflowStates, len(s.workflowStates))
		for i, w := range s.workflowStates {
			states[i] = &WorkflowState{ID: w.ID, Cursor: ""}
		}
	}

	newCh := make(chan []*cs.Segment)
	s.registeredServices = append(s.registeredServices, &listener{
		listener: newCh,
		states:   states,
	})
	return newCh, nil
}

// pollAndNotify fetches all the missing links from the given workflows.
func (s *synchronizer) pollAndNotify(ctx context.Context) error {
	for _, w := range s.workflowStates {
		variables := map[string]interface{}{
			"id":    w.ID,
			"limit": DefaultPagination,
		}

		rsp := rspData{}
		rsp.WorkflowByRowID.Links.PageInfo.HasNextPage = true
		for rsp.WorkflowByRowID.Links.PageInfo.HasNextPage {
			// the cursor acts as an offset to fetch links from.
			if w.Cursor != "" {
				variables["cursor"] = w.Cursor
			}

			err := s.client.CallTraceGql(ctx, pollQuery, variables, &rsp)
			if err != nil {
				log.Errorf("API returned error %s, keeping running...", err)
				break
			}

			segments, err := rsp.WorkflowByRowID.Links.Edges.Segments()
			if err != nil {
				s.closeListeners()
				return err
			}
			if len(segments) > 0 {
				log.Infof("Synced %d links\n", len(segments))
				w.Cursor = rsp.WorkflowByRowID.Links.PageInfo.EndCursor
				// send the synced segments to the registered services.
				// compare the current cursor to the cursor specified by each service:
				// - if the current cursor is anterior or equal, do not send any updates.
				// - else send all the segments starting from the service's cursor.
				for _, service := range s.registeredServices {
					serviceState, ok := service.states.Get(w.ID)
					// if the service does not subscribe to this workflow, skip it.
					if !ok {
						continue
					}
					gap, err := CompareCursors(w.Cursor, serviceState.Cursor)
					if err != nil {
						log.Errorf("error comparing cursors: %s", err)
					}
					if gap > 0 {
						service.listener <- rsp.WorkflowByRowID.Links.Edges.Slice(serviceState.Cursor)
						serviceState.Cursor = w.Cursor
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
