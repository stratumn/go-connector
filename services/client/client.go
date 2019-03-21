package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stratumn/go-connector/services/decryption"

	logging "github.com/ipfs/go-log"
	"github.com/pkg/errors"
)

//go:generate mockgen -package mockclient -destination mockclient/mockclient.go go-connector/services/client StratumnClient

// log is the logger for the protocol.
var log = logging.Logger("client")

// StratumnClient is the client interface to Stratumn services.
type StratumnClient interface {
	TraceClient
	AccountClient
}

type client struct {
	urlTrace   string
	urlAccount string
	httpClient *http.Client
	decryptor  decryption.Decryptor

	// The PEM encoded signing private key of the conenctor.
	signingPrivateKey []byte

	authToken string
}

func newClient(traceURL string, accountURL string, signingPrivateKey []byte, decryptor decryption.Decryptor) StratumnClient {
	httpClient := &http.Client{Timeout: time.Second * 10}

	return &client{
		urlTrace:          traceURL,
		urlAccount:        accountURL,
		httpClient:        httpClient,
		decryptor:         decryptor,
		signingPrivateKey: signingPrivateKey,
	}
}

type gqlError struct {
	Message string
	Status  int
}

type gqlResponse struct {
	Data   interface{}
	Errors []gqlError
}

// Helper that calls the graphql endpoint and renews the token when necessary.
func (c *client) callGqlEndpoint(ctx context.Context, url string, query string, variables map[string]interface{}, rsp interface{}) error {
	if err := c.checkAndRenewToken(ctx); err != nil {
		return err
	}

	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", c.authToken))

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	gqlRsp := gqlResponse{
		Data: rsp,
	}

	err = json.NewDecoder(r.Body).Decode(&gqlRsp)
	if err != nil {
		return err
	}

	if len(gqlRsp.Errors) > 0 {
		// return the first error
		return errors.Errorf("graphql (%d): %s", gqlRsp.Errors[0].Status, gqlRsp.Errors[0].Message)
	}

	return nil
}
