package chainscript_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-crypto/encryption"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	csutils "github.com/stratumn/go-connector/lib/chainscript"
)

func decrypt(t *testing.T, l *chainscript.Link, privKey []byte) ([]byte, error) {
	var md struct {
		Recipients []*struct {
			PubKey       string
			SymmetricKey []byte
		}
	}
	err := json.Unmarshal(l.GetMeta().GetData(), &md)

	var encData []byte
	err = json.Unmarshal(l.GetData(), &encData)
	require.NoError(t, err)

	symKey := md.Recipients[0].SymmetricKey

	return encryption.Decrypt(privKey, append(symKey, encData...))
}

func TestNewLink(t *testing.T) {

	pub, _, err := keys.GenerateKey(x509.RSA)
	require.NoError(t, err)

	publicKeys := []*csutils.PublicKeyInfo{
		&csutils.PublicKeyInfo{
			ID:        "1",
			PublicKey: pub,
		},
	}

	data := map[string]interface{}{
		"aw": "yeah",
	}
	meta := map[string]interface{}{
		"monkey": "man",
	}

	t.Run("initialises a map", func(t *testing.T) {
		initLink, err := csutils.NewLink(context.Background(),
			publicKeys,
			"wfID",
			data,
			meta,
			"action",
			"processState",
			[]string{"tag"},
			nil,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, initLink.Meta.MapId)
		assert.Equal(t, "wfID", initLink.Meta.Process.Name)
		assert.Equal(t, "processState", initLink.Meta.Process.State)
		assert.Equal(t, "action", initLink.Meta.Action)
		assert.Equal(t, float64(1), initLink.Meta.Priority)
		assert.EqualValues(t, []string{"tag"}, initLink.Meta.Tags)
		assert.Nil(t, initLink.Meta.Refs)
		var decodedMeta map[string]interface{}
		require.NoError(t, json.Unmarshal(initLink.Meta.GetData(), &decodedMeta))
		assert.Equal(t, "man", decodedMeta["monkey"])
	})

	t.Run("appends to a map", func(t *testing.T) {
		l, err := chainscript.NewLinkBuilder("wfID", "map").Build()
		require.NoError(t, err)
		prevLinkhash, err := l.Hash()
		require.NoError(t, err)

		appendLink, err := csutils.NewLink(context.Background(),
			publicKeys,
			l.Meta.Process.Name,
			data,
			nil,
			"action",
			"processState",
			[]string{"tag"},
			&chainscript.Segment{Link: l, Meta: &chainscript.SegmentMeta{LinkHash: prevLinkhash}},
			&chainscript.LinkReference{
				LinkHash: prevLinkhash,
				Process:  l.Meta.Process.Name,
			},
		)

		require.NoError(t, err)
		assert.Equal(t, l.Meta.MapId, appendLink.Meta.MapId)
		assert.Equal(t, l.Meta.Process.Name, appendLink.Meta.Process.Name)
		assert.Equal(t, l.Meta.Priority+1, appendLink.Meta.Priority)
		assert.Len(t, appendLink.Meta.Refs, 1)
	})

}

func TestEncryptLink(t *testing.T) {
	data := map[string]interface{}{
		"bond": "james",
	}
	dataBytes, _ := json.Marshal(data)

	pub, priv, err := keys.GenerateKey(x509.RSA)
	require.NoError(t, err)

	publicKeys := []*csutils.PublicKeyInfo{
		&csutils.PublicKeyInfo{
			ID:        "1",
			PublicKey: pub,
		},
	}

	t.Run("correctly encrypts the link", func(t *testing.T) {
		metadata := map[string]interface{}{
			"random": "data",
		}
		l, err := chainscript.NewLinkBuilder("process", "map").WithData(data).WithMetadata(metadata).Build()
		require.NoError(t, err)

		// encrypt the link
		err = csutils.EncryptLink(context.Background(), l, publicKeys)
		require.NoError(t, err)
		assert.NotEqual(t, data, l.Data)

		// decrypt it
		res, err := decrypt(t, l, priv)
		require.NoError(t, err)
		require.Equal(t, dataBytes, res)
	})

	t.Run("correctly encrypts the link when metadata is empty", func(t *testing.T) {
		l, err := chainscript.NewLinkBuilder("process", "map").WithData(data).Build()
		require.NoError(t, err)

		// encrypt the link
		err = csutils.EncryptLink(context.Background(), l, publicKeys)
		require.NoError(t, err)
		assert.NotEqual(t, data, l.Data)

		// decrypt it
		res, err := decrypt(t, l, priv)
		require.NoError(t, err)
		require.Equal(t, dataBytes, res)
	})

	t.Run("fails when recipients are empty", func(t *testing.T) {
		l, err := chainscript.NewLinkBuilder("process", "map").WithData(data).Build()
		require.NoError(t, err)

		err = csutils.EncryptLink(context.Background(), l, []*csutils.PublicKeyInfo{})
		require.EqualError(t, err, csutils.ErrMissingRecipients.Error())
	})

}
