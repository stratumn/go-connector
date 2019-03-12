package decryption

import (
	"context"

	cs "github.com/stratumn/go-chainscript"
)

// Decryptor decrypt links using the connector's key.
// Each link mist contain data and meta.recipients.
type Decryptor interface {
	// Decrypt a single link.
	DecryptLink(context.Context, *cs.Link) error
	// Decrypt a list of links. e.g. trace.links.nodes
	DecryptTrace(context.Context, []*cs.Link) error
}

type decryptor struct {
	encryptionPrivateKey []byte
}

func newDecryptor(k []byte) Decryptor {
	return &decryptor{
		encryptionPrivateKey: k,
	}
}

func (d *decryptor) DecryptLink(ctx context.Context, l *cs.Link) error {
	return nil
}

func (d *decryptor) DecryptTrace(ctx context.Context, l []*cs.Link) error {
	return nil
}
