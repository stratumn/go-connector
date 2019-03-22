package bleveparser

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/blevesearch/bleve"
	cs "github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-connector/services/livesync"
)

var (
	// ErrSyncStopped is returned when the subscription channel is closed by the synchronizer service.
	ErrSyncStopped = errors.New("synchronizer service stopped")
)

type parser struct {
	idx          bleve.Index
	synchronizer livesync.Synchronizer
}

// saveLinks stores the links in the key/value store.
// links are indexed by linkHash and are serialized to JSON.
func (p *parser) saveLinks(ctx context.Context, links []*cs.Link) error {
	b := p.idx.NewBatch()
	for _, l := range links {
		if err := indexLink(ctx, b, l); err != nil {
			return err
		}
	}

	return p.idx.Batch(b)
}

func indexLink(ctx context.Context, b *bleve.Batch, l *cs.Link) error {
	// Unmarshal link data.
	var data interface{}
	err := l.StructurizeData(&data)
	if err != nil {
		return err
	}

	// Marshal raw link into bytes.
	lb, err := json.Marshal(l)
	if err != nil {
		return err
	}

	// Unmarshal metadata.
	var md map[string]interface{}
	err = json.Unmarshal(l.Meta.Data, &md)
	if err != nil {
		return err
	}

	// Use link hash as document key.
	lh, err := l.Hash()
	if err != nil {
		return err
	}

	// Unmarshal link meta into a map. This converts the []byte into base64 strings.
	var lm map[string]interface{}
	lmb, err := json.Marshal(l.Meta)
	if err != nil {
		return err
	}
	err = json.Unmarshal(lmb, &lm)
	if err != nil {
		return err
	}

	return b.Index(
		lh.String(),
		map[string]interface{}{
			"type":     "root",
			"raw":      string(lb),
			"meta":     lm,
			"metadata": md,
			"data":     data,
		},
	)
}

// run subscribes to the livesync service and waits for updates.
// It returns an error in case the channel is closed.
func (p *parser) run(ctx context.Context) error {
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
			return nil
		}
	}
}
