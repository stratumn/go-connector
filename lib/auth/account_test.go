package auth_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-connector/lib/auth"
)

const (
	validToken = "Bearer super-secret-token"

	accountAPIError    = "fail"
	accountAPIResponse = `{"email":"hello@stratumn.com","accountId":"1","otherAccountIds":["1","2"],"userId":"3"}`

	apiResponse = "ok"
)

func mockStratumnAccount() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/info", func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("authorization") != validToken {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(accountAPIError))
			return
		}
		fmt.Fprint(w, accountAPIResponse)
	})
	ts := httptest.NewServer(mux)
	return ts
}

func mockAPI(middleware func(http.HandlerFunc) http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/any", middleware(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, apiResponse)
	}))
	ts := httptest.NewServer(mux)
	return ts
}

func TestStratumnAccountMiddleware(t *testing.T) {

	t.Run("Fails if the account URL is not specified", func(t *testing.T) {
		_, err := auth.NewStratumnAccountMiddleware("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not instantiate Stratumn Account Auth Middleware")
	})

	t.Run("Fails if the request does not have an auth token", func(t *testing.T) {
		accountMock := mockStratumnAccount()
		defer accountMock.Close()

		m, err := auth.NewStratumnAccountMiddleware(accountMock.URL)
		require.NoError(t, err)

		apiMock := mockAPI(m.WithAuth)
		defer apiMock.Close()

		rsp, err := http.Get(apiMock.URL + "/any")
		require.NoError(t, err)

		b, _ := ioutil.ReadAll(rsp.Body)
		assert.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
		assert.Equal(t, []byte(auth.ErrMissingToken.Error()), b)
	})

	t.Run("Fails if account API returns a 401", func(t *testing.T) {
		accountMock := mockStratumnAccount()
		defer accountMock.Close()

		m, err := auth.NewStratumnAccountMiddleware(accountMock.URL)
		require.NoError(t, err)

		apiMock := mockAPI(m.WithAuth)
		defer apiMock.Close()

		req, _ := http.NewRequest("GET", apiMock.URL+"/any", nil)
		req.Header.Set("authorization", "Bearer bad token")
		rsp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		b, _ := ioutil.ReadAll(rsp.Body)
		assert.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
		assert.Equal(t, []byte(accountAPIError), b)
	})

	t.Run("Serves the next request", func(t *testing.T) {
		accountMock := mockStratumnAccount()
		defer accountMock.Close()

		m, err := auth.NewStratumnAccountMiddleware(accountMock.URL)
		require.NoError(t, err)

		apiMock := mockAPI(m.WithAuth)
		defer apiMock.Close()

		req, _ := http.NewRequest("GET", apiMock.URL+"/any", nil)
		req.Header.Set("authorization", validToken)
		rsp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		b, _ := ioutil.ReadAll(rsp.Body)
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
		assert.Equal(t, []byte(apiResponse), b)
	})
}
