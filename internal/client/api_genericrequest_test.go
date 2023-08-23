package client_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestGoodGenericRequest(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	method := http.MethodGet
	path := "my/path"
	expectedRespBodyBytes := []byte("myresponse")

	s := NewMockTestServer(&auth, t)
	s.AddHandler(method, path, MockServerHandlerGeneratorReturnBytes(expectedRespBodyBytes))
	defer s.Close()

	c, _ := s.Client()

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	resp, err := c.ApiGeneric(ctx, req)
	if err != nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("Generic request execute failed: %s; %s", err, b)
	}

	responseBodyBytes, err := resp.BodyBytes()
	if err != nil {
		t.Fatalf("Generid request execute did not produce and body: %s", err)
	}
	if string(responseBodyBytes) != string(expectedRespBodyBytes) {
		t.Errorf("Generic request returned bad body: %s", string(responseBodyBytes))
	}

}

func TestGoodGenericRequestJSON(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	method := http.MethodPost
	path := "my/path"
	expectedResp := map[string]string{
		"first":  "one",
		"second": "two",
	}

	s := NewMockTestServer(&auth, t)
	s.AddHandler(method, path, MockServerHandlerGeneratorReturnJson(expectedResp))
	defer s.Close()

	c, _ := s.Client()

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	resp, err := c.ApiGeneric(ctx, req)
	if err != nil {
		b := []byte{}
		resp.Body.Read(b) //nolint:errcheck
		t.Fatalf("Generic request execute failed: %s; %s", err, b)
	}

	var responseBodyMap map[string]string

	if err := resp.JSONMarshallBody(&responseBodyMap); err != nil {
		t.Fatalf("Generic request execute did not produce and body: %s", err)
	}

	for k, v := range responseBodyMap {
		if ev, ok := expectedResp[k]; !ok {
			t.Errorf("JSON Body missing key: %s", k)
		} else if ev != v {
			t.Errorf("JSON Body had wrong value for %s: %s != %s", k, v, ev)
		}
	}

}

func TestBadRequestNotFound(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	method := http.MethodPost
	path := "my/path"

	s := NewMockTestServer(&auth, t)
	s.AddHandler(method, path, MockServerHandlerGeneratorReturnResponseStatus(http.StatusNotFound))
	defer s.Close()

	c, _ := s.Client()

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	if _, err := c.ApiGeneric(ctx, req); err == nil {
		t.Error("BadRequest didn't fail")
	} else if !errors.Is(err, client.ErrUnknownTarget) {
		t.Errorf("BadRequest did not give the right error.")
	}
}

func TestGoodGenericAuthenticatedRequest(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	method := http.MethodPost
	path := "my/path"

	s := NewMockTestServer(&auth, t)
	s.AddHandler(method, path, MockServerHandlerGeneratorReturnBytes([]byte{}))
	defer s.Close()

	c, _ := s.Client()

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	resp, err := c.ApiAuthorizedGeneric(ctx, req)
	if err != nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("Authenticated request execute failed: %s; %s", err, b)
	}

}

func TestAuthenticatedPreventsUnauthenticatedRequest(t *testing.T) {
	ctx := context.Background()

	method := http.MethodGet
	path := "mypath"

	serverAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	clientAuth := client.Auth{
		Username: "notmyuser",
		Password: "notmypassword",
	}

	s := NewMockTestServer(&serverAuth, t)
	s.AddHandler(method, path, MockServerHandlerGeneratorReturnBytes([]byte{}))
	defer s.Close()

	u, _ := url.Parse(s.testServer.URL)
	c, _ := client.NewClient(u, &clientAuth, s.testServer.Client())

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	resp, err := c.ApiAuthorizedGeneric(ctx, req)
	if err == nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("UnAuthenticated request execute was allowed: %s; %s", err, b)
	}

}
