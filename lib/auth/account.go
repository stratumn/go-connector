package auth

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
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
	if accountURL == "" {
		return nil, errors.New("Stratumn Account API URL is required")
	}
	return &StratumnAccountMiddleware{accountURL}, nil
}

// WithAuth is a middleware function.
// The incoming request must have an 'authorization' header, which is relayed
// to the 'GET /info' route of the Account API.
// The request is rejected if a 401 is returned and goes through otherwise.
func (s *StratumnAccountMiddleware) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Logged connection from %s", r.RemoteAddr)
		infoReq, err := http.NewRequest("GET", s.AccountURL+"/info", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// forward the authorization token to Account API.
		token := r.Header.Get("authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("an authorization tokne must be provided"))
			return
		}
		infoReq.Header.Set("authorization", token)

		infoResp, err := http.DefaultClient.Do(infoReq)
		if err != nil {
			b, _ := ioutil.ReadAll(infoResp.Body)
			w.WriteHeader(infoResp.StatusCode)
			w.Write(b)
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
