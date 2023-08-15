package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// RequestFromTarget build simple http.Request from relative API target and bytes array for a body.
func (c *Client) RequestFromTargetAndBytesBody(ctx context.Context, method, target string, body []byte) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, c.reqURLFromTarget(target), bytes.NewBuffer(body))
}

// RequestFromTarget build simple http.Request from relative API target and JSON serialized struct for a body.
func (c *Client) RequestFromTargetAndJSONBody(ctx context.Context, method, target string, body interface{}) (*http.Request, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return c.RequestFromTargetAndBytesBody(ctx, method, target, bodyBytes)
}
