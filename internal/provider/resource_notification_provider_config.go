package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type notificationProviderConfigResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*notificationProviderConfigResource)(nil)
var _ resource.ResourceWithImportState = (*notificationProviderConfigResource)(nil)

func NewNotificationProviderConfigResource() resource.Resource {
	return &notificationProviderConfigResource{}
}

func (r *notificationProviderConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_provider_config"
}

func (r *notificationProviderConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *notificationProviderConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_notification_provider_config` stores the credentials for a `cidaas_notification_service_setup` via `notifications-srv`. `config_data` is the wizard-shaped JSON payload cidaas expects (`commProvider`, `commMethod`, `schemaData`); verification happens manually in the cidaas service-desk UI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as `service_setup_id`.",
			},
			"service_setup_id": schema.StringAttribute{
				Required:    true,
				Description: "The `cidaas_notification_service_setup` ID this config belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config_data": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "JSON-encoded provider credentials, e.g. `jsonencode({ commProvider = \"custom-ses-email\", commMethod = \"email\", schemaData = { ... } })`.",
			},
		},
	}
}

func (r *notificationProviderConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan NotificationProviderConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.provider.client.UpsertNotificationProviderConfig(plan.ServiceSetupId.ValueString(), plan.ConfigData.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating notification provider config", "Could not create provider config, unexpected error: "+err.Error())
		return
	}

	plan.ID = plan.ServiceSetupId

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationProviderConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationProviderConfig
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.provider.client.GetNotificationProviderConfig(state.ServiceSetupId.ValueString()); err != nil {
		if err.Error() == "resource not found" {
			req.State.RemoveResource(ctx)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading notification provider config",
			"Could not read provider config for service setup "+state.ServiceSetupId.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationProviderConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan NotificationProviderConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.provider.client.UpsertNotificationProviderConfig(plan.ServiceSetupId.ValueString(), plan.ConfigData.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating notification provider config", "Could not update provider config, unexpected error: "+err.Error())
		return
	}

	plan.ID = plan.ServiceSetupId

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete only removes the resource from Terraform state - notifications-srv has no DELETE
// for provider configs. Deleting the related service setup (when allowed) removes the
// remote configuration.
func (r *notificationProviderConfigResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

// ImportState can't recover config_data - notifications-srv may redact credential fields
// on read (see GetNotificationProviderConfig), so the next plan will show config_data
// changing from null to the configured value; that's expected.
func (r *notificationProviderConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_setup_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_data"), types.StringNull())...)
}
