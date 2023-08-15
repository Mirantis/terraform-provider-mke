//go:build integration
// +build integration

package client_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

const (
	ENVHost     = "MKE_HOST"
	ENVUsername = "MKE_USER"
	ENVPassword = "MKE_PASS"
)

var (
	ErrMissingIntegrationEnvVar        = errors.New("required env var prevented creation of an integration test client")
	ErrIntergrationClientGenerateError = errors.New("could not generate integration test client")
)

// MakeIntegrationClient creates a new MKE API Client from ENV variables and a client
func MakeIntegrationClient() (*client.Client, error) {
	var c *client.Client

	host, ok := os.LookupEnv(ENVHost)
	if !ok {
		return c, fmt.Errorf("%w; %s ENV variable missing", ErrMissingIntegrationEnvVar, ENVHost)
	} else if host == "" {
		return c, fmt.Errorf("%w; %s ENV variable empty", ErrMissingIntegrationEnvVar, ENVHost)
	}
	username, ok := os.LookupEnv(ENVUsername)
	if !ok {
		return c, fmt.Errorf("%w; %s ENV variable missing", ErrMissingIntegrationEnvVar, ENVUsername)
	} else if username == "" {
		return c, fmt.Errorf("%w; %s ENV variable empty", ErrMissingIntegrationEnvVar, ENVUsername)
	}
	password, ok := os.LookupEnv(ENVPassword)
	if !ok {
		return c, fmt.Errorf("%w; %s ENV variable missing", ErrMissingIntegrationEnvVar, ENVPassword)
	} else if password == "" {
		return c, fmt.Errorf("%w; %s ENV variable empty", ErrMissingIntegrationEnvVar, ENVPassword)
	}

	auth := client.NewAuthUP(username, password)

	apiURL, err := url.Parse(host)
	if err != nil {
		return c, fmt.Errorf("%w; %s", ErrIntergrationClientGenerateError, err)
	}

	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	if tempC, err := client.NewClient(apiURL, &auth, &hc); err != nil {
		return c, fmt.Errorf("%w; %s", ErrIntergrationClientGenerateError, err)
	} else {
		c = &tempC
	}

	return c, nil
}

func TestGenerateIntegrationTestClient(t *testing.T) {
	ctx := context.Background()

	if c, err := MakeIntegrationClient(); err != nil {
		t.Fatalf("Failed generating test client: %s", err)
	} else if c == nil {
		t.Fatal("Failed to create client")
	} else if err := c.ApiPing(ctx); err != nil {
		t.Fatalf("Failed pinging integration MKE client: %s", err)
	}
}
