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

// saveLinks stores the links in the bleve store.
// links are indexed by linkHash, data and metadata are deserialized.
// raw contains the non-indexed raw link used to recreate the full link.
func (p *parser) saveSegments(ctx context.Context, segments []*cs.Segment) error {
	b := p.idx.NewBatch()
	for _, s := range segments {
		if err := indexLink(ctx, b, s.Link); err != nil {
			return err
		}
	}

	return p.idx.Batch(b)
}

func indexLink(ctx context.Context, b *bleve.Batch, l *cs.Link) error {
	// Unmarshal link data.
	var data interface{}
	_ = l.StructurizeData(&data)

	// Marshal raw link into bytes.
	lb, err := json.Marshal(l)
	if err != nil {
		return err
	}

	// Unmarshal metadata.
	var md map[string]interface{}
	_ = json.Unmarshal(l.Meta.Data, &md)

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
	segmentsChan := p.synchronizer.Register(nil)

	for {
		select {
		case segments, more := <-segmentsChan:
			if !more {
				return ErrSyncStopped
			}
			if segments != nil {
				err := p.saveSegments(ctx, segments)
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
