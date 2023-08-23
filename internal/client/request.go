package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

	r, err := c.RequestFromTargetAndBytesBody(ctx, method, target, bodyBytes)
	if err == nil {
		r.Header.Set("Content-Type", "application/json")
	}

	return r, err
}

func requestDebug(req *http.Request) string {
	bbr, _ := req.GetBody()
	bb, _ := io.ReadAll(bbr)

	re := struct {
		Headers http.Header `json:"headers"`
		Body    string      `json:"body"`
	}{
		Headers: req.Header,
		Body:    string(bb),
	}

	rj, _ := json.MarshalIndent(re, "\n", "  ")

	return string(rj)
}
