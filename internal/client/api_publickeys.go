package client

import (
	"context"
	"fmt"
	"net/http"
)

const (
	// /accounts/{accountNameOrID}/publicKeys url.
	URLTargetPatternForPublicKeys = "accounts/%s/publicKeys"
	// /accounts/{accountNameOrID}/publicKeys/{keyID} url.
	URLTargetPatternForPublicKey = "accounts/%s/publicKeys/%s"
)

type GetKeysResponse struct {
	AccountPubKeys []AccountPublicKey `json:"accountPublicKeys"`
	NextPageStart  string             `json:"nextPageStart"`
}

// ApiPublicKeyList list all of the public keys.
func (c *Client) ApiPublicKeyList(ctx context.Context, account string) ([]AccountPublicKey, error) {
	u := fmt.Sprintf(URLTargetPatternForPublicKeys, account)

	var keys []AccountPublicKey

	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodGet, u, []byte{})
	if err != nil {
		return keys, err
	}

	for {
		resp, err := c.doAuthorizedRequest(req)
		if err != nil {
			return keys, err
		}

		var respContents GetKeysResponse

		if err := resp.JSONMarshallBody(&respContents); err != nil {
			return keys, err
		}

		keys = append(keys, respContents.AccountPubKeys...)

		if respContents.NextPageStart == "" {
			break
		}

		reqQuery := req.URL.Query()
		reqQuery.Set("age", respContents.NextPageStart)
		req.URL.RawQuery = reqQuery.Encode()

	}

	return keys, nil
}

// ApiPublicKeyRetrieve retrieve a specific account key.
func (c *Client) ApiPublicKeyRetrieve(ctx context.Context, account, keyid string) (AccountPublicKey, error) {
	u := fmt.Sprintf(URLTargetPatternForPublicKey, account, keyid)

	var k AccountPublicKey

	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodGet, u, []byte{})
	if err != nil {
		return k, err
	}

	resp, err := c.doAuthorizedRequest(req)
	if err != nil {
		return k, err
	}

	if err = resp.JSONMarshallBody(&k); err != nil {
		return k, err
	}

	return k, nil
}

// ApiPublicKeyDelete delete a specific account key.
func (c *Client) ApiPublicKeyDelete(ctx context.Context, account, keyid string) error {
	u := fmt.Sprintf(URLTargetPatternForPublicKey, account, keyid)

	req, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodDelete, u, []byte{})
	if err != nil {
		return err
	}

	_, err = c.doAuthorizedRequest(req)
	return err
}
