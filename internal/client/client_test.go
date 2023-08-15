package client_test

import (
	"testing"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

func TestClientSimple(t *testing.T) {
	u := "me"
	p := "asdfasdf"

	c, err := client.NewClientSimple("/", u, p)
	if err != nil {
		t.Fatal("failed to create client using simple constructor.")
	}

	if c.Username() != u {
		t.Errorf("Simple client had bad username: %s != %s", c.Username(), u)
	}

}
