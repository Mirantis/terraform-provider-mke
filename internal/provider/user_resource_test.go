package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
	provider "mke" {
		endpoint = "test"
		username = "test"
		password = "test"
	}`
)

const (
	TestingVersion = "test"
)

func TestUserResourceDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testUserResourceDefault(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mke_user.test", "name", TestingVersion),
				),
			},
			// ImportState testing
			{
				ResourceName: "mke_user.test",
				ImportState:  true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
				resource "mke_user" "test" {
				name = "blah"
				password = "blahblah"
				full_name = "blah"
			}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mke_user.test", "name", "blah"),
					resource.TestCheckResourceAttr("mke_user.test", "password", "blahblah"),
					resource.TestCheckResourceAttr("mke_user.test", "full_name", "blah"),
					resource.TestCheckResourceAttr("mke_user.test", "is_admin", "false"),
					resource.TestCheckResourceAttrSet("mke_user.test", "id"),
				),
			},
			// Delete is called implicitly
		},
	})
}

func testUserResourceDefault() string {
	return `
	resource "mke_user" "test" {
		name = "test"
		password = "testtest"
		full_name = "test"
	}`
}
