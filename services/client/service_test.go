package client_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stratumn/go-connector/services/client"
	"github.com/stratumn/go-connector/services/decryption"
	"github.com/stratumn/go-connector/services/decryption/mockdecryptor"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-crypto/signatures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	key = "-----BEGIN ED25519 PRIVATE KEY-----\nMFACAQAwBwYDK2VwBQAEQgRAdWZGknUkmPqtcx3Riy9f99gjCQYIzs3qcxfJ9Z2i\nDSYuwrHWBktWrvBGpaSdmW4kygSRALBlmQgvHmOrJRyC8w==\n-----END ED25519 PRIVATE KEY-----\n"
	q   = "The query"
)

var (
	v = map[string]interface{}{"life": "42"}
)

type testRsp struct{ Value string }

func TestClientService_TraceClient(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 1000,
		IssuedAt:  time.Now().Unix() - 1000,
	}).SignedString([]byte("plap"))

	traceServer := createMockServer(t, token, 0, `{"data": {"value": "42"}}`)
	accountServer := createMockServer(t, token, 1, "")

	defer traceServer.Close()
	defer accountServer.Close()

	config := client.Config{
		TraceURL:          traceServer.URL,
		AccountURL:        accountServer.URL,
		SigningPrivateKey: key,
	}

	s := &client.Service{}
	s.SetConfig(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	c := s.Expose().(client.StratumnClient)
	var rsp testRsp

	// The first call is supposed to log the user in.
	err := c.CallTraceGql(ctx, q, v, &rsp)

	assert.NoError(t, err)
	assert.Equal(t, "42", rsp.Value)

	// We make a second call to check that login is not called twice.
	err = c.CallTraceGql(ctx, q, v, &rsp)
	assert.NoError(t, err)
}

func TestClientService_AccountClient(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 1000,
		IssuedAt:  time.Now().Unix() - 1000,
	}).SignedString([]byte("plap"))

	ts := createMockServer(t, token, 1, `{"data": {"value": "42"}}`)

	defer ts.Close()

	config := client.Config{
		TraceURL:          "fake news",
		AccountURL:        ts.URL,
		SigningPrivateKey: key,
	}

	s := &client.Service{}
	s.SetConfig(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	c := s.Expose().(client.StratumnClient)
	var rsp testRsp

	// The first call is supposed to log the user in.
	err := c.CallAccountGql(ctx, q, v, &rsp)

	assert.NoError(t, err)
	assert.Equal(t, "42", rsp.Value)

	// We make a second call to check that login is not called twice.
	err = c.CallAccountGql(ctx, q, v, &rsp)
	assert.NoError(t, err)
}

func TestClientService_TokenExpired(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() - 1,
		IssuedAt:  time.Now().Unix() - 1000,
	}).SignedString([]byte("plap"))

	ts := createMockServer(t, token, 2, `{"data": {"value": "42"}}`)

	defer ts.Close()

	config := client.Config{
		TraceURL:          ts.URL,
		AccountURL:        ts.URL,
		SigningPrivateKey: key,
	}

	s := &client.Service{}
	s.SetConfig(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	c := s.Expose().(client.StratumnClient)
	var rsp testRsp

	// The first call is supposed to log the user in.
	err := c.CallTraceGql(ctx, q, v, &rsp)

	assert.NoError(t, err)
	assert.Equal(t, "42", rsp.Value)

	// Check that the token is renewed.
	err = c.CallTraceGql(ctx, q, v, &rsp)
	assert.NoError(t, err)
}

func TestClientService_LinkDecryption(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 1000,
		IssuedAt:  time.Now().Unix() - 1000,
	}).SignedString([]byte("plap"))

	linkData := []byte("https://bit.ly/1nab8Fa")
	encLinkData, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9iaXQubHkvMW5hYjhGYQo=")
	recipients := []*decryption.Recipient{&decryption.Recipient{PubKey: "plap", SymmetricKey: []byte("zou")}}
	link := map[string]interface{}{
		"data": encLinkData,
		"meta": map[string]interface{}{"recipients": recipients},
	}

	lb, _ := json.Marshal(link)
	traceServer := createMockServer(t, token, 0, fmt.Sprintf(`{"data": {"link": %s}}`, string(lb)))
	accountServer := createMockServer(t, token, 1, "")

	defer traceServer.Close()
	defer accountServer.Close()

	config := client.Config{
		TraceURL:          traceServer.URL,
		AccountURL:        accountServer.URL,
		SigningPrivateKey: key,
		Decryption:        "decryption",
	}

	s := &client.Service{}
	s.SetConfig(config)

	ctrl := gomock.NewController(t)
	mockDec := mockdecryptor.NewMockDecryptor(ctrl)

	s.Plug(map[string]interface{}{"decryption": mockDec})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	c := s.Expose().(client.StratumnClient)

	t.Run("struct with string data", func(t *testing.T) {
		var rsp struct {
			Link struct {
				Data string
				Meta struct {
					Recipients []*decryption.Recipient
				}
			}
		}

		mockDec.EXPECT().DecryptLinkData(ctx, encLinkData, recipients).Times(1).Return(linkData, nil)

		err := c.CallTraceGql(ctx, q, v, &rsp)
		assert.NoError(t, err)
		assert.Equal(t, string(linkData), rsp.Link.Data)
	})

	t.Run("struct with []byte data", func(t *testing.T) {
		var rsp struct {
			Link struct {
				Data []byte
				Meta struct {
					Recipients []*decryption.Recipient
				}
			}
		}

		mockDec.EXPECT().DecryptLinkData(ctx, encLinkData, recipients).Times(1).Return(linkData, nil)

		err := c.CallTraceGql(ctx, q, v, &rsp)
		assert.NoError(t, err)
		assert.Equal(t, linkData, rsp.Link.Data)
	})

	t.Run("struct with interface data", func(t *testing.T) {
		var rsp struct {
			Link struct {
				Data interface{}
				Meta struct {
					Recipients []*decryption.Recipient
				}
			}
		}

		mockDec.EXPECT().DecryptLinkData(ctx, encLinkData, recipients).Times(1).Return(linkData, nil)

		err := c.CallTraceGql(ctx, q, v, &rsp)
		assert.NoError(t, err)
		assert.Equal(t, linkData, rsp.Link.Data)
	})

	t.Run("interface", func(t *testing.T) {
		// In this case, the response will be unmarshaled into maps.
		var rsp interface{}

		mockDec.EXPECT().DecryptLinkData(ctx, encLinkData, recipients).Times(1).Return(linkData, nil)

		err := c.CallTraceGql(ctx, q, v, &rsp)
		r := rsp.(map[string]interface{})
		l := r["link"].(map[string]interface{})

		assert.NoError(t, err)
		assert.Equal(t, linkData, l["data"])
	})
}

func TestClientService_LinkListDecryption(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 1000,
		IssuedAt:  time.Now().Unix() - 1000,
	}).SignedString([]byte("plap"))

	linkData1 := []byte("https://bit.ly/1nab8Fa")
	encLinkData1, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9iaXQubHkvMW5hYjhGYQo=")
	recipients := []*decryption.Recipient{&decryption.Recipient{PubKey: "plap", SymmetricKey: []byte("zou")}}
	link1 := map[string]interface{}{
		"data": encLinkData1,
		"meta": map[string]interface{}{"recipients": recipients},
	}

	linkData2 := []byte("https://bit.ly/IqT6zt")
	encLinkData2, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9iaXQubHkvSXFUNnp0Cg==")
	link2 := map[string]interface{}{
		"data": encLinkData2,
		"meta": map[string]interface{}{"recipients": recipients},
	}

	lb1, _ := json.Marshal(link1)
	lb2, _ := json.Marshal(link2)
	traceServer := createMockServer(t, token, 0, fmt.Sprintf(`{"data": {"links": [%s, %s]}}`, string(lb1), string(lb2)))
	accountServer := createMockServer(t, token, 1, "")

	defer traceServer.Close()
	defer accountServer.Close()

	config := client.Config{
		TraceURL:          traceServer.URL,
		AccountURL:        accountServer.URL,
		SigningPrivateKey: key,
		Decryption:        "decryption",
	}

	s := &client.Service{}
	s.SetConfig(config)

	ctrl := gomock.NewController(t)
	mockDec := mockdecryptor.NewMockDecryptor(ctrl)

	s.Plug(map[string]interface{}{"decryption": mockDec})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	c := s.Expose().(client.StratumnClient)

	var rsp struct {
		Links []struct {
			Data string
			Meta struct {
				Recipients []*decryption.Recipient
			}
		}
	}

	mockDec.EXPECT().DecryptLinkData(ctx, encLinkData1, recipients).Times(1).Return(linkData1, nil)
	mockDec.EXPECT().DecryptLinkData(ctx, encLinkData2, recipients).Times(1).Return(linkData2, nil)

	err := c.CallTraceGql(ctx, q, v, &rsp)
	assert.NoError(t, err)
	assert.Equal(t, string(linkData1), rsp.Links[0].Data)
	assert.Equal(t, string(linkData2), rsp.Links[1].Data)
}

func TestClientService_NoLinkDecryption(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 1000,
		IssuedAt:  time.Now().Unix() - 1000,
	}).SignedString([]byte("plap"))

	encLinkData, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9iaXQubHkvMW5hYjhGYQo=")
	recipients := []*decryption.Recipient{}
	link := map[string]interface{}{
		"data": encLinkData,
		"meta": map[string]interface{}{"recipients": recipients},
	}

	lb, _ := json.Marshal(link)
	traceServer := createMockServer(t, token, 0, fmt.Sprintf(`{"data": {"link": %s}}`, string(lb)))
	accountServer := createMockServer(t, token, 1, "")

	defer traceServer.Close()
	defer accountServer.Close()

	config := client.Config{
		TraceURL:          traceServer.URL,
		AccountURL:        accountServer.URL,
		SigningPrivateKey: key,
		Decryption:        "decryption",
	}

	s := &client.Service{}
	s.SetConfig(config)

	ctrl := gomock.NewController(t)
	mockDec := mockdecryptor.NewMockDecryptor(ctrl)

	s.Plug(map[string]interface{}{"decryption": mockDec})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runningCh := make(chan struct{})

	go s.Run(ctx, func() { runningCh <- struct{}{} }, func() {})
	<-runningCh

	c := s.Expose().(client.StratumnClient)

	var rsp struct {
		Link struct {
			Data []byte
			Meta struct {
				Recipients []*decryption.Recipient
			}
		}
	}

	mockDec.EXPECT().DecryptLinkData(ctx, encLinkData, recipients).Times(0)

	err := c.CallTraceGql(ctx, q, v, &rsp)
	assert.NoError(t, err)
	assert.Equal(t, encLinkData, rsp.Link.Data)
}

// ============================================================================
// 																	Helpers
// ============================================================================

func createMockServer(t *testing.T, token string, maxLogin int, rsp string) *httptest.Server {

	cntLogin := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/login":
			// Only one login must be done
			cntLogin += 1
			require.True(t, cntLogin < maxLogin+1, fmt.Sprintf("Login should be called only %d times", maxLogin))

			tkn := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
			tb, err := base64.StdEncoding.DecodeString(tkn)
			require.NoError(t, err)

			var sig signatures.Signature
			err = json.Unmarshal(tb, &sig)
			require.NoError(t, err)

			assert.NoError(t, signatures.Verify(&sig))

			// Check that the public key corresponds the our private key.
			_, pk, _ := keys.ParseSecretKey([]byte(key))
			pubKey, _ := keys.EncodePublicKey(pk)
			assert.Equal(t, pubKey, sig.PublicKey)

			fmt.Fprintf(w, `{"token": "%s"}`, token)
			return

		case "/graphql":
			require.Equal(t, fmt.Sprintf("Bearer %s", token), r.Header.Get("authorization"))

			var req map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			assert.Equal(t, q, req["query"])
			assert.Equal(t, v, req["variables"])

			fmt.Fprintln(w, rsp)
			return
		}
	}))
}
