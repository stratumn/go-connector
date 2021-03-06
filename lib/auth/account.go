package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Those are the errors returned by the middleware.
var (
	ErrMissingToken = errors.New("an authorization token must be provided")

	ErrUnauthorizedAccount = errors.New("user is not part of the authorized entities")
)

// Middleware is the interface exposing a middleware function providing authentication.
type Middleware interface {
	WithAuth(next http.HandlerFunc) http.HandlerFunc
}

// StratumnAccountMiddleware implements the Middleware interface.
// It provides authentication using Stratumn Account API.
type StratumnAccountMiddleware struct {
	AccountURL         string
	AuthorizedAccounts []string
}

// AccountInfo contains the data returned by StratumnAccount `GET /info` call.
type AccountInfo struct {
	AccountID       string   `json:"accountId"`
	OtherAccountIDs []string `json:"otherAccountIds"`
}

// NewStratumnAccountMiddleware returns a new instance of StratumnAccountMiddleware.
func NewStratumnAccountMiddleware(accountURL string, authorizedAccounts []string) (Middleware, error) {
	if _, err := url.ParseRequestURI(accountURL); err != nil {
		return nil, errors.Wrap(err, "could not instantiate Stratumn Account Auth Middleware")
	}
	return &StratumnAccountMiddleware{accountURL, authorizedAccounts}, nil
}

// WithAuth is a middleware function.
// The incoming request must have an 'authorization' header, which is relayed
// to the 'GET /info' route of the Account API.
// The request is rejected if a 401 is returned and goes through otherwise.
func (s *StratumnAccountMiddleware) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		infoReq, err := http.NewRequest("GET", s.AccountURL+"/info", nil)
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		// forward the authorization token to Account API.
		token := r.Header.Get("authorization")
		if token == "" {
			writeResponse(w, http.StatusUnauthorized, []byte(ErrMissingToken.Error()))
			return
		}
		infoReq.Header.Set("authorization", token)

		infoResp, err := http.DefaultClient.Do(infoReq)
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
			return
		}
		if infoResp.StatusCode >= 400 {
			b, _ := ioutil.ReadAll(infoResp.Body)
			writeResponse(w, http.StatusUnauthorized, b)
			return
		}

		info := AccountInfo{}
		err = json.NewDecoder(infoResp.Body).Decode(&info)
		if err != nil {
			writeResponse(w, http.StatusUnauthorized, []byte(err.Error()))
			return
		}

		err = s.checkAuth(&info)
		if err != nil {
			writeResponse(w, http.StatusUnauthorized, []byte(err.Error()))
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *StratumnAccountMiddleware) checkAuth(info *AccountInfo) error {
	if len(s.AuthorizedAccounts) == 0 {
		return nil
	}

	for _, authorized := range s.AuthorizedAccounts {
		if info.AccountID == authorized {
			return nil
		}
		for _, entityID := range info.OtherAccountIDs {
			if entityID == authorized {
				return nil
			}
		}
	}

	return ErrUnauthorizedAccount
}

func writeResponse(w http.ResponseWriter, statusCode int, response []byte) {
	w.WriteHeader(statusCode)
	_, _ = w.Write(response)
}
