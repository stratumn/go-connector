package decryption

import (
	"context"
	"encoding/json"
	"errors"

	cs "github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-crypto/encryption"
	"github.com/stratumn/go-crypto/keys"
)

var (
	// ErrNotInRecipients is returned when the connector is not one
	// of the recipients of the link to decrypt.
	ErrNotInRecipients = errors.New("the link was not encrypted for us")

	// ErrNoData is returned when the link contains no data.
	ErrNoData = errors.New("the link contains no data")
)

// Decryptor decrypt links using the connector's key.
// Each link mist contain data and meta.recipients.
type Decryptor interface {
	// Decrypt a single link.
	DecryptLink(context.Context, *cs.Link) error
	// Decrypt a list of links. e.g. trace.links.nodes
	DecryptLinks(context.Context, []*cs.Link) error
}

type decryptor struct {
	encryptionPrivateKey []byte
	encryptionPublicKey  []byte
}

func newDecryptor(sk []byte) (Decryptor, error) {
	// Get the public key.
	_, pk, err := keys.ParseSecretKey(sk)
	if err != nil {
		return nil, err
	}
	pkBytes, err := keys.EncodePublicKey(pk)
	if err != nil {
		return nil, err
	}

	return &decryptor{
		encryptionPrivateKey: sk,
		encryptionPublicKey:  pkBytes,
	}, nil
}

type recipient struct {
	PubKey       string
	SymmetricKey []byte
}

type metadata struct {
	Recipients []*recipient
}

func (d *decryptor) DecryptLink(ctx context.Context, l *cs.Link) error {

	if l.GetData() == nil {
		return ErrNoData
	}

	var md metadata
	err := json.Unmarshal(l.GetMeta().GetData(), &md)
	if err != nil {
		return err
	}

	data, err := d.decryptLinkData(ctx, l.GetData(), md.Recipients)
	if err != nil {
		return err
	}

	l.Data = data
	return nil
}

func (d *decryptor) DecryptLinks(ctx context.Context, links []*cs.Link) error {
	for _, l := range links {
		if err := d.DecryptLink(ctx, l); err != nil {
			return err
		}
	}
	return nil
}

func (d *decryptor) decryptLinkData(ctx context.Context, data []byte, recipients []*recipient) ([]byte, error) {
	// Get the symmetric key that was RSA-encrypted for us.
	var symKey []byte
	for i := range recipients {
		if recipients[i].PubKey == string(d.encryptionPublicKey) {
			symKey = recipients[i].SymmetricKey
			break
		}
	}

	if symKey == nil {
		return nil, ErrNotInRecipients
	}

	return encryption.Decrypt(d.encryptionPrivateKey, append(symKey, data...))
}
