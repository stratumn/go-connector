package livesync_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

	"github.com/stratumn/go-connector/services/client/mockclient"
	"github.com/stratumn/go-connector/services/livesync"

	"github.com/stretchr/testify/assert"
)

const (
	rspData1 = `{"workflowByRowId":{"id":"WORKFLOWID1","name":"Some Workflow","links":{"pageInfo":{"endCursor":"endCursor","hasNextPage":true},"nodes":[{"id":"ID1","raw":{"data":"data","meta":{"data":"metadata","mapId":"06d158e6-9cd0-460a-8ef9-fac1b7883fbc"},"version":"1.0.0","signatures":[]}},{"id":"ID2","raw":{"data":"data","meta":{"data":"metadata","mapId":"06d158e6-9cd0-460a-8ef9-fac1b7883fbc","prevLinkHash":"KfaCRwiJAN9QpYKc3q9+/ryIKhdjD7DdhmaCMkQPwyc="},"version":"1.0.0","signatures":[]}}]}}}`
	rspData2 = `{"workflowByRowId":{"id":"WORKFLOWID1","name":"Some Workflow","links":{"pageInfo":{"endCursor":"endCursor", "hasNextPage":false},"nodes":[{"id":"ID1","raw":{"data":"data","meta":{"data":"metadata","mapId":"06d158e6-9cd0-460a-8ef9-fac1b7883fbc"},"version":"1.0.0","signatures":[]}}]}}}`
	rspData3 = `{"workflowByRowId":{"id":"WORKFLOWID1","name":"Some Workflow","links":{"pageInfo":{"endCursor":"endCursor", "hasNextPage":false}}}}`
)

var (
	watchedWorkflows = []uint{1, 2}
)

func TestLivesyncService(t *testing.T) {

	ctrl := gomock.NewController(t)
	client := mockclient.NewMockStratumnClient(ctrl)
	config := livesync.Config{
		PollInterval:     10,
		WatchedWorkflows: watchedWorkflows,
	}

	s := &livesync.Service{}
	s.SetConfig(config)
	s.Plug(map[string]interface{}{
		"stratumnClient": client,
	})
	runningCh := make(chan struct{})
	stoppingCh := make(chan struct{})

	t.Run("Polls the API and send updates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			err := s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, context.Canceled.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		synchronizer := s.Expose().(livesync.Synchronizer)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspData1), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "cursor": "endCursor", "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspData2), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[1], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspData3), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(2)

		updates := synchronizer.Register()

		// wait for the first update
		links := <-updates
		assert.Len(t, links, 2)

		// wait for the second update
		links = <-updates
		assert.Len(t, links, 1)

		cancel()
		<-stoppingCh
	})

	t.Run("Returns an error when the API call failed", func(t *testing.T) {
		apiError := errors.New("error")

		ctx := context.Background()

		go func() {
			err := s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, apiError.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(apiError)

		<-stoppingCh
	})

}
