package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	EndpointDefaultScheme = "https"
)

var (
	ErrCouldNotCreateClient = errors.New("could not create a client")
)

// Client MKE client.
type Client struct {
	apiURL     *url.URL
	auth       *Auth
	HTTPClient *http.Client
}

// NewClient from a string URL and u/p.
func NewClientSimple(endpoint, username, password string) (Client, error) {
	HTTPClient := &http.Client{}
	auth := NewAuthUP(username, password)

	apiURL, err := url.Parse(endpoint)
	if err != nil {
		return Client{}, err
	}

	return NewClient(apiURL, &auth, HTTPClient)
}

// NewUnsafeSSLClient that allows self-signed SSL from a string URL and u/p.
func NewUnsafeSSLClient(endpoint, username, password string) (Client, error) {
	HTTPClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	auth := NewAuthUP(username, password)

	apiURL, err := url.Parse(endpoint)
	if err != nil {
		return Client{}, fmt.Errorf("%w; %s; empty endpoint", ErrCouldNotCreateClient, err)
	}

	return NewClient(apiURL, &auth, HTTPClient)
}

// NewClient creates a new MKE API Client from raw components.
func NewClient(apiURL *url.URL, auth *Auth, HTTPClient *http.Client) (Client, error) {
	if apiURL == nil {
		return Client{}, fmt.Errorf("%w; empty endpoint", ErrCouldNotCreateClient)
	}
	return Client{
		apiURL:     apiURL,
		HTTPClient: HTTPClient,
		auth:       auth,
	}, nil
}

// Build a request URL string from the client endpoint and an API target path.
func (c *Client) reqURLFromTarget(target string) string {
	// target should be a relative path, and will be treated as a relative reference
	// to the client URL
	// @see https://pkg.go.dev/net/url#URL.ResolveReference
	if c == nil {
		panic("Tried to generate relative URL from a client, but the client was nil")
	}
	if c.apiURL == nil {
		panic("Tried to generate relative URL from a client, but the client URL was nil")
	}
	targetURL, err := url.Parse(target)
	if err != nil {
		panic(fmt.Errorf("tried to generate relative URL from a client: %s", err))
	}
	relativeURL := c.apiURL.ResolveReference(targetURL)

	return relativeURL.String()
}

// Username retrieve username string for auth, so that we don't expose the whole auth struct.
func (c *Client) Username() string {
	return c.auth.Username
}
