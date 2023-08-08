package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestRequestBodyBytes(t *testing.T) {
	ctx := context.Background()
	expectedBody := "ABCDEFGHIJKL"
	p := "mypath"
	u, _ := url.Parse("https://localhost")

	c, _ := client.NewClient(u, nil, nil)

	r, err := c.RequestFromTargetAndBytesBody(ctx, http.MethodPost, p, []byte(expectedBody))
	if err != nil {
		t.Error("Failed to read request bytes")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		t.Error("Failed to read request bytes")
	}

	if string(bodyBytes) != expectedBody {
		t.Errorf("JSON req builder has unexpected body bytes: %s != %s", string(bodyBytes), expectedBody)
	}
}

func TestRequestBodyJson(t *testing.T) {
	ctx := context.Background()
	body := map[string]string{
		"first":  "one",
		"second": "two",
	}
	p := "mypath"
	u, _ := url.Parse("https://localhost")

	c, _ := client.NewClient(u, nil, nil)

	r, err := c.RequestFromTargetAndJSONBody(ctx, http.MethodPost, p, body)
	if err != nil {
		t.Error("Failed to read request bytes")
	}

	expectedBodyBytes, _ := json.Marshal(body)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		t.Error("Failed to read request bytes")
	}

	if string(bodyBytes) != string(expectedBodyBytes) {
		t.Errorf("JSON req builder has unexpected body bytes: %s != %s", string(bodyBytes), string(expectedBodyBytes))
	}
}
