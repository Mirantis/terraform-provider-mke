package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestGoodAuthorizedRequest(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	method := http.MethodGet
	path := "mypath"
	expectedRespBodyBytes := []byte("myresponse")

	s := NewMockTestServer(&auth, t)
	s.AddHandler(method, path, MockServerHandlerGeneratorReturnBytes(expectedRespBodyBytes))
	defer s.Close()

	c, _ := s.Client()

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	if _, err := c.ApiAuthorizedGeneric(ctx, req); err != nil {
		t.Errorf("Authorized request execute failed: %s", err)
	}
}

func TestBadAuthorizedRequest(t *testing.T) {
	ctx := context.Background()

	method := http.MethodGet
	path := "mypath"
	expectedRespBodyBytes := []byte("myresponse")

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
	s.AddHandler(http.MethodGet, client.URLTargetForAuth, MockServerHandlerGeneratorReturnBytes(expectedRespBodyBytes))
	defer s.Close()

	u, _ := url.Parse(s.testServer.URL)
	c, err := client.NewClient(u, &clientAuth, s.testServer.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	req, err := c.RequestFromTargetAndBytesBody(ctx, method, path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	if _, err := c.ApiAuthorizedGeneric(ctx, req); err == nil {
		t.Error("Bad authorization in request did not produce an error")
	} else if !errors.Is(err, client.ErrUnauthorizedReq) {
		t.Errorf("Wrong error received for bad auth: %s", err)
	}

}

func TestBearerTokenHeaderStringGenerate(t *testing.T) {
	token := "ASDJFLKASDF"
	headerString := client.BearerTokenHeaderValue(token)

	if !strings.Contains(headerString, "Bearer") {
		t.Error("Bearer header test token build fail")
	}
}
