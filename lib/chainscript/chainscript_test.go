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

func TestEncryptLink(t *testing.T) {
	data := map[string]interface{}{
		"bond": "james",
	}
	dataBytes, _ := json.Marshal(data)

	pub, priv, err := keys.GenerateKey(x509.RSA)
	require.NoError(t, err)

	t.Run("correctly encrypts the link", func(t *testing.T) {
		metadata := map[string]interface{}{
			"random": "data",
		}
		l, err := chainscript.NewLinkBuilder("process", "map").WithData(data).WithMetadata(metadata).Build()
		require.NoError(t, err)

		publicKeys := []*csutils.PublicKeyInfo{
			&csutils.PublicKeyInfo{
				ID:        "1",
				PublicKey: pub,
			},
		}

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

		publicKeys := []*csutils.PublicKeyInfo{
			&csutils.PublicKeyInfo{
				ID:        "1",
				PublicKey: pub,
			},
		}

		// encrypt the link
		err = csutils.EncryptLink(context.Background(), l, publicKeys)
		require.NoError(t, err)
		assert.NotEqual(t, data, l.Data)

		// decrypt it
		res, err := decrypt(t, l, priv)
		require.NoError(t, err)
		require.Equal(t, dataBytes, res)
	})
}
