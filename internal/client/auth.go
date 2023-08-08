package client

import (
	"fmt"
	"net/http"
)

/**

# Authentication handling

MKE implements authentication using a bearer token that can be generated using
a username/password login to an authentication API target.
It is unclear how long this token lasts.S

@TODO we should probably do more intelligent caching/renewal of tokens. We don't
	have enough awareness to know how to tune that yet.
*/

const (
	HeaderKeyAuthorization = "Authorization"
)

// Auth container for data related to authentication.
// @see MKE Auth struct for auth/login.
type Auth struct {
	Code     string `json:"code"`
	Password string `json:"password"`
	Token    string `json:"token"`
	UseTLS   bool   `json:"useTLS"`
	Username string `json:"username"`
}

// NewAuthSimple constructor for Auth from username and password.
func NewAuthUP(username, password string) Auth {
	return Auth{
		Username: username,
		Password: password,
	}
}

// authorizeRequest adds a token header to a request to authenticate it.
// This will retrieve a request if none has been retrieved.
// This does not validate the token in any way (yet).
func (c *Client) authorizeRequest(req *http.Request) error {
	if c.auth.Token == "" {
		if err := c.ApiLogin(req.Context()); err != nil {
			return err
		}
	}

	req.Header.Add(HeaderKeyAuthorization, BearerTokenHeaderValue(c.auth.Token))

	return nil
}

// BearerTokenHeaderValue convert an auth token into the auth header value.
func BearerTokenHeaderValue(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
