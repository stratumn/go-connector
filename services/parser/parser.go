package parser

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	cs "github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-node/core/db"

	"github.com/stratumn/go-connector/services/livesync"
)

var (
	// ErrSyncStopped is returned when the subscription channel is closed by the synchronizer service.
	ErrSyncStopped = errors.New("synchronizer service stopped")

	// LinkPrefix is used to prefix keys in the store.
	LinkPrefix = []byte("link")
)

type parser struct {
	db           db.DB
	synchronizer livesync.Synchronizer
}

// saveLinks stores the links in the key/value store.
// links are indexed by linkHash and are serialized to JSON.
func (p *parser) saveLinks(ctx context.Context, links []*cs.Link) error {
	for _, link := range links {

		lh, err := link.Hash()
		if err != nil {
			return err
		}

		linkBytes, err := json.Marshal(link)
		if err != nil {
			return err
		}
		err = p.db.Put(append(LinkPrefix, lh...), linkBytes)

		if err != nil {
			return err
		}
	}
	return nil
}

// run subscribes to the livesync service and waits for updates.
// It returns an error in case the channel is closed.
func (p *parser) run(ctx context.Context) error {
	// pass nil to subscribe to all updates
	linkChan := p.synchronizer.Register(nil)

	for {
		select {
		case links, more := <-linkChan:
			if !more {
				return ErrSyncStopped
			}
			if links != nil {
				err := p.saveLinks(ctx, links)
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
