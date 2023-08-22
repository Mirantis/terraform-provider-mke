package client_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestGoodPing(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodGet, client.URLTargetForPing, MockServerHandlerGeneratorReturnResponseStatus(http.StatusOK))
	defer s.Close()

	c, _ := s.Client()

	if err := c.ApiPing(ctx); err != nil {
		t.Fatalf("Could not make a ping: %s", err)
	}
}

func TestNotFoundPing(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodGet, client.URLTargetForPing, MockServerHandlerGeneratorReturnResponseStatus(http.StatusNotFound))
	defer s.Close()

	c, _ := s.Client()

	if err := c.ApiPing(ctx); err == nil {
		t.Fatalf("Ping was expected to fail: %s", err)
	}
}
