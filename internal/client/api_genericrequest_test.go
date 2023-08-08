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
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}
	expectedRespBodyBytes := []byte("myresponse")

	svr := MockTestServer(nil, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnBytes(expectedRespBodyBytes),
	})

	u, _ := url.Parse(svr.URL)
	c, err := client.NewClient(u, nil, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	req, err := c.RequestFromTargetAndBytesBody(ctx, mockRequest.Method, mockRequest.Path, []byte{})
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
		t.Fatalf("Authorized request execute did not produce and body: %s", err)
	}
	if string(responseBodyBytes) != string(expectedRespBodyBytes) {
		t.Errorf("Authorized request returned bad body: %s", string(responseBodyBytes))
	}

}

func TestGoodGenericRequestJSON(t *testing.T) {
	ctx := context.Background()
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodPost,
	}
	expectedResp := map[string]string{
		"first": "one",
	}

	svr := MockTestServer(nil, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnJson(expectedResp),
	})

	url, _ := url.Parse(svr.URL)
	c, err := client.NewClient(url, nil, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	req, err := c.RequestFromTargetAndBytesBody(ctx, mockRequest.Method, mockRequest.Path, []byte{})
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
		t.Fatalf("Authorized request execute did not produce and body: %s", err)
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
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}

	svr := MockTestServer(nil, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnResponseStatus(http.StatusNotFound),
	})

	url, _ := url.Parse(svr.URL)
	c, err := client.NewClient(url, nil, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	req, err := c.RequestFromTargetAndBytesBody(ctx, mockRequest.Method, mockRequest.Path, []byte{})
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
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}
	auth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}

	svr := MockTestServer(&auth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnBytes([]byte{}),
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

	resp, err := c.ApiAuthorizedGeneric(ctx, req)
	if err != nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("Authenticated request execute failed: %s; %s", err, b)
	}

}

func TestAuthenticatedPreventsUnauthenticatedRequest(t *testing.T) {
	ctx := context.Background()
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}
	auth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}

	svr := MockTestServer(&auth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnBytes([]byte{}),
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

	resp, err := c.ApiGeneric(ctx, req)
	if err == nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("UnAuthenticated request execute was allowed: %s; %s", err, b)
	}

}

func TestBadGenericAuthenticatedRequest(t *testing.T) {
	ctx := context.Background()
	mockRequest := MockHandlerKey{
		Path:   "mypath",
		Method: http.MethodGet,
	}
	srvAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	clAuth := client.Auth{
		Username: "notmyuser",
		Password: "notmypassword",
	}

	svr := MockTestServer(&srvAuth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnBytes([]byte{}),
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

	resp, err := c.ApiGeneric(ctx, req)
	if err == nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("Authenticated request execute passed when it should have failed: %s; %s", err, b)
	}

}
