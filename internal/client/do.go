package client

import (
	"fmt"
	"io"
	"net/http"
)

/**
This is tested via the api_genericrequest.go public methods.
*/

// doAuthorizedRequest perform an http request for an endpoint that requires auth.
func (c *Client) doAuthorizedRequest(req *http.Request) (*Response, error) {
	if err := c.authorizeRequest(req); err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// doRequest perform http request, catch http errors and return response as io.ReaderCloser.
func (c *Client) doRequest(req *http.Request) (*Response, error) {
	apiRes, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	res := &Response{
		Response: apiRes,
	}

	if res.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)

		if res.StatusCode == http.StatusUnauthorized {
			return res, fmt.Errorf("%w: Unauthorized: %d : %s", ErrUnauthorizedReq, res.StatusCode, b)
		}
		if res.StatusCode == http.StatusNotFound {
			return res, fmt.Errorf("%w: Not Found: %d : %s", ErrUnknownTarget, res.StatusCode, b)
		}
		if res.StatusCode == http.StatusInternalServerError {
			return res, fmt.Errorf("%w: Server Error: %d : %s", ErrServerError, res.StatusCode, b)
		}

		return res, fmt.Errorf("%w: Status code: %d : %s", ErrResponseError, res.StatusCode, b)
	}

	return res, nil
}
