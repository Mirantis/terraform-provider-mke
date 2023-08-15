package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestSimpleGetKeys(t *testing.T) {
	ctx := context.Background()
	auth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	mockRequest := MockHandlerKey{
		Path:   fmt.Sprintf(client.URLTargetPatternForPublicKeys, auth.Username),
		Method: http.MethodGet,
	}
	keysResp := client.GetKeysResponse{
		AccountPubKeys: []client.AccountPublicKey{
			{
				ID: "ASDF",
			},
		},
	}

	svr := MockTestServer(&auth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnJson(keysResp),
	})

	u, _ := url.Parse(svr.URL)
	c, err := client.NewClient(u, &auth, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	keys, err := c.ApiPublicKeyList(ctx, auth.Username)
	if err != nil {
		t.Fatalf("get keys request failed: %s", err)
	}

	if len(keys) == 0 {
		t.Error("no keys returned")
	}

}

func TestSimpleDeleteKey(t *testing.T) {
	ctx := context.Background()
	auth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	keyID := "ASDFASDF"
	mockRequest := MockHandlerKey{
		Path:   fmt.Sprintf(client.URLTargetPatternForPublicKey, auth.Username, keyID),
		Method: http.MethodDelete,
	}

	svr := MockTestServer(&auth, MockHandlerMap{
		mockRequest: MockServerHandlerGeneratorReturnResponseStatus(http.StatusOK),
	})

	u, _ := url.Parse(svr.URL)
	c, err := client.NewClient(u, &auth, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiPublicKeyDelete(ctx, auth.Username, keyID); err != nil {
		t.Fatalf("Failed to delete key: %s", err)
	}

}
