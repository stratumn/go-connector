package decryption

import (
	"context"
	"encoding/json"
	"errors"

	cs "github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-crypto/encryption"
	"github.com/stratumn/go-crypto/keys"
)

//go:generate mockgen -package mockdecryptor -destination mockdecryptor/mockdecryptor.go github.com/stratumn/go-connector/services/decryption Decryptor

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
	// Decrypt a single link. The decryption is done in place.
	DecryptLink(context.Context, *cs.Link) error
	// DecryptLinks decrypts a list of links. e.g. trace.links.nodes. The decryption is done in place.
	DecryptLinks(context.Context, []*cs.Link) error
	// DecryptLinkData decrypts data given a list of recipients and returns the decrypted data.
	DecryptLinkData(ctx context.Context, data []byte, recipients []*Recipient) ([]byte, error)
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

// Recipient is a decryption recipient.
type Recipient struct {
	PubKey       string
	SymmetricKey []byte
}

type metadata struct {
	Recipients []*Recipient
}

func (d *decryptor) DecryptLinkData(ctx context.Context, data []byte, recipients []*Recipient) ([]byte, error) {
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

func (d *decryptor) DecryptLink(ctx context.Context, l *cs.Link) error {

	if l.GetData() == nil {
		return ErrNoData
	}

	var md metadata
	err := json.Unmarshal(l.GetMeta().GetData(), &md)
	if err != nil {
		return err
	}

	data, err := d.DecryptLinkData(ctx, l.GetData(), md.Recipients)
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
