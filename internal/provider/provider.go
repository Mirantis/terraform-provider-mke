package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
)

const (
	// ProviderName name for the provider in the Mirantis namespace.
	ProviderName = "mke"

	// TestingVersion if the provider version is this, then the provider
	//   will start in testing mode, where it never reaches the API.
	TestingVersion = "test"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MKEProvider{
			version: version,
		}
	}
}

// MKEProvider defines the provider implementation.
type MKEProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *MKEProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = ProviderName
	resp.Version = p.version
}

func (p *MKEProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "MKE API Endpoint address with schema; e.g. https://my.mke.com",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "MKE API username",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "MKE API user password",
				Required:            true,
				Sensitive:           true,
			},

			"unsafe_ssl_client": schema.BoolAttribute{
				MarkdownDescription: "Bypass SSL validation for hte API server. Use only for development systems",
				Optional:            true,
			},
		},
	}
}

func (p *MKEProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model MKEProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if p.version == TestingVersion {
		model.testingMode = types.BoolValue(true)
	}

	resp.ResourceData = model
	resp.DataSourceData = model

}

func (p *MKEProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMKEClientBundleResource,
	}
}

func (p *MKEProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// MKEProviderModel describes the provider data model.
type MKEProviderModel struct {
	testingMode types.Bool

	Endpoint  types.String `tfsdk:"endpoint"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	UnsafeSSL types.Bool   `tfsdk:"unsafe_ssl_client"`
}

// Client MKE client generation.
func (pm MKEProviderModel) Client() (client.Client, error) {
	if pm.UnsafeSSL.ValueBool() {
		return client.NewUnsafeSSLClient(pm.Endpoint.ValueString(), pm.Username.ValueString(), pm.Password.ValueString())
	}
	return client.NewClientSimple(pm.Endpoint.ValueString(), pm.Username.ValueString(), pm.Password.ValueString())
}

// TestingMode is the provider in testing mode?
func (pm MKEProviderModel) TestingMode() bool {
	return pm.testingMode.ValueBool()
}
