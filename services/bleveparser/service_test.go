package bleveparser_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	cs "github.com/stratumn/go-chainscript"

	"github.com/stretchr/testify/assert"

	parser "github.com/stratumn/go-connector/services/bleveparser"
	"github.com/stratumn/go-connector/services/blevestore/mockblevestore"
	"github.com/stratumn/go-connector/services/livesync/mocksynchronizer"
)

func TestParserService(t *testing.T) {

	// We test the parser with a in-memory bleve index.
	// The main reason I didn't mock it is that batches are not interfaces in bleve...
	// Which makes it pretty hard with gomock.

	// mock service dependencies
	ctrl := gomock.NewController(t)
	synchronizer := mocksynchronizer.NewMockSynchronizer(ctrl)
	mockStore := mockblevestore.NewMockIndex(ctrl)

	// init parser service
	p := parser.Service{}
	p.SetConfig(parser.Config{
		Store: "blevestore",
	})
	p.Plug(map[string]interface{}{
		"livesync":   synchronizer,
		"blevestore": mockStore,
	})

	t.Run("links are saved in store", func(t *testing.T) {
		// add a timeout to the context in case the cancelFunc is not called
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)

		// parser must register to livesync updates
		segmentsChan := make(chan []*cs.Segment)
		synchronizer.EXPECT().Register(gomock.Nil()).Return(segmentsChan).Times(1)

		// run service
		runningCh := make(chan struct{})
		stoppingCh := make(chan struct{})
		go func() {
			err := p.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, context.Canceled.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		// ensure the links are saved in store
		// then cancel the context in order to stop the service

		l1, _ := cs.NewLinkBuilder("p", "map1").Build()
		l2, _ := cs.NewLinkBuilder("p", "map2").Build()
		s1, _ := l1.Segmentify()
		s2, _ := l2.Segmentify()

		// This is an ungly little hack, but there is no other way to have a valid batch...
		i, _ := bleve.NewMemOnly(bleve.NewIndexMapping())
		b := i.NewBatch()

		mockStore.EXPECT().NewBatch().Return(b).Times(1)
		mockStore.EXPECT().Batch(b).Do(func(b *bleve.Batch) {
			fmt.Println(b.Size())
			assert.Equal(t, 2, b.Size())
			cancel()
		}).Times(1)

		segmentsChan <- []*cs.Segment{s1, s2}

		<-stoppingCh
	})

	t.Run("returns an error when livesync stops", func(t *testing.T) {
		// add a timeout to the context in case no error is returned
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		// parser must register to livesync updates
		segmentsChan := make(chan []*cs.Segment)
		synchronizer.EXPECT().Register(gomock.Nil()).Return(segmentsChan).Times(1)

		// run service
		runningCh := make(chan struct{})
		stoppingCh := make(chan struct{})
		go func() {
			err := p.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, parser.ErrSyncStopped.Error())
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		// closing the link channel should trigger an error and stop the service
		close(segmentsChan)

		<-stoppingCh
	})

	t.Run("returns an error when saving a link failed", func(t *testing.T) {
		// add a timeout to the context in case no error is returned
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		// parser must register to livesync updates
		segmentsChan := make(chan []*cs.Segment)
		synchronizer.EXPECT().Register(gomock.Nil()).Return(segmentsChan).Times(1)

		// run service
		runningCh := make(chan struct{})
		stoppingCh := make(chan struct{})
		go func() {
			err := p.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, "wololololo")
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		// send a new link through the channel and
		// ensure that it triggers an error
		l, _ := cs.NewLinkBuilder("p", "map").Build()
		s, _ := l.Segmentify()

		// This is an ungly little hack, but there is no other way to have a valid batch...
		i, _ := bleve.NewMemOnly(bleve.NewIndexMapping())
		b := i.NewBatch()

		mockStore.EXPECT().NewBatch().Return(b).Times(1)
		mockStore.EXPECT().Batch(b).Return(errors.New("wololololo")).Times(1)
		segmentsChan <- []*cs.Segment{s}

		<-stoppingCh
	})
}
