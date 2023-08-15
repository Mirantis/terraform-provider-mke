package client_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestGoodAuthRequest(t *testing.T) {
	ctx := context.Background()
	srvAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	clAuth := client.Auth{
		Username: srvAuth.Username,
		Password: srvAuth.Password,
	}

	svr := MockTestServer(&srvAuth, MockHandlerMap{})

	u, _ := url.Parse(svr.URL)
	c, err := client.NewClient(u, &clAuth, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiLogin(ctx); err != nil {
		t.Error("Login request failed")
	}

	if clAuth.Token != srvAuth.Token {
		t.Errorf("ApiLogin did not set the expected token: %s != %s", clAuth.Token, srvAuth.Token)
	}
}

func TestBadAuthRequest(t *testing.T) {
	ctx := context.Background()
	srvAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}

	svr := MockTestServer(&srvAuth, MockHandlerMap{})

	url, _ := url.Parse(svr.URL)

	clAuthBadUsername := client.Auth{
		Username: "notmyser",
		Password: srvAuth.Password,
	}

	c, err := client.NewClient(url, &clAuthBadUsername, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiLogin(ctx); err == nil {
		t.Error("Login request did not fail with bad username")
	}

	clAuthBadPassword := client.Auth{
		Username: srvAuth.Username,
		Password: "notmypassword",
	}

	c, err = client.NewClient(url, &clAuthBadPassword, svr.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiLogin(ctx); err == nil {
		t.Error("Login request did not fail with bad password")
	}
}
