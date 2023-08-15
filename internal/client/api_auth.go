package client

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	URLTargetForAuth = "auth/login"
)

// MKE API json response for successful token retrieval.
type loginResponse struct {
	Token string `json:"auth_token"`
}

// NewLoginResponse create a login response.
func NewLoginResponse(token string) loginResponse {
	return loginResponse{Token: token}
}

func (lr loginResponse) Bytes() []byte {
	lrb, _ := json.Marshal(lr)
	return lrb
}

// apiLogin update client Auth with a new token from an API auth request.
func (c *Client) ApiLogin(ctx context.Context) error {
	req, err := c.RequestFromTargetAndJSONBody(ctx, http.MethodPost, URLTargetForAuth, c.auth)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}

	var loginResp loginResponse

	if err := resp.JSONMarshallBody(&loginResp); err != nil {
		return err
	}

	c.auth.Token = loginResp.Token

	return nil
}
