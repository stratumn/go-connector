package chainscript

import (
	"context"
	"encoding/json"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-crypto/aes"
	"github.com/stratumn/go-crypto/encryption"
)

// PublicKeyInfo contains the public key and its ID.
type PublicKeyInfo struct {
	ID        string
	PublicKey []byte
}

// EncryptLink encrypts the link's data with the provided public keys.
// The link is modified in place.
func EncryptLink(ctx context.Context, link *chainscript.Link, recipientsKeys []*PublicKeyInfo) error {
	data, aesKey, err := aes.Encrypt(link.Data)
	if err != nil {
		return err
	}
	err = link.SetData(data)
	if err != nil {
		return err
	}

	metaData := map[string]interface{}{}
	metadataBytes := link.GetMeta().GetData()
	if len(metadataBytes) != 0 {
		err = json.Unmarshal(metadataBytes, &metaData)
		if err != nil {
			return err
		}
	}

	recipients, err := createRecipientsKeys(ctx, recipientsKeys, aesKey)
	if err != nil {
		return err
	}

	metaData["recipients"] = recipients

	return link.SetMetadata(metaData)
}

// LinkRecipient is the type a recipient should have in the link's metadata.
type LinkRecipient struct {
	PubKeyID     string `json:"pubKeyId"`
	PubKey       string `json:"pubKey"`
	SymmetricKey []byte `json:"symmetricKey"`
}

func createRecipientsKeys(ctx context.Context, recipients []*PublicKeyInfo, key []byte) ([]*LinkRecipient, error) {
	res := make([]*LinkRecipient, len(recipients))

	for i, r := range recipients {
		encryptedKey, err := encryption.EncryptShort(r.PublicKey, key)
		if err != nil {
			return nil, err
		}

		res[i] = &LinkRecipient{
			PubKeyID:     r.ID,
			PubKey:       string(r.PublicKey),
			SymmetricKey: encryptedKey,
		}
	}

	return res, nil
}
