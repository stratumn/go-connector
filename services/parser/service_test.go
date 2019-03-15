package parser_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	cs "github.com/stratumn/go-chainscript"

	// "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"go-connector/services/livesync/mocksynchronizer"
	"go-connector/services/memorystore/mockmemorystore"
	"go-connector/services/parser"
)

func TestParserService(t *testing.T) {

	// mock service dependencies
	ctrl := gomock.NewController(t)
	synchronizer := mocksynchronizer.NewMockSynchronizer(ctrl)
	memorystore := mockmemorystore.NewMockDB(ctrl)

	// init parser service
	p := parser.Service{}
	p.SetConfig(parser.Config{
		Store: "memorystore",
	})
	p.Plug(map[string]interface{}{
		"livesync":    synchronizer,
		"memorystore": memorystore,
	})

	t.Run("links are saved in store", func(t *testing.T) {
		// add a timeout to the context in case the cancelFunc is not called
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)

		// parser must register to livesync updates
		linksChan := make(chan []*cs.Link)
		synchronizer.EXPECT().Register().Return(linksChan).Times(1)

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
		newLink, _ := cs.NewLinkBuilder("p", "map").Build()
		lh, _ := newLink.Hash()
		lBytes, _ := json.Marshal(newLink)
		memorystore.EXPECT().Put([]byte(lh), lBytes).Do(func(k, v []byte) { cancel() }).Times(1)
		linksChan <- []*cs.Link{newLink}

		<-stoppingCh
	})

	t.Run("returns an error when livesync stops", func(t *testing.T) {
		// add a timeout to the context in case no error is returned
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		// parser must register to livesync updates
		linksChan := make(chan []*cs.Link)
		synchronizer.EXPECT().Register().Return(linksChan).Times(1)

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
		close(linksChan)

		<-stoppingCh
	})

	t.Run("returns an error when saving a link failed", func(t *testing.T) {
		// add a timeout to the context in case no error is returned
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		// parser must register to livesync updates
		linksChan := make(chan []*cs.Link)
		synchronizer.EXPECT().Register().Return(linksChan).Times(1)

		// run service
		runningCh := make(chan struct{})
		stoppingCh := make(chan struct{})
		go func() {
			err := p.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
			assert.EqualError(t, err, "Put failed")
			stoppingCh <- struct{}{}
		}()
		<-runningCh

		// send a new link through the channel and
		// ensure that it triggers an error
		newLink, _ := cs.NewLinkBuilder("p", "map").Build()
		lh, _ := newLink.Hash()
		lBytes, _ := json.Marshal(newLink)
		memorystore.EXPECT().Put([]byte(lh), lBytes).Return(errors.New("Put failed")).Times(1)
		linksChan <- []*cs.Link{newLink}

		<-stoppingCh
	})
}
