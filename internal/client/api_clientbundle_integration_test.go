//go:build integration
// +build integration

package client_test

import (
	"context"
	"testing"
)

func TestCreateAndDeleteClientBundle(t *testing.T) {
	ctx := context.Background()

	c, err := MakeIntegrationClient()
	if err != nil {
		t.Fatalf("Failed generating test client: %s", err)
	}

	cb, err := c.ApiClientBundleCreate(ctx)
	if err != nil {
		t.Fatalf("Failed generating client bundle: %s", err)
	}

	if _, err := c.ApiClientBundleGetPublicKey(ctx, cb); err != nil {
		t.Fatalf("Failed retrieving client bundle public key: %s", err)
	}

	if err := c.ApiClientBundleDelete(ctx, cb); err != nil {
		t.Fatalf("Failed deleting client bundle: %s", err)
	}
}
