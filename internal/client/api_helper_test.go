package client_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"unicode/utf8"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

/**
Here we define a mock http server which will pretend to be an MKE API

It will handle authentication responses, and any additional routes that you pass it.
Routes are added as MockHandlers matching URLs.

Authentication is handled as a passed in Auth struct. The .Username and .Password are
verified and then the .Token is returned. If not authentication is to be done then
a nil Auth should be provided, and an error will occur if authentication is attempted.

*/

type MockHandlerMap map[MockHandlerKey]MockHandler

type MockHandlerKey struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// a function that can handler an API HTTP request to mock it
type MockHandler func(w http.ResponseWriter, r *http.Request)

// Generate a test API server which will be usable for testing API Calls
// if auth passed is nil, then no authentication occurs, otherwise U/P are expected for auth, and Token is returned
func MockTestServer(auth *client.Auth, handlers MockHandlerMap) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// strip the leading slash from the path
		path := r.URL.Path
		_, i := utf8.DecodeRuneInString(path)
		path = path[i:]
		reqMockHandlerKey := MockHandlerKey{
			Method: r.Method,
			Path:   path,
		}

		// if the request is for the authentication path, then handle it.
		if path == client.URLTargetForAuth {
			if auth != nil {
				reqBodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusNotAcceptable)
				}

				var reqAuth client.Auth

				if err := json.Unmarshal(reqBodyBytes, &reqAuth); err != nil {
					w.WriteHeader(http.StatusNotAcceptable)
				}

				if auth.Username != reqAuth.Username || auth.Password != reqAuth.Password {
					w.WriteHeader(http.StatusUnauthorized)
				}

				lr := client.NewLoginResponse(auth.Token)

				w.Write(lr.Bytes()) //nolint:errcheck
			} else {
				w.WriteHeader(http.StatusNotAcceptable)
			}

			return
		}

		if handler, ok := handlers[reqMockHandlerKey]; !ok {
			w.WriteHeader(http.StatusNotFound)
			b, _ := json.Marshal(reqMockHandlerKey)
			w.Write(b)                        //nolint:errcheck
			w.Write([]byte(" NOT FOUND IN ")) //nolint:errcheck
			for k := range handlers {
				b, _ := json.Marshal(k)
				w.Write(b) //nolint:errcheck
			}
		} else {
			authHeader := r.Header.Get(client.HeaderKeyAuthorization)
			if auth != nil {
				if expectedAuthHeader := client.BearerTokenHeaderValue(auth.Token); authHeader != expectedAuthHeader {
					w.WriteHeader(http.StatusUnauthorized)
				}
			} else {
				if authHeader != "" {
					w.WriteHeader(http.StatusNotAcceptable)
				}
			}

			handler(w, r)
		}

	}))
}

// MockServerHandlerGeneratorReturnResponseStatus generates a MockHandler which just sets http status.
func MockServerHandlerGeneratorReturnResponseStatus(status int) MockHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

// MockServerHandlerGeneratorReturnBytes generates a MockHandler which just returns bytes.
func MockServerHandlerGeneratorReturnBytes(expected []byte) MockHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(expected) //nolint:errcheck
	}
}

// MockServerHandlerGeneratorReturnJson generatres a MockHandler which returns a JSON serialized argument.
func MockServerHandlerGeneratorReturnJson(expected interface{}) MockHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		expectedBytes, _ := json.Marshal(expected)
		w.Write(expectedBytes) //nolint:errcheck
	}
}
