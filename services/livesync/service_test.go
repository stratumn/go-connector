package livesync_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

	"github.com/stratumn/go-connector/services/client/mockclient"
	"github.com/stratumn/go-connector/services/livesync"

	"github.com/stretchr/testify/assert"
)

const (
	rspWithNextPage = `{"workflowByRowId":{"id":"WyJ3b3JrZmxvd3MiLDIxMV0=","name":"AXA CLP - MyData employment status","links":{"edges":[{"cursor":"cursor1","node":{"raw":{"data":"eyJBVVRIX0NPREUiOiJzb21lIGNvZGUiLCJDTElFTlRfSUQiOiJteSBkYXRhIGNsaWVudCBJRCIsIkNMSUVOVF9TRUNSRVQiOiJteSBkYXRhIGNsaWVudCBzZWNyZXQifQ==","meta":{"data":"eyJjcmVhdGVkQnlJZCI6IjEyNyIsImZvcm1JZCI6IjI5NDUiLCJncm91cElkIjoiNjQ4IiwiaW5wdXRzIjpudWxsLCJsYXN0Rm9ybUlkIjoiMjk0NCIsIm93bmVySWQiOiIxNjQifQ==","mapId":"ddf953b8-d1fb-4e0d-8416-320c823760d9","action":"Consent request","process":{"name":"211","state":"FREE"},"clientId":"github.com/stratumn/go-chainscript","priority":2,"outDegree":1,"prevLinkHash":"z4iiGAePyFOs9gzgI2ITGU6XVqh7kLRkmflTTqZi100="},"version":"1.0.0","signatures":[{"version":"1.0.0","publicKey":"LS0tLS1CRUdJTiBFRDI1NTE5IFBVQkxJQyBLRVktLS0tLQpNQ293QlFZREsyVndBeUVBNXdrYkdSMmR5cHp3Mm1ma0RXWi9xaCtSNW5DNlVDN0k5WmgxUU9BU1NsWT0KLS0tLS1FTkQgRUQyNTUxOSBQVUJMSUMgS0VZLS0tLS0K","signature":"LS0tLS1CRUdJTiBNRVNTQUdFLS0tLS0KamQxRVN2T0pNbnBoTGFqWHQ1cC9nbnJmbFoxVjE1M3haQUxYK01TSFBpTnA1WHEyenhkc2J4bnBpd2UxWVZHQgpIVCtwdEFkSkVEczhWVkFheWhXUERRPT0KLS0tLS1FTkQgTUVTU0FHRS0tLS0tCg==","payloadPath":"[version,data,meta]"}]}}},{"cursor":"cursor2","node":{"raw":{"data":"eyJVU0VSX0lEIjoiMjMifQ==","meta":{"data":"eyJjcmVhdGVkQnlJZCI6IjEyNyIsImZvcm1JZCI6IjI5NDQiLCJncm91cElkIjoiNjQ4IiwiaW5wdXRzIjpudWxsLCJvd25lcklkIjoiMTY0In0=","tags":["USER-ID-23"],"mapId":"ddf953b8-d1fb-4e0d-8416-320c823760d9","action":"Initialization","process":{"name":"211","state":"FREE"},"clientId":"github.com/stratumn/go-chainscript","priority":1,"outDegree":1},"version":"1.0.0","signatures":[{"version":"1.0.0","publicKey":"LS0tLS1CRUdJTiBFRDI1NTE5IFBVQkxJQyBLRVktLS0tLQpNQ293QlFZREsyVndBeUVBNXdrYkdSMmR5cHp3Mm1ma0RXWi9xaCtSNW5DNlVDN0k5WmgxUU9BU1NsWT0KLS0tLS1FTkQgRUQyNTUxOSBQVUJMSUMgS0VZLS0tLS0K","signature":"LS0tLS1CRUdJTiBNRVNTQUdFLS0tLS0KenpMRDhXNFNLOGNKTWxEdlpxdEVUVGxzWGIxM0VYd2QrT1ZpRERaVHlzYVZGbFdtRzc5KzIxSkFmU3pkdjFQagpGTGE3NU9OQWU4SjRMd0ZxMnJFTEFRPT0KLS0tLS1FTkQgTUVTU0FHRS0tLS0tCg==","payloadPath":"[version,data,meta]"}]}}}],"pageInfo":{"hasNextPage":true,"endCursor":"endCursor"}}}}`
	rspLastPage     = `{"workflowByRowId":{"id":"WyJ3b3JrZmxvd3MiLDIxMV0=","name":"AXA CLP - MyData employment status","links":{"edges":[{"cursor":"cursor3","node":{"raw":{"data":"eyJVU0VSX0lEIjoiMjMifQ==","meta":{"data":"eyJjcmVhdGVkQnlJZCI6IjEyNyIsImZvcm1JZCI6IjI5NDQiLCJncm91cElkIjoiNjQ4IiwiaW5wdXRzIjpudWxsLCJvd25lcklkIjoiMTY0In0=","tags":["USER-ID-23"],"mapId":"ddf953b8-d1fb-4e0d-8416-320c823760d9","action":"Initialization","process":{"name":"211","state":"FREE"},"clientId":"github.com/stratumn/go-chainscript","priority":1,"outDegree":1},"version":"1.0.0","signatures":[{"version":"1.0.0","publicKey":"LS0tLS1CRUdJTiBFRDI1NTE5IFBVQkxJQyBLRVktLS0tLQpNQ293QlFZREsyVndBeUVBNXdrYkdSMmR5cHp3Mm1ma0RXWi9xaCtSNW5DNlVDN0k5WmgxUU9BU1NsWT0KLS0tLS1FTkQgRUQyNTUxOSBQVUJMSUMgS0VZLS0tLS0K","signature":"LS0tLS1CRUdJTiBNRVNTQUdFLS0tLS0KenpMRDhXNFNLOGNKTWxEdlpxdEVUVGxzWGIxM0VYd2QrT1ZpRERaVHlzYVZGbFdtRzc5KzIxSkFmU3pkdjFQagpGTGE3NU9OQWU4SjRMd0ZxMnJFTEFRPT0KLS0tLS1FTkQgTUVTU0FHRS0tLS0tCg==","payloadPath":"[version,data,meta]"}]}}}],"pageInfo":{"hasNextPage":false,"endCursor":"endCursor"}}}}`
	rspWithoutLinks = `{"workflowByRowId":{"id":"WyJ3b3JrZmxvd3MiLDIxMV0=","name":"AXA CLP - MyData employment status","links":{"edges":[],"pageInfo":{"hasNextPage":false,"endCursor":"endCursor"}}}}`
)

var (
	watchedWorkflows = []uint{1, 2}
)

func TestLivesyncService(t *testing.T) {

	ctrl := gomock.NewController(t)
	runningCh := make(chan struct{})
	stoppingCh := make(chan struct{})

	t.Run("Polls the API and send updates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
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

		go func() {
			err := s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, context.Canceled.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		synchronizer := s.Expose().(livesync.Synchronizer)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithNextPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "cursor": "endCursor", "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspLastPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[1], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(2)

		updates := synchronizer.Register(nil)

		// wait for the first update
		links := <-updates
		assert.Len(t, links, 2)

		// wait for the second update
		links = <-updates
		assert.Len(t, links, 1)

		cancel()
		<-stoppingCh
	})

	t.Run("Subscribe to specific workflow", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
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

		go func() {
			err := s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, context.Canceled.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		synchronizer := s.Expose().(livesync.Synchronizer)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspLastPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[1], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspLastPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()

		// only subscribe to the first workflow
		updates := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: ""})

		// wait for the first update
		links := <-updates
		assert.Len(t, links, 1)

		select {
		case <-updates:
			t.Error("should not receive updates")
		case <-time.NewTicker(time.Millisecond * 20).C:
			break
		}

		cancel()
		<-stoppingCh
	})

	t.Run("Subscribe from specific cursor", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
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

		go func() {
			err := s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, context.Canceled.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		synchronizer := s.Expose().(livesync.Synchronizer)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithNextPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[1], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspLastPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()

		// only subscribe to the first workflow from cursor1
		updates := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: "cursor1"})

		// wait for the first update
		// it should only containe one link since we started at cursor1 (the only remaining cursor is cursor2)
		links := <-updates
		assert.Len(t, links, 1)

		select {
		case <-updates:
			t.Error("should not receive updates")
		case <-time.NewTicker(time.Millisecond * 20).C:
			break
		}

		cancel()
		<-stoppingCh
	})
	t.Run("Returns an error when the API call failed", func(t *testing.T) {
		apiError := errors.New("error")

		ctx := context.Background()
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
