package livesync_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	cs "github.com/stratumn/go-chainscript"
	"github.com/stretchr/testify/assert"

	"github.com/stratumn/go-connector/services/client/mockclient"
	"github.com/stratumn/go-connector/services/livesync"
)

var (
	watchedWorkflows = []string{"1", "2"}
	cursor1          = makeCursor(1)
	cursor2          = makeCursor(2)
	cursor3          = makeCursor(3)

	rspWithNextPage = fmt.Sprintf(`{"workflowByRowId":{"id":"WyJ3b3JrZmxvd3MiLDIxMV0=","name":"AXA CLP - MyData employment status","links":{"edges":[{"cursor":"%s","node":{"linkHash":"deadbeefdeadbeef","raw":{"data":"eyJBVVRIX0NPREUiOiJzb21lIGNvZGUiLCJDTElFTlRfSUQiOiJteSBkYXRhIGNsaWVudCBJRCIsIkNMSUVOVF9TRUNSRVQiOiJteSBkYXRhIGNsaWVudCBzZWNyZXQifQ==","meta":{"data":"eyJjcmVhdGVkQnlJZCI6IjEyNyIsImZvcm1JZCI6IjI5NDUiLCJncm91cElkIjoiNjQ4IiwiaW5wdXRzIjpudWxsLCJsYXN0Rm9ybUlkIjoiMjk0NCIsIm93bmVySWQiOiIxNjQifQ==","mapId":"ddf953b8-d1fb-4e0d-8416-320c823760d9","action":"Consent request","process":{"name":"211","state":"FREE"},"clientId":"github.com/stratumn/go-chainscript","priority":2,"outDegree":1,"prevLinkHash":"z4iiGAePyFOs9gzgI2ITGU6XVqh7kLRkmflTTqZi100="},"version":"1.0.0","signatures":[{"version":"1.0.0","publicKey":"LS0tLS1CRUdJTiBFRDI1NTE5IFBVQkxJQyBLRVktLS0tLQpNQ293QlFZREsyVndBeUVBNXdrYkdSMmR5cHp3Mm1ma0RXWi9xaCtSNW5DNlVDN0k5WmgxUU9BU1NsWT0KLS0tLS1FTkQgRUQyNTUxOSBQVUJMSUMgS0VZLS0tLS0K","signature":"LS0tLS1CRUdJTiBNRVNTQUdFLS0tLS0KamQxRVN2T0pNbnBoTGFqWHQ1cC9nbnJmbFoxVjE1M3haQUxYK01TSFBpTnA1WHEyenhkc2J4bnBpd2UxWVZHQgpIVCtwdEFkSkVEczhWVkFheWhXUERRPT0KLS0tLS1FTkQgTUVTU0FHRS0tLS0tCg==","payloadPath":"[version,data,meta]"}]}}},{"cursor":"%s","node":{"raw":{"data":"eyJVU0VSX0lEIjoiMjMifQ==","meta":{"data":"eyJjcmVhdGVkQnlJZCI6IjEyNyIsImZvcm1JZCI6IjI5NDQiLCJncm91cElkIjoiNjQ4IiwiaW5wdXRzIjpudWxsLCJvd25lcklkIjoiMTY0In0=","tags":["USER-ID-23"],"mapId":"ddf953b8-d1fb-4e0d-8416-320c823760d9","action":"Initialization","process":{"name":"211","state":"FREE"},"clientId":"github.com/stratumn/go-chainscript","priority":1,"outDegree":1},"version":"1.0.0","signatures":[{"version":"1.0.0","publicKey":"LS0tLS1CRUdJTiBFRDI1NTE5IFBVQkxJQyBLRVktLS0tLQpNQ293QlFZREsyVndBeUVBNXdrYkdSMmR5cHp3Mm1ma0RXWi9xaCtSNW5DNlVDN0k5WmgxUU9BU1NsWT0KLS0tLS1FTkQgRUQyNTUxOSBQVUJMSUMgS0VZLS0tLS0K","signature":"LS0tLS1CRUdJTiBNRVNTQUdFLS0tLS0KenpMRDhXNFNLOGNKTWxEdlpxdEVUVGxzWGIxM0VYd2QrT1ZpRERaVHlzYVZGbFdtRzc5KzIxSkFmU3pkdjFQagpGTGE3NU9OQWU4SjRMd0ZxMnJFTEFRPT0KLS0tLS1FTkQgTUVTU0FHRS0tLS0tCg==","payloadPath":"[version,data,meta]"}]}}}],"pageInfo":{"hasNextPage":true,"endCursor":"%s"}}}}`, cursor1, cursor2, cursor2)
	rspLastPage     = fmt.Sprintf(`{"workflowByRowId":{"id":"WyJ3b3JrZmxvd3MiLDIxMV0=","name":"AXA CLP - MyData employment status","links":{"edges":[{"cursor":"%s","node":{"raw":{"data":"eyJVU0VSX0lEIjoiMjMifQ==","meta":{"data":"eyJjcmVhdGVkQnlJZCI6IjEyNyIsImZvcm1JZCI6IjI5NDQiLCJncm91cElkIjoiNjQ4IiwiaW5wdXRzIjpudWxsLCJvd25lcklkIjoiMTY0In0=","tags":["USER-ID-23"],"mapId":"ddf953b8-d1fb-4e0d-8416-320c823760d9","action":"Initialization","process":{"name":"211","state":"FREE"},"clientId":"github.com/stratumn/go-chainscript","priority":1,"outDegree":1},"version":"1.0.0","signatures":[{"version":"1.0.0","publicKey":"LS0tLS1CRUdJTiBFRDI1NTE5IFBVQkxJQyBLRVktLS0tLQpNQ293QlFZREsyVndBeUVBNXdrYkdSMmR5cHp3Mm1ma0RXWi9xaCtSNW5DNlVDN0k5WmgxUU9BU1NsWT0KLS0tLS1FTkQgRUQyNTUxOSBQVUJMSUMgS0VZLS0tLS0K","signature":"LS0tLS1CRUdJTiBNRVNTQUdFLS0tLS0KenpMRDhXNFNLOGNKTWxEdlpxdEVUVGxzWGIxM0VYd2QrT1ZpRERaVHlzYVZGbFdtRzc5KzIxSkFmU3pkdjFQagpGTGE3NU9OQWU4SjRMd0ZxMnJFTEFRPT0KLS0tLS1FTkQgTUVTU0FHRS0tLS0tCg==","payloadPath":"[version,data,meta]"}]}}}],"pageInfo":{"hasNextPage":false,"endCursor":"%s"}}}}`, cursor3, cursor3)
	rspWithoutLinks = `{"workflowByRowId":{"id":"WyJ3b3JrZmxvd3MiLDIxMV0=","name":"AXA CLP - MyData employment status","links":{"edges":[],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}`
)

func makeCursor(index int) string {
	cursor := fmt.Sprintf(`["natural",%d]`, index)
	return base64.StdEncoding.EncodeToString([]byte(cursor))
}

func TestLivesyncService(t *testing.T) {

	ctrl := gomock.NewController(t)
	stoppingCh := make(chan struct{})

	t.Run("Polls the API and send updates", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

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

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithNextPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "cursor": cursor2, "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspLastPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "cursor": cursor3, "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[1], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()

		synchronizer := s.Expose().(livesync.Synchronizer)
		updates, err := synchronizer.Register(nil)
		require.NoError(t, err)

		go func() {
			// wait for the first update
			segments := <-updates
			assert.Len(t, segments, 2)
			firstLinkHash, _ := hex.DecodeString("deadbeefdeadbeef")

			assert.Equal(t, segments[0].LinkHash(), cs.LinkHash(firstLinkHash))

			// wait for the other updates
			for {
				select {
				case segments, ok := <-updates:
					if !ok {
						stoppingCh <- struct{}{}
					} else {
						assert.Len(t, segments, 1)
					}
				}
			}

		}()

		err = s.Run(ctx, func() {}, func() {})
		assert.EqualError(t, err, context.DeadlineExceeded.Error())
		<-stoppingCh
	})

	t.Run("Automatically polls a workflow when registering to it", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		client := mockclient.NewMockStratumnClient(ctrl)
		// do not watch any workflow
		config := livesync.Config{
			PollInterval:     10,
			WatchedWorkflows: []string{},
		}
		s := &livesync.Service{}
		s.SetConfig(config)
		s.Plug(map[string]interface{}{
			"stratumnClient": client,
		})

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithNextPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "cursor": cursor2, "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspLastPage), rsp)
				assert.NoError(t, err)
				return nil
			}).Times(1)
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "cursor": cursor3, "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()
		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[1], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithoutLinks), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()

		synchronizer := s.Expose().(livesync.Synchronizer)
		// subscribe to the first workflow
		updates, err := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: ""})
		require.NoError(t, err)

		go func() {
			// wait for the first update
			segments := <-updates
			assert.Len(t, segments, 2)
			segments = <-updates
			assert.Len(t, segments, 1)
			stoppingCh <- struct{}{}
		}()

		err = s.Run(ctx, func() {}, func() {})
		assert.EqualError(t, err, context.DeadlineExceeded.Error())

		<-stoppingCh
	})

	t.Run("Subscribe to specific workflow", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

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
		synchronizer := s.Expose().(livesync.Synchronizer)
		updates, err := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: ""})
		require.NoError(t, err)

		go func() {
			// wait for the first update
			segments := <-updates
			assert.Len(t, segments, 1)

			select {
			case _, ok := <-updates:
				if ok {
					t.Error("should not receive updates")
				}
			case <-time.After(time.Millisecond * 20):
				break
			}
			stoppingCh <- struct{}{}
		}()

		err = s.Run(ctx, func() {}, func() {})
		assert.EqualError(t, err, context.DeadlineExceeded.Error())

		<-stoppingCh
	})

	t.Run("Subscribe from specific cursor", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

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
		synchronizer := s.Expose().(livesync.Synchronizer)
		updates, err := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: cursor1})
		require.NoError(t, err)

		go func() {
			// wait for the first update
			// it should only contain one segment since we started at cursor1 (the only remaining cursor is cursor2)
			segments := <-updates
			assert.Len(t, segments, 1)
			select {
			case _, ok := <-updates:
				if ok {
					t.Error("should not receive updates")
				}
			case <-time.After(time.Millisecond * 20):
				break
			}
			stoppingCh <- struct{}{}

		}()

		err = s.Run(ctx, func() {}, func() {})
		assert.EqualError(t, err, context.DeadlineExceeded.Error())

		<-stoppingCh
	})

	t.Run("Lower the cursor when registering to past updates", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

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

		client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Eq(map[string]interface{}{"id": watchedWorkflows[0], "limit": livesync.DefaultPagination}), gomock.Any()).
			DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				err := json.Unmarshal([]byte(rspWithNextPage), rsp)
				assert.NoError(t, err)
				return nil
			}).AnyTimes()

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

		synchronizer := s.Expose().(livesync.Synchronizer)

		subscriber1, err := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: cursor1})
		require.NoError(t, err)

		go func() {
			// wait for the first update.
			// it should only contain one segment since we started at cursor1 (the only remaining cursor is cursor2).
			segments := <-subscriber1
			assert.Len(t, segments, 1)

			// a second subscriber registers to updates from the beginning,
			// therefore it should receive 2 updates.
			subscriber2, err := synchronizer.Register(livesync.WorkflowStates{watchedWorkflows[0]: ""})
			require.NoError(t, err)
			segments = <-subscriber2
			assert.Len(t, segments, 2)

			// both subscribers have received all updates, they should not be notified anymore.
			select {
			case _, ok := <-subscriber1:
				if ok {
					t.Error("should not receive updates")
				}
			case _, ok := <-subscriber2:
				if ok {
					t.Error("should not receive updates")
				}
			case <-time.After(time.Millisecond * 20):
				break
			}
			stoppingCh <- struct{}{}
		}()

		err = s.Run(ctx, func() {}, func() {})
		assert.EqualError(t, err, context.DeadlineExceeded.Error())

		<-stoppingCh
	})

	t.Run("Keep running whent he API call failed", func(t *testing.T) {
		apiError := errors.New("error")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

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

		gomock.InOrder(
			client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(apiError).Times(3),
			client.EXPECT().CallTraceGql(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
				cancel()
				return apiError
			}).AnyTimes(),
		)

		err := s.Run(ctx, func() {}, func() {})
		assert.EqualError(t, err, context.Canceled.Error())

		<-stoppingCh
	})

}

func TestCompareCursors(t *testing.T) {
	t.Run("Returns the difference between two cursors", func(t *testing.T) {
		c1 := makeCursor(1)
		c2 := makeCursor(4)

		diff, err := livesync.CompareCursors(c2, c1)
		require.NoError(t, err)
		assert.Equal(t, 3, diff)

		diff, err = livesync.CompareCursors(c1, c2)
		require.NoError(t, err)
		assert.Equal(t, -3, diff)
	})

	t.Run("Returns the difference when a cursor is empty", func(t *testing.T) {
		c1 := ""
		c2 := makeCursor(1)

		diff, err := livesync.CompareCursors(c2, c1)
		require.NoError(t, err)
		assert.Equal(t, 1, diff)

		diff, err = livesync.CompareCursors(c1, c2)
		require.NoError(t, err)
		assert.Equal(t, -1, diff)
	})

	t.Run("fails when a cursor is not base64 encoded", func(t *testing.T) {
		c1 := "wow"
		c2 := makeCursor(1)

		_, err := livesync.CompareCursors(c2, c1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wow: bad cursor")
	})

	t.Run("fails when a cursor is not an array", func(t *testing.T) {
		c1 := base64.StdEncoding.EncodeToString([]byte("wow"))
		c2 := makeCursor(1)

		_, err := livesync.CompareCursors(c2, c1)
		assert.EqualError(t, err, "wow: cursor is not an array")
	})

	t.Run("fails when a cursor has bad length", func(t *testing.T) {
		c1 := base64.StdEncoding.EncodeToString([]byte(`["wow"]`))
		c2 := makeCursor(1)

		_, err := livesync.CompareCursors(c2, c1)
		assert.EqualError(t, err, `["wow"]: cursor is not an array`)
	})

	t.Run("fails when cursor has no index", func(t *testing.T) {
		c1 := base64.StdEncoding.EncodeToString([]byte(`["wow", "amazing"]`))
		c2 := makeCursor(1)

		_, err := livesync.CompareCursors(c2, c1)
		assert.EqualError(t, err, `["wow", "amazing"]: cursor does not have an index`)
	})
}
