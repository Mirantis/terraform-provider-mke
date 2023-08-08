package client

/**
A generic API request, primarily used for testing.

Try not to use these for things other than testing. It is primarily used for the
api_auth_test.go testing.

@TODO is this a security weakness?

*/

import (
	"context"
	"net/http"
)

// ApiGeneric send a generic http request to the MKE API.
func (c *Client) ApiGeneric(ctx context.Context, req *http.Request) (*Response, error) {
	return c.doRequest(req)
}

// ApiAuthorizedGeneric send a authenticated generic http request to the MKE API.
func (c *Client) ApiAuthorizedGeneric(ctx context.Context, req *http.Request) (*Response, error) {
	return c.doAuthorizedRequest(req)
}
