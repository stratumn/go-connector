package chainscript

import (
	"context"
	"encoding/json"
	"errors"

	uuid "github.com/satori/go.uuid"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-crypto/aes"
	"github.com/stratumn/go-crypto/encryption"
)

var (
	// ErrMissingRecipients is the error returned when no recipients public keys were provided.
	ErrMissingRecipients = errors.New("no recipients were provided")
)

// NewLink returns a new chainscript link ready to be sent by the trace client.
// Note that it encrypts the form data, but does not sign the link since this is done in the client.
func NewLink(
	ctx context.Context,
	recipients []*PublicKeyInfo,
	wfID string,
	formData interface{},
	metadata interface{},
	action string,
	processState string,
	tags []string,
	prevSegment *chainscript.Segment,
	refs ...*chainscript.LinkReference) (*chainscript.Link, error) {

	var mapID string
	var priority float64
	if prevSegment != nil {
		prevLink := prevSegment.Link
		mapID = prevLink.Meta.MapId
		priority = prevLink.Meta.Priority + 1
	} else {
		mapID = uuid.NewV4().String()
		priority = 1
	}

	linkBuilder := chainscript.NewLinkBuilder(wfID, mapID).
		WithData(formData).
		WithAction(action).
		WithProcessState(processState).
		WithDegree(1).
		WithPriority(priority)
	if metadata != nil {
		linkBuilder = linkBuilder.WithMetadata(metadata)
	}
	if prevSegment != nil {
		prevLinkHash, _ := prevSegment.Link.Hash()
		if prevSegment.Meta != nil {
			prevLinkHash = prevSegment.Meta.LinkHash
		}
		linkBuilder = linkBuilder.WithParent(prevLinkHash)
	}
	if len(refs) > 0 {
		linkBuilder = linkBuilder.WithRefs(refs...)
	}
	if len(tags) > 0 {
		linkBuilder = linkBuilder.WithTags(tags...)
	}
	link, err := linkBuilder.Build()
	if err != nil {
		return nil, err
	}

	err = EncryptLink(ctx, link, recipients)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// PublicKeyInfo contains the public key and its ID.
type PublicKeyInfo struct {
	ID        string
	PublicKey []byte
}

// EncryptLink encrypts the link's data with the provided public keys.
// The link is modified in place.
func EncryptLink(ctx context.Context, link *chainscript.Link, recipientsKeys []*PublicKeyInfo) error {
	if len(recipientsKeys) == 0 {
		return ErrMissingRecipients
	}

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
