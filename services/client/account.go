package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-crypto/signatures"
)

// AccountClient defines all the possible interactions with Account.
type AccountClient interface {
	// CallAccountGql makes a call to the Account graphql endpoint.
	CallAccountGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error
}

func (c *client) CallAccountGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
	return c.callGqlEndpoint(ctx, c.urlAccount+"/graphql", query, variables, rsp)
}

type tokenBody struct {
	Iat int64 `json:"iat"`
	Exp int64 `json:"exp"`
}

func (c *client) login(ctx context.Context) (string, error) {
	log.Info("Login")

	tb := tokenBody{
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(time.Minute * 5).Unix(),
	}

	b, err := json.Marshal(tb)
	if err != nil {
		return "", err
	}

	sig, err := signatures.Sign(c.signingPrivateKey, b)
	if err != nil {
		return "", err
	}

	token, err := json.Marshal(sig)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, c.urlAccount+"/login", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString(token)))
	r, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	if r.StatusCode != http.StatusOK {
		return "", errors.Errorf("HTTP error %d", r.StatusCode)
	}

	var rsp struct{ Token string }

	err = json.NewDecoder(r.Body).Decode(&rsp)
	if err != nil {
		return "", err
	}

	return rsp.Token, nil
}

func (c *client) checkAndRenewToken(ctx context.Context) error {
	// Check if the token is still valid
	if c.authToken != "" {
		p := jwt.Parser{}
		cl := &jwt.StandardClaims{}
		_, _, err := p.ParseUnverified(c.authToken, cl)
		if err != nil {
			return err
		}

		if cl.ExpiresAt > time.Now().Unix()+1 {
			// The token is still valid.
			return nil
		}
	}

	t, err := c.login(ctx)
	if err != nil {
		return err
	}
	c.authToken = t
	return nil
}
