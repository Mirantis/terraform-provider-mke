package client_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestGoodPing(t *testing.T) {
	ctx := context.Background()
	mockRequest := MockHandlerKey{
		Path:   client.URLTargetForPing,
		Method: http.MethodGet,
	}

	svr := MockTestServer(nil, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnResponseStatus(http.StatusOK),
	})

	url, _ := url.Parse(svr.URL)
	c, err := client.NewClient(url, nil, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiPing(ctx); err != nil {
		t.Fatalf("Could not make a ping: %s", err)
	}
}

func TestNotFoundPing(t *testing.T) {
	ctx := context.Background()
	mockRequest := MockHandlerKey{
		Path:   client.URLTargetForPing,
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

	if err := c.ApiPing(ctx); err == nil {
		t.Fatalf("Ping was expected to fail: %s", err)
	}
}
