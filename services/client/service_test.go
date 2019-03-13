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

	"go-connector/services/client"

	jwt "github.com/dgrijalva/jwt-go"
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

	traceServer := createMockServer(t, token, 0)
	accountServer := createMockServer(t, token, 1)

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

	ts := createMockServer(t, token, 1)

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

	ts := createMockServer(t, token, 2)

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

// ============================================================================
// 																	Helpers
// ============================================================================

func createMockServer(t *testing.T, token string, maxLogin int) *httptest.Server {

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

			fmt.Fprintln(w, `{"data": {"value": "42"}}`)
			return
		}
	}))
}
