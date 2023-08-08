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
	auth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}
	expectedRespBodyBytes := []byte("myresponse")

	svr := MockTestServer(&auth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnBytes(expectedRespBodyBytes),
	})

	url, _ := url.Parse(svr.URL)
	c, err := client.NewClient(url, &auth, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	req, err := c.RequestFromTargetAndBytesBody(ctx, mockRequest.Method, mockRequest.Path, []byte{})
	if err != nil {
		t.Fatalf("Could not make a request: %s", err)
	}

	if _, err := c.ApiAuthorizedGeneric(ctx, req); err != nil {
		t.Errorf("Authorized request execute failed: %s", err)
	}

}

func TestBadAuthorizedRequest(t *testing.T) {
	ctx := context.Background()
	srvAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	clAuth := client.Auth{
		Username: "notmyuser",
		Password: "notmypassword",
	}
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}
	expectedRespBodyBytes := []byte("myresponse")

	svr := MockTestServer(&srvAuth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnBytes(expectedRespBodyBytes),
	})

	url, _ := url.Parse(svr.URL)
	c, err := client.NewClient(url, &clAuth, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	req, err := c.RequestFromTargetAndBytesBody(ctx, mockRequest.Method, mockRequest.Path, []byte{})
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
		t.Error("Bearer header token build fail")
	}
}
