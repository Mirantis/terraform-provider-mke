package client

import (
	"context"

	"net/http"
)

const (
	URLTargetForPing = "_ping"
)

// ApiPing Ping the endpoint
// @note MKE allows node specific pings, and a loadbalancer ping will
//
//	just connect to any node. This makes this precarious for cluster health.
func (c *Client) ApiPing(ctx context.Context) error {
	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodGet, URLTargetForPing, []byte{})
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
