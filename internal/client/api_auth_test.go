package client_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestGoodAuthRequest(t *testing.T) {
	ctx := context.Background()

	serverAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	clientAuth := client.Auth{
		Username: serverAuth.Username,
		Password: serverAuth.Password,
	}

	s := NewMockTestServer(&serverAuth, t)
	defer s.Close()

	u, _ := url.Parse(s.testServer.URL)
	c, err := client.NewClient(u, &clientAuth, s.testServer.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiLogin(ctx); err != nil {
		t.Error("Login request failed")
	}

	if clientAuth.Token != serverAuth.Token {
		t.Errorf("ApiLogin did not set the expected token: %s != %s", clientAuth.Token, serverAuth.Token)
	}
}

func TestBadAuthRequest(t *testing.T) {
	ctx := context.Background()
	serverAuth := client.Auth{
		Username: "myuser",
		Password: "mypassword",
		Token:    "mytoken",
	}
	clientAuthBadUsername := client.Auth{
		Username: "notmyser",
		Password: serverAuth.Password,
	}
	clientAuthBadPassword := client.Auth{
		Username: serverAuth.Username,
		Password: "notmypassword",
	}

	s := NewMockTestServer(&serverAuth, t)
	defer s.Close()

	u, _ := url.Parse(s.testServer.URL)

	c, err := client.NewClient(u, &clientAuthBadUsername, s.testServer.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiLogin(ctx); err == nil {
		t.Error("Login request did not fail with bad username")
	}

	c, err = client.NewClient(u, &clientAuthBadPassword, s.testServer.Client())
	if err != nil {
		t.Fatalf("Could not make a client: %s", err)
	}

	if err := c.ApiLogin(ctx); err == nil {
		t.Error("Login request did not fail with bad password")
	}
}
