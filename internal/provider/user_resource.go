package provider

import (
	"context"
	"fmt"

	"github.com/Mirantis/terraform-provider-mke/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &UserResource{}

type UserResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Password types.String `tfsdk:"password"`
	FullName types.String `tfsdk:"full_name"`
	IsAdmin  types.Bool   `tfsdk:"is_admin"`
	Id       types.String `tfsdk:"id"`
}

type UserResource struct {
	providerModel MKEProviderModel
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(3, 16)},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password of the user",
				Required:            true,
				Sensitive:           true,
				Validators:          []validator.String{stringvalidator.LengthBetween(8, 16)},
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "The full name of the user",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"is_admin": schema.BoolAttribute{
				MarkdownDescription: "Is the user an admin",
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Optional:            true,
			},
		},
		MarkdownDescription: "User resource",
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	pass := data.Password.ValueString()
	if pass == "" {
		pass = client.GeneratePass()
		data.Password = basetypes.NewStringValue(pass)
	}

	acc := client.CreateAccount{
		Name:       data.Name.ValueString(),
		Password:   pass,
		FullName:   data.FullName.ValueString(),
		IsAdmin:    data.IsAdmin.ValueBool(),
		IsOrg:      false,
		SearchLDAP: false,
	}

	if resp.Diagnostics.HasError() {
		return
	}

	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}
	if r.providerModel.TestingMode() {
		resp.Diagnostics.AddWarning("testing mode warning", "mke user resource handler is in testing mode, no creation will be run.")
		data.Id = basetypes.NewStringValue(TestingVersion)
	} else {
		rAcc, err := cl.ApiCreateAccount(ctx, acc)
		if err != nil {
			resp.Diagnostics.AddError("Create account error", err.Error())
			return
		}

		tflog.Trace(ctx, fmt.Sprintf("created User resource `%s`", data.Name.ValueString()))

		data.Id = basetypes.NewStringValue(rAcc.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read user resource")
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}
	if r.providerModel.TestingMode() {
		resp.Diagnostics.AddWarning("testing mode warning", "mke user resource handler is in testing mode, no read will be run.")
		data.Id = types.StringValue(TestingVersion)
	} else {
		rAcc, err := cl.ApiReadAccount(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read account error", err.Error())
			return
		}
		data.Id = types.StringValue(rAcc.ID)
		data.Name = types.StringValue(rAcc.Name)
		data.FullName = types.StringValue(rAcc.FullName)
		data.IsAdmin = types.BoolValue(rAcc.IsAdmin)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update user resource")

	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}
	if r.providerModel.TestingMode() {
		resp.Diagnostics.AddWarning("testing mode warning", "mke user resource handler is in testing mode, no update will be run.")
		data.Id = types.StringValue(TestingVersion)
	} else {
		user := client.UpdateAccount{
			FullName: data.FullName.ValueString(),
			IsAdmin:  data.IsAdmin.ValueBool(),
		}
		rAcc, err := cl.ApiUpdateAccount(ctx, data.Id.ValueString(), user)
		tflog.Debug(ctx, fmt.Sprintf("The retuerned 'user' %+v", rAcc))

		if err != nil {
			resp.Diagnostics.AddError("Update account error", err.Error())
			return
		}

		// Overwrite user with refreshed state
		data.Id = types.StringValue(rAcc.ID)
		data.Name = types.StringValue(rAcc.Name)
		data.FullName = types.StringValue(rAcc.FullName)
		data.IsAdmin = types.BoolValue(rAcc.IsAdmin)
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	tflog.Debug(ctx, "Updated 'user' resource", map[string]any{"success": true})
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	cl, err := r.providerModel.Client()
	if err != nil {
		resp.Diagnostics.AddError("MKE provider could not create a client", fmt.Sprintf("An error occurred creating the client: %s", err.Error()))
		return
	}
	if r.providerModel.TestingMode() {
		resp.Diagnostics.AddWarning("testing mode warning", "mke user resource handler is in testing mode, no deletion will be run.")
	} else if err := cl.ApiDeleteAccount(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete account error", err.Error())
		return
	}

	tflog.Debug(ctx, "Deleted user resource", map[string]any{"success": true})
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
