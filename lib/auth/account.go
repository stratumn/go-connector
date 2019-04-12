package auth

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Those are the errors returned by the middleware.
var (
	ErrMissingToken = errors.New("an authorization token must be provided")
)

// Middleware is the interface exposing a middleware function providing authentication.
type Middleware interface {
	WithAuth(next http.HandlerFunc) http.HandlerFunc
}

// StratumnAccountMiddleware implements the Middleware interface.
// It provides authentication using Stratumn Account API.
type StratumnAccountMiddleware struct {
	AccountURL string
}

// NewStratumnAccountMiddleware returns a new instance of StratumnAccountMiddleware.
func NewStratumnAccountMiddleware(accountURL string) (Middleware, error) {
	if _, err := url.ParseRequestURI(accountURL); err != nil {
		return nil, errors.Wrap(err, "could not instantiate Stratumn Account Auth Middleware")
	}
	return &StratumnAccountMiddleware{accountURL}, nil
}

// WithAuth is a middleware function.
// The incoming request must have an 'authorization' header, which is relayed
// to the 'GET /info' route of the Account API.
// The request is rejected if a 401 is returned and goes through otherwise.
func (s *StratumnAccountMiddleware) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		infoReq, err := http.NewRequest("GET", s.AccountURL+"/info", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// forward the authorization token to Account API.
		token := r.Header.Get("authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(ErrMissingToken.Error()))
			return
		}
		infoReq.Header.Set("authorization", token)

		infoResp, err := http.DefaultClient.Do(infoReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if infoResp.StatusCode == http.StatusUnauthorized {
			b, _ := ioutil.ReadAll(infoResp.Body)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(b)
			return
		}

		next.ServeHTTP(w, r)
	}
}
