package client_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestSimpleGetKeys(t *testing.T) {
	ctx := context.Background()

	auth := commonTestAuth
	keysResp := client.GetKeysResponse{
		AccountPubKeys: []client.AccountPublicKey{
			{
				ID: "ASDF",
			},
		},
	}

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodGet, fmt.Sprintf(client.URLTargetPatternForPublicKeys, auth.Username), MockServerHandlerGeneratorReturnJson(keysResp))

	c, _ := s.Client()

	keys, err := c.ApiPublicKeyList(ctx, auth.Username)
	if err != nil {
		t.Fatalf("get keys request failed: %s", err)
	}

	if len(keys) == 0 {
		t.Error("no keys returned")
	}

}

func TestDeleteKeySuccess(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth
	keyID := "ASDFASDF"

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodDelete, fmt.Sprintf(client.URLTargetPatternForPublicKey, auth.Username, keyID), MockServerHandlerGeneratorReturnResponseStatus(http.StatusOK))

	c, _ := s.Client()

	if err := c.ApiPublicKeyDelete(ctx, auth.Username, keyID); err != nil {
		t.Fatalf("Failed to delete key: %s", err)
	}
}

func TestDeleteKeyFailure(t *testing.T) {
	ctx := context.Background()
	auth := commonTestAuth
	keyID := "ASDFASDF"

	s := NewMockTestServer(&auth, t)
	s.AddHandler(http.MethodDelete, fmt.Sprintf(client.URLTargetPatternForPublicKey, auth.Username, keyID), MockServerHandlerGeneratorReturnResponseStatus(http.StatusNotFound))

	c, _ := s.Client()

	if err := c.ApiPublicKeyDelete(ctx, auth.Username, keyID); err == nil {
		t.Fatalf("Did not receive expected delete error")
	}
}
