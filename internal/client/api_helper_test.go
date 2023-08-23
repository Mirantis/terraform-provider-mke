package client_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"unicode/utf8"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

/**
Here we define a mock http server which will pretend to be an MKE API

It will handle authentication responses, and any additional routes that you pass it.
Routes are added as MockHandlers matching URL path/method pairs.

Authentication is handled as a passed in Auth struct. The .Username and .Password are
verified and then the .Token is returned. If not authentication is to be done then
a nil Auth should be provided, and an error will occur if authentication is attempted.

  e.g.

  ```
    auth := client.Auth // commonTestAuth is available for this

	s := NewMockTestServer(&auth) // leave auth nil for unauthenticated urls
    s.AddHandler(http.MethodPost, "my/post/url/" + auth.Username, MockServerHandlerGeneratorReturnJson(... some struct here ...))
	s.AddHandler(http.MethodDelete, "my/del/url", MockServerHandlerGeneratorReturnResponseStatus(http.StatusOK))

    s.AddHandler(http.MethodGet, "my/home", func(w http.ResponseWriter, r *http.Request) {
        // do some custom stuff
        w.WriteResponse(http.StatusOK)
    }

	c, err := s.Client() // get the server to make you the api client
	if err != nil { // this can usually be ignored
		t.Fatalf("Could not make a client: %s", err)
	}

    // call client methods here that rely on the above routes
    // confirm that you receive the expected client responses/errors related to the set URLs
  ```
*/

var (
	commonTestAuth client.Auth = client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
)

// MockTestHandler a handler function with to be used for the path/method provided.
type MockTestHandler struct {
	path    string
	method  string
	handler http.HandlerFunc
}

// Match target matcher based on passed path/method
// @TODO allow wildcarding.
func (mth MockTestHandler) Match(method string, path string) bool {
	return method == mth.method && path == mth.path
}

// MockTestServer httptest.Server wrapper with easier to register handlers and client generation.
//
//	Create a server, register handlers, and ask it for an http Client for your MKE client.
type MockTestServer struct {
	handlers   []MockTestHandler
	auth       *client.Auth
	testServer *httptest.Server
	t          *testing.T
}

// Generate a test API server which will be usable for testing API Calls
// if auth passed is nil, then no authentication occurs, otherwise U/P are expected for auth, and Token is returned
func NewMockTestServer(auth *client.Auth, t *testing.T) *MockTestServer {
	ms := MockTestServer{
		handlers: []MockTestHandler{},
		auth:     auth,
		t:        t,
	}
	ms.testServer = httptest.NewServer(http.HandlerFunc(ms.handle))

	if auth != nil {
		ms.AddHandler(http.MethodPost, client.URLTargetForAuth, MockServerHandlerGeneratorAuth(*auth))
	}

	return &ms
}

// Close the test server
func (s *MockTestServer) Close() {
	s.testServer.Close()
}

// Client generate a new api client against the server
func (s *MockTestServer) Client() (client.Client, error) {
	u, _ := url.Parse(s.testServer.URL)
	return client.NewClient(u, s.auth, s.testServer.Client())
}

// AddHandler add a handler for a path/method
func (s *MockTestServer) AddHandler(method string, path string, handler http.HandlerFunc) {
	s.handlers = append(s.handlers, MockTestHandler{
		method:  method,
		path:    path,
		handler: handler,
	})
}

// Handle an http request by finding the first matching Handler and running it
func (s *MockTestServer) handle(w http.ResponseWriter, r *http.Request) {
	// strip the leading slash from the path
	p := r.URL.Path
	_, i := utf8.DecodeRuneInString(p)
	p = p[i:]

	m := r.Method

	for _, h := range s.handlers {
		if h.Match(m, p) {
			h.handler(w, r)
			return
		}
	}

	s.t.Errorf("test server received request to unexpected target: Method:%s Path:%s :: %+v", m, p, s.handlers)
	w.WriteHeader(http.StatusNotFound)
}

// Shared handlers

// MockServerAuthHandler handle auth requests, check u/p and return token
func MockServerHandlerGeneratorAuth(auth client.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// Use these to set quick handlers for common actions such as status set, or quick return

// MockServerHandlerGeneratorReturnResponseStatus generates a http.HandlerFunc which just sets http status.
func MockServerHandlerGeneratorReturnResponseStatus(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

// MockServerHandlerGeneratorReturnBytes generates a http.HandlerFunc which just returns bytes.
func MockServerHandlerGeneratorReturnBytes(expected []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(expected) //nolint:errcheck
	}
}

// MockServerHandlerGeneratorReturnJson generatres a http.HandlerFunc which returns a JSON serialized argument.
func MockServerHandlerGeneratorReturnJson(expected interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expectedBytes, _ := json.Marshal(expected)
		w.Write(expectedBytes) //nolint:errcheck
	}
}
