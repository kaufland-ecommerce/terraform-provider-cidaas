package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type notificationServiceSetupResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*notificationServiceSetupResource)(nil)
var _ resource.ResourceWithImportState = (*notificationServiceSetupResource)(nil)

func NewNotificationServiceSetupResource() resource.Resource {
	return &notificationServiceSetupResource{}
}

func (r *notificationServiceSetupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_service_setup"
}

func (r *notificationServiceSetupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *notificationServiceSetupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_notification_service_setup` manages a communication provider service setup (an email/sms/ivr/push provider connection) via `notifications-srv`. Pair with `cidaas_notification_provider_config` for the actual credentials. `status` is verified manually in the cidaas service-desk UI and is not driven by Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service setup ID, assigned by the API.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name of the service setup.",
			},
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Service ID, e.g. `custom-ses-email`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"communication_methods": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Communication methods handled by this setup: `email`, `sms`, `ivr`, `push`.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description.",
			},
			"has_remote_templates": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether templates for this setup are managed remotely.",
			},
			"parent_service_setup_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Parent service setup ID. Omit to let the platform auto-fill it when applicable; the value is stored after create/refresh.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_category": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("comm_prov"),
				Description: "Service category.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Setup status (`in-progress`, `active`, `inactive`). Verify manually in the cidaas service-desk to move it from `in-progress` to `active`.",
			},
		},
	}
}

// buildNotificationServiceSetup converts a Terraform plan into the client wire struct.
func buildNotificationServiceSetup(ctx context.Context, plan NotificationServiceSetup) (client.NotificationServiceSetup, diag.Diagnostics) {
	var methods []string
	diags := plan.CommunicationMethods.ElementsAs(ctx, &methods, false)
	if diags.HasError() {
		return client.NotificationServiceSetup{}, diags
	}

	return client.NotificationServiceSetup{
		ID:                   plan.ID.ValueString(),
		Name:                 plan.Name.ValueString(),
		ServiceId:            plan.ServiceId.ValueString(),
		CommunicationMethods: methods,
		Description:          plan.Description.ValueString(),
		HasRemoteTemplates:   plan.HasRemoteTemplates.ValueBool(),
		ParentServiceSetupId: plan.ParentServiceSetupId.ValueString(),
		ServiceCategory:      plan.ServiceCategory.ValueString(),
	}, diags
}

// notificationServiceSetupToState maps an API response onto Terraform state.
func notificationServiceSetupToState(ctx context.Context, setup *client.NotificationServiceSetup) (NotificationServiceSetup, diag.Diagnostics) {
	methods, diags := types.SetValueFrom(ctx, types.StringType, setup.CommunicationMethods)
	if diags.HasError() {
		return NotificationServiceSetup{}, diags
	}

	return NotificationServiceSetup{
		ID:                   types.StringValue(setup.ID),
		Name:                 types.StringValue(setup.Name),
		ServiceId:            types.StringValue(setup.ServiceId),
		CommunicationMethods: methods,
		Description:          types.StringValue(setup.Description),
		HasRemoteTemplates:   types.BoolValue(setup.HasRemoteTemplates),
		ParentServiceSetupId: types.StringValue(setup.ParentServiceSetupId),
		ServiceCategory:      types.StringValue(setup.ServiceCategory),
		Status:               types.StringValue(setup.Status),
	}, diags
}

func (r *notificationServiceSetupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan NotificationServiceSetup
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setup, diags := buildNotificationServiceSetup(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.CreateNotificationServiceSetup(setup)
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification service setup", "Could not create service setup, unexpected error: "+err.Error())
		return
	}

	state, diags := notificationServiceSetupToState(ctx, created)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationServiceSetupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationServiceSetup
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setup, err := r.provider.client.GetNotificationServiceSetup(state.ID.ValueString())
	if err != nil {
		if err.Error() == "resource not found" {
			req.State.RemoveResource(ctx)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading notification service setup",
			"Could not read service setup with id "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	newState, diags := notificationServiceSetupToState(ctx, setup)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationServiceSetupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan NotificationServiceSetup
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var state NotificationServiceSetup
	diags = append(diags, req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.provider.client.UpdateNotificationServiceSetup(state.ID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating notification service setup",
			"Service setup ID: "+state.ID.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	newState, sdiags := notificationServiceSetupToState(ctx, updated)
	resp.Diagnostics.Append(sdiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (r *notificationServiceSetupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var state NotificationServiceSetup
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.DeleteNotificationServiceSetup(state.ID.ValueString())
	if err != nil && err.Error() != "resource not found" {
		resp.Diagnostics.AddError(
			"Error deleting notification service setup",
			"Could not delete service setup, unexpected error: "+err.Error()+". If the setup is still active, deactivate it in the cidaas service-desk first.",
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *notificationServiceSetupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	setup, err := r.provider.client.GetNotificationServiceSetup(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing notification service setup",
			"Could not read service setup with id "+req.ID+": "+err.Error(),
		)
		return
	}

	state, diags := notificationServiceSetupToState(ctx, setup)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
