package parser

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/db"

	cs "github.com/stratumn/go-chainscript"

	"github.com/stratumn/go-connector/services/livesync"
)

var (
	// ErrSyncStopped is returned when the subscription channel is closed by the synchronizer service.
	ErrSyncStopped = errors.New("synchronizer service stopped")
)

type parser struct {
	db           db.DB
	synchronizer livesync.Synchronizer
}

// saveLinks stores the links in the key/value store.
// links are indexed by linkHash and are serialized to JSON.
func (p *parser) saveLinks(links []*cs.Link) error {
	for _, link := range links {

		lh, err := link.Hash()
		if err != nil {
			return err
		}

		linkBytes, err := json.Marshal(link)
		if err != nil {
			return err
		}
		err = p.db.Put(lh, linkBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

// run subscribes to the livesync service and waits for updates.
// It returns an error in case the channel is closed.
func (p *parser) run(ctx context.Context) error {
	linkChan := p.synchronizer.Register()

	for {
		select {
		case links, more := <-linkChan:
			if !more {
				return ErrSyncStopped
			}
			if links != nil {
				err := p.saveLinks(links)
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
