package provider_test

import (
	"testing"

	fr_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
	"github.com/Mirantis/terraform-provider-mke/internal/provider"
)

func TestClientBundleResourceSanity(t *testing.T) {
	// Throw an exception if this resources doesn't meet the requirements of a Resource
	var _ fr_resource.Resource = &provider.MKEClientBundleResource{} //nolint:typecheck
}

func TestModelFromClientBundle(t *testing.T) {
	id := "id"
	meta_name := id // maybe we should not force this equivalence.
	kube_host := "https://my.kubernetes.example:6443"

	cb := client.ClientBundle{
		ID: id,

		Kube: &client.ClientBundleKube{
			Config: "APIVersion: 1.1",
			Host:   kube_host,
		},
		Meta: client.ClientBundleMeta{
			Name: meta_name,

			StackOrchestrator: client.ClientBundleMetaEndpointDocker,
		},
	}
	m := provider.ClientBundleResourceModel{}

	ds := m.FromClientBundle(cb)

	if ds.HasError() {
		for _, d := range ds {
			t.Errorf("Error occurred converting ClientBundle to Model: %s : %s", d.Detail(), d.Summary())
		}
	}

	if cb.ID != id {
		t.Errorf("CB ID was unset: %s", cb.ID)
	}
	if m.Id.ValueString() != id {
		t.Errorf("Incorrect client bundle to model: ID : %s != %s", cb.ID, m.Id.ValueString())
	}
	if m.KubeHost.ValueString() != kube_host {
		t.Errorf("Incorrect client bundle to model: KubeHost : %s != %s", cb.Kube.Host, m.KubeHost.ValueString())
	}
	if m.StackOrchestrator.ValueString() != client.ClientBundleMetaEndpointDocker {
		t.Errorf("Incorrect client bundle to model: StackOrchestrator : %s != %s", cb.Meta.StackOrchestrator, m.StackOrchestrator.ValueString())
	}
}

func TestModelToClientBundle(t *testing.T) {
	id := "id"
	kube_yaml := "ApiVersion: someversion"
	kube_host := "https://my.kubernetes.example:6443"
	ca_cert := "ca-cert"

	cb := client.ClientBundle{}
	m := provider.ClientBundleResourceModel{
		Id:     types.StringValue(id),
		CaCert: types.StringValue(ca_cert),

		KubeYaml: types.StringValue(kube_yaml),
		KubeHost: types.StringValue(kube_host),
	}

	ds := m.ToClientBundle(&cb)

	if ds.HasError() {
		for _, d := range ds {
			t.Errorf("Error occurred converting ClientBundle to Model: %s : %s", d.Detail(), d.Summary())
		}
	}

	if m.Id.ValueString() != id {
		t.Errorf("Model Id was unset; %s", m.Id.ValueString())
	}
	if cb.ID != id {
		t.Errorf("Incorrect model to clientbundle: ID : %s != %s", cb.ID, m.Id.ValueString())
	}
	if cb.CACert != ca_cert {
		t.Errorf("Incorrect model to clientbundle: CaCert: %s != %s", cb.CACert, m.CaCert.ValueString())
	}
	if cb.Kube.Config != kube_yaml {
		t.Errorf("Incorrect model to clientbundle: KubeYaml: %s != %s", cb.Kube.Config, m.KubeYaml.ValueString())
	}
}

func TestAccMKEClientBundleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMKEClientBundleResource_minimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mke_clientbundle.test", "label", "my client bundle"),
					resource.TestCheckResourceAttr("mke_clientbundle.test", "id", provider.DummyClientBundle.ID),
					resource.TestCheckResourceAttr("mke_clientbundle.test", "kube_yaml", provider.DummyClientBundle.Kube.Config),
					resource.TestCheckResourceAttr("mke_clientbundle.test", "ca_cert", provider.DummyClientBundle.CACert),
					resource.TestCheckResourceAttr("mke_clientbundle.test", "kube_host", provider.DummyClientBundle.Kube.Host),
				),
			},
		},
	})
}

func testAccMKEClientBundleResource_minimal() string {
	return `
provider "mke" {
    endpoint = "https://my.mke.test"
    username = "user"
    password = "password"
}

resource "mke_clientbundle" "test" {
    label = "my client bundle"
}
`
}
