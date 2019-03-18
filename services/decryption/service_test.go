package decryption_test

import (
	"context"
	"encoding/json"
	"testing"

	"go-connector/services/decryption"

	cs "github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-crypto/aes"
	"github.com/stratumn/go-crypto/encryption"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	key     = "-----BEGIN RSA PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC9EesXKgTov82E\nUE+R/bn+MFhLhSAF0qX6le70E5lq/sYWsklBo+7MmfJOA+a3WMuWqMeLGkzeIO5q\nBLHIQB/ttKYG3Oto9ebsJ3/fuBSDtmYJftiGH0lGu/etLRXmzn0fEP5kG0Wk+aPt\nTFHWDb7hfsAyBZtB+2j9w0rb2Vn2GNV/SySevWw+YEFrXd4H6FSwqcCdsfcOw1Vp\nmxcNq7MoxX+QzrHwesvEjH+7I4mTu52Bzxo4rerZ1fgf4JyGY4ZmYlpO3bazZkNF\n44SlJNtgATrkcJPFcjikvdf7R3kldbO8aO6oI7Iv5pcSuwUO2JHnl3zYWJ5Mj3yG\nV9FoyWtnAgMBAAECggEACLB+IX5o41mNVHtsbMVAexI1vKLNqfbYcf/aD5WnA2fa\nKsje3QlvvC+HF1bOj7ahBLeVFCuNRNg2nODCWvW3gfE/gCk/GH/UsR3PnrSTHMqR\nPfZ6dQ+TCpEw/OCJlSYAUiExz+AA/2gJxKoxSKkxEKQGqUXgsNOhK6iCFECVBd5h\n4emuRh4cIEm4HqSmCRLy1gWKocuC9ynuu0UHCl4XkIPwHxrmwABy4BmAXbPRNJh7\n98cNX0t+yG16eblacjZ8vSevvym4GNH4ajl5nyoTesRS54v0JMTWErwGgfPARluC\nr6l6nydqmttHVvc/zMBc3VSuSYOzufOtlhoiZ97XgQKBgQD+YUz1z6VaI4q0oIQW\nem88AHZI0uJ3ntPz0jSHpcJEK/OMDdDK5ODRoy9s6JnuVXN8LqfFzNst/9nxFl+2\nM34cOKODF4yicOkyGaocfU0LGBxWHn8aCJZo8N84BA/sWerueWX/Hpb3EGrogYj3\nnPRxZ+HOsKZ7tCSO4xFl42ow0QKBgQC+RiWamQFxF/UiAoqCDTgrTeqodv6wgYmR\nbOynQWYKIkJtpLtzNSV2LIz4X0VdVUGfbsZW8QTgjiXv/JldAuxNpr1ndm1cHkaQ\n/Heik6PptQrUICIeOTSkj/56D4DLgs5WKSjRkNsdhaexGJ/4Bskhtv53kp5Pq9jH\naNLJLummtwKBgQCSRfcQHg/R7kATL33kwxB1azqZE5KgAFeWi5gjLCCyPKe2MDeQ\ng933DiP2NyZUkxRuIxHcPrkGEWoMJLZyuddZeQQlHISE3/JoGbPk3/ROXdXle3HQ\n0YFT5LYmqsdRPD9IU8xf0AI1HV6sRdgxsjIph/ejd5az6VlgRJe7g/KLEQKBgE5K\nYrKn/lXge7bQwNkeQ1xeJQ3IWKebxVUXMpDncer9icO/onmXBqEHV8HiwZHTwLqv\nQ+EGLvGOy8FheGEzELQqxYhKzFi5BGQn3boBcdJ58ciyqBczhpunvBfRRTd3zRra\nuLbyGZaeJg/SiA/wCtZai3370DQMC5iRYxnwuaclAoGBAL9b4vqZ7u5biF2aMXEp\n0Wpa/bYlEzwK2jbCNsZEi2d1tIebPSIMrUVlMuJLYLlk7ipSy5kVi9cm1Ir7Dmwc\n5xB5C4pmqYEpWunyZhZF8O70x7BF7Fwhjv9W6AIy2fvVmJJTUUb7cCtKRgwLMA+R\n5F6I2rsybDkznOInX67dO/Ui\n-----END RSA PRIVATE KEY-----\n"
	otherPk = "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnEHluaVmVFDzc2K47ntl\n9khvzX567pSgCZsOy4iuUSuQ1mGVRUFkcaUz2/xIZSbJHjwpi1lGmJItp92v0cZo\nUEn0ln0nI6UNRK3+MhA0ZyYFb8xs0UCe1OafEHVkuApGS0GVaraRp1LNLZYGPOQF\nHKkuA5b4l9imEJ5bxJIRJJQTIj10+RB4UFFj7WvsEd6oXp+3iS8SKumDF+sMQDPf\n8r+umOFFhm4f4nxSPP6qh85awqfVSVBM4lyXVf+xmhpSp50F18GGdGg8jiCtR7tC\nEcFQH/xUWx+VO1O3NJqLe0wIYneZTExfjEAVhs5yVXe6oyLSmtCfZcxurVQi26Xq\nMQIDAQAB\n-----END RSA PUBLIC KEY-----\n"
)

var (
	_, pub, _ = keys.ParseSecretKey([]byte(key))
	pk, _     = keys.EncodePublicKey(pub)
)

func TestDecryptionService_DecryptLink(t *testing.T) {
	config := decryption.Config{
		EncryptionPrivateKey: key,
	}

	s := &decryption.Service{}
	s.SetConfig(config)

	ctx := context.Background()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	d := s.Expose().(decryption.Decryptor)

	// Create link data and encrypt it.
	data := map[string]interface{}{
		"life":  "42",
		"pizza": "yolo",
	}

	l := createEncryptedLink(t, data, [][]byte{pk, []byte(otherPk)})
	err := d.DecryptLink(ctx, l)
	assert.NoError(t, err)

	var decrypted interface{}
	err = l.StructurizeData(&decrypted)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDecryptionService_DecryptLinks(t *testing.T) {
	config := decryption.Config{
		EncryptionPrivateKey: key,
	}

	s := &decryption.Service{}
	s.SetConfig(config)

	ctx := context.Background()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	d := s.Expose().(decryption.Decryptor)

	data1 := map[string]interface{}{"life": "42", "pizza": "yolo"}
	l1 := createEncryptedLink(t, data1, [][]byte{pk, []byte(otherPk)})
	data2 := map[string]interface{}{"plap": "zou"}
	l2 := createEncryptedLink(t, data2, [][]byte{pk, []byte(otherPk)})

	err := d.DecryptLinks(ctx, []*cs.Link{l1, l2})
	assert.NoError(t, err)

	var decrypted1 interface{}
	err = l1.StructurizeData(&decrypted1)
	assert.NoError(t, err)
	assert.Equal(t, data1, decrypted1)

	var decrypted2 interface{}
	err = l2.StructurizeData(&decrypted2)
	assert.NoError(t, err)
	assert.Equal(t, data2, decrypted2)
}

func TestDecryptionService_NotARecipient(t *testing.T) {
	config := decryption.Config{
		EncryptionPrivateKey: key,
	}

	s := &decryption.Service{}
	s.SetConfig(config)

	ctx := context.Background()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	d := s.Expose().(decryption.Decryptor)

	// Create link data and encrypt it.
	data := map[string]interface{}{
		"life": 42,
		"plap": map[string]string{"pizza": "yolo"},
	}

	l := createEncryptedLink(t, data, [][]byte{[]byte(otherPk)})
	err := d.DecryptLink(ctx, l)
	assert.EqualError(t, err, decryption.ErrNotInRecipients.Error())
}

func TestDecryptionService_NotData(t *testing.T) {
	config := decryption.Config{
		EncryptionPrivateKey: key,
	}

	s := &decryption.Service{}
	s.SetConfig(config)

	ctx := context.Background()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	d := s.Expose().(decryption.Decryptor)

	l := createEncryptedLink(t, nil, [][]byte{[]byte(otherPk)})
	l.Data = nil
	err := d.DecryptLink(ctx, l)
	assert.EqualError(t, err, decryption.ErrNoData.Error())
}

// ============================================================================
// 																	Helpers
// ============================================================================

func createEncryptedLink(t *testing.T, data map[string]interface{}, pks [][]byte) *cs.Link {
	dataBytes, _ := json.Marshal(data)
	encData, aesKey, err := aes.Encrypt(dataBytes)

	// Format the metadata.
	md := map[string]interface{}{
		"recipients": createRecipients(t, pks, aesKey),
	}

	l, err := cs.NewLinkBuilder("p", "m").WithMetadata(md).Build()
	l.Data = encData
	require.NoError(t, err)

	return l
}

func createRecipients(t *testing.T, pks [][]byte, key []byte) []*decryption.Recipient {
	res := make([]*decryption.Recipient, len(pks))

	for i, pk := range pks {
		encryptedKey, err := encryption.EncryptShort(pk, key)
		require.NoError(t, err)

		res[i] = &decryption.Recipient{
			PubKey:       string(pk),
			SymmetricKey: encryptedKey,
		}
	}

	return res
}
