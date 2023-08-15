package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	ErrCBNotFound = errors.New("Client bundle was not found on the MKE host cluster")

	// DummyClientBundle client bundle used during unit ACC tests
	DummyClientBundle client.ClientBundle = client.ClientBundle{
		ID: "id",

		PrivateKey: "my-priv-key",
		PublicKey:  "my-pub-key",
		Cert:       "my-cert",
		CACert:     "my-ca-cert",
		Kube: &client.ClientBundleKube{
			Config:   "my-kube-yaml",
			Host:     "my-kube-host",
			Insecure: "my-kube-insecure",
		},
		Meta: client.ClientBundleMeta{
			Name:        "id",
			Description: "my-cluster",

			KubernetesHost:          "my-kube-host",
			KubernetesSkipVerifyTLS: true,

			DockerHost:          "my-docker-host",
			DockerSkipVerifyTLS: true,

			StackOrchestrator: "kubernetes",
		},
	}
)

// ClientBundleResourceModel describes the resource data model.
type ClientBundleResourceModel struct {
	Id types.String `tfsdk:"id"`

	Label types.String `tfsdk:"label"`

	PublicKey  types.String `tfsdk:"public_key"`
	PrivateKey types.String `tfsdk:"private_key"`
	ClientCert types.String `tfsdk:"client_cert"`
	CaCert     types.String `tfsdk:"ca_cert"`

	KubeYaml          types.String `tfsdk:"kube_yaml"`
	KubeHost          types.String `tfsdk:"kube_host"`
	KubeSkipTLSVerify types.Bool   `tfsdk:"kube_skiptlsverify"`

	DockerHost          types.String `tfsdk:"docker_host"`
	DockerSkipTLSVerify types.Bool   `tfsdk:"docker_skiptlsverify"`

	StackOrchestrator types.String `tfsdk:"orchestrator"`
}

// FromClientBundle interpret a client.ClientBundle to populate this model.
func (m *ClientBundleResourceModel) FromClientBundle(cb client.ClientBundle) diag.Diagnostics {
	diags := diag.Diagnostics{}

	m.Id = types.StringValue(cb.Meta.Name)

	m.PublicKey = types.StringValue(cb.PublicKey)
	m.PrivateKey = types.StringValue(cb.PrivateKey)
	m.ClientCert = types.StringValue(cb.Cert)
	m.CaCert = types.StringValue(cb.CACert)

	if cb.Kube != nil {
		// if the client bundle came with Kube values use them
		m.KubeYaml = types.StringValue(cb.Kube.Config)
		m.KubeHost = types.StringValue(cb.Kube.Host)
		m.KubeSkipTLSVerify = types.BoolValue(cb.Meta.KubernetesSkipVerifyTLS)
	} else if cb.Meta.KubernetesHost != "" {
		// otherwise if the cb has meta data for kube, use that
		m.KubeHost = types.StringValue(cb.Meta.KubernetesHost)
		m.KubeSkipTLSVerify = types.BoolValue(cb.Meta.KubernetesSkipVerifyTLS)
	} else {
		// otherwise we are likely dealing with swarm only MKE
		m.KubeYaml = types.StringNull()
		m.KubeHost = types.StringNull()
		m.KubeSkipTLSVerify = types.BoolNull()
	}

	m.DockerHost = types.StringValue(cb.Meta.DockerHost)
	m.DockerSkipTLSVerify = types.BoolValue(cb.Meta.DockerSkipVerifyTLS)

	m.StackOrchestrator = types.StringValue(cb.Meta.StackOrchestrator)

	return diags
}

// ToClientBundle convert this model to a client.ClientBundle struct.
func (m ClientBundleResourceModel) ToClientBundle(cb *client.ClientBundle) diag.Diagnostics {
	diags := diag.Diagnostics{}

	cb.ID = m.Id.ValueString()

	cb.PublicKey = m.PublicKey.ValueString()
	cb.PrivateKey = m.PrivateKey.ValueString()
	cb.Cert = m.ClientCert.ValueString()
	cb.CACert = m.CaCert.ValueString()

	if !(m.KubeYaml.IsUnknown() && m.KubeYaml.IsNull()) {
		if cb.Kube == nil {
			cb.Kube = &client.ClientBundleKube{}
		}

		cb.Kube.Config = m.KubeYaml.ValueString()
		cb.Kube.Host = m.KubeHost.ValueString()
		cb.Meta.KubernetesHost = m.KubeHost.ValueString()

		cb.Kube.ClientKey = m.PrivateKey.ValueString()
		cb.Kube.ClientCertificate = m.ClientCert.ValueString()
		cb.Kube.CACertificate = m.CaCert.ValueString()
	}

	cb.Meta.DockerHost = m.DockerHost.ValueString()
	cb.Meta.DockerSkipVerifyTLS = m.DockerSkipTLSVerify.ValueBool()

	cb.Meta.StackOrchestrator = m.StackOrchestrator.ValueString()

	return diags
}

type MKEClientBundleResource struct {
	providerModel MKEProviderModel
}

func NewMKEClientBundleResource() resource.Resource {
	return &MKEClientBundleResource{}
}

func (r *MKEClientBundleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clientbundle"
}

func (r *MKEClientBundleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Client Bundle resource for access to MKE api/kubernetes/docker-swarm features.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"label": schema.StringAttribute{
				MarkdownDescription: "Label used for the client bundle",
				Required:            true,
			},

			"public_key": schema.StringAttribute{
				MarkdownDescription: "MKE Public key for the user",
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "MKE Private key for the user",
				Computed:            true,
				Sensitive:           true,
			},
			"client_cert": schema.StringAttribute{
				MarkdownDescription: "MKE Client certificate for the user",
				Computed:            true,
			},
			"ca_cert": schema.StringAttribute{
				MarkdownDescription: "MKE Server CA certificate",
				Computed:            true,
			},

			"kube_yaml": schema.StringAttribute{
				MarkdownDescription: "MKE Kubernetes API client configuration yaml file",
				Computed:            true,
				Sensitive:           true,
			},
			"kube_host": schema.StringAttribute{
				MarkdownDescription: "MKE Kubernetes API host endpoint",
				Computed:            true,
			},
			"kube_skiptlsverify": schema.BoolAttribute{
				MarkdownDescription: "MKE Kubernetes endpoint TLS should not be verified",
				Computed:            true,
			},

			"docker_host": schema.StringAttribute{
				MarkdownDescription: "MKE Docker swarm endpoint",
				Computed:            true,
			},
			"docker_skiptlsverify": schema.BoolAttribute{
				MarkdownDescription: "MKE Docker endpoint TLS should not be verified",
				Computed:            true,
			},

			"orchestrator": schema.StringAttribute{
				MarkdownDescription: "Stack Orchestrator for the MKE instance, either 'docker' for docker-swarm, 'kubernetes', or 'all'",
				Computed:            true,
			},
		},
	}
}

func (r *MKEClientBundleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	lpm, ok := req.ProviderData.(MKEProviderModel)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *MKEProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	tflog.Debug(ctx, "Successfully interpeted provider model", map[string]interface{}{})
	r.providerModel = lpm
}

func (r *MKEClientBundleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m ClientBundleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to interpret plan on create", "Could not interpret plan into ClientBundle request")
		return
	}

	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}

	if r.providerModel.TestingMode() {
		cb := DummyClientBundle
		resp.Diagnostics.AddWarning("ClientBundle in testing mode", fmt.Sprintf("Client Bundle creation not executed because the resource is in testing mode: %s", cb.ToJSON()))

		m.FromClientBundle(cb)
	} else {
		cb, err := cl.ApiClientBundleCreate(ctx, m.Label.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("MKE client could not create a client bundle", fmt.Sprintf("An error occurred creating the client bundle: %s", err.Error()))
			return
		}

		resp.Diagnostics.Append(m.FromClientBundle(cb)...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "Failed to convert ClientBundle response from the API into the ClientBundle models", map[string]interface{}{})
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *MKEClientBundleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m ClientBundleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to interpret plan on read", "Could not interpret plan into ClientBundle model")
		return
	}
	// If there is no existing state, then there is nothing to read.
	if m.Id.IsNull() {
		return
	}
	tflog.Warn(ctx, "Read() found state:", map[string]interface{}{"model": m})
	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}
	cb := client.ClientBundle{}
	resp.Diagnostics.Append(m.ToClientBundle(&cb)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to convert Client Bundle Model", "Could not interpret plan Client Bundle model into client ClientBundle")
		return
	}
	if r.providerModel.TestingMode() {
		resp.Diagnostics.AddWarning("ClientBundle in testing mode", "ClientBundle read/confirm not executed because the resource is in testing mode")
		return
	}
	if _, err := cl.ApiClientBundleGetPublicKey(ctx, cb); err != nil {
		if errors.Is(err, client.ErrFailedToFindClientBundleMKEPublicKey) {
			// we have a bundle in state, but it doesn't exist in MKE so it should be removed
			// @todo check that we haven't suffered from a connectivity failure
			resp.Diagnostics.AddWarning("Client Bundle in state not found in MKE API", fmt.Errorf("%w; %s", ErrCBNotFound, err).Error())
			resp.State.RemoveResource(ctx)
		}
	}
}

func (r *MKEClientBundleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// There is no option to update a client bundle, but at the same time there is no schema that can be changed, so an update shouldn't be needed.
	// There is an issue of expiry options on client bundles, but this is a new feature in MKE that wasn't rolled out 2023/08
}

func (r *MKEClientBundleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m ClientBundleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to interpret state on delete", "Could not interpret plan into ClientBundle model")
		return
	}

	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}

	cb := client.ClientBundle{}
	resp.Diagnostics.Append(m.ToClientBundle(&cb)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to convert Client Bundle Model", "Could not interpret plan Client Bundle model into client ClientBundle")
	}

	if r.providerModel.TestingMode() {
		resp.Diagnostics.AddWarning("ClientBundle in testing mode", "ClientBundle deletion not executed because the resource is in testing mode")
		resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
	} else {

		if err := cl.ApiClientBundleDelete(ctx, cb); err != nil {
			resp.Diagnostics.AddError("Failed to delete Client Bundle", fmt.Sprintf("MKE Client could not delete the client bundle: %s", err.Error()))
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}
