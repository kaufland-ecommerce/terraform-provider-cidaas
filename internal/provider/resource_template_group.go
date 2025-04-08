package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
	"strings"
)

type templateGroupResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*templateGroupResource)(nil)

func NewTemplateGroupResource() resource.Resource {
	return &templateGroupResource{}
}

func (r *templateGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_group"
}

func (r *templateGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *templateGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_template_group` manages Template Groups in the tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique Name of the Template Group",
			},
			"comm_settings": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The communication settings for the Template Group",
				Attributes: map[string]schema.Attribute{
					"email": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Email communication configuration for the Template Group",
						Attributes: map[string]schema.Attribute{
							"communication_method": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"sender_address": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
						},
					},
					"ivr": schema.SingleNestedAttribute{
						Required:    true,
						Description: "IVR communication for the Template Group",
						Attributes: map[string]schema.Attribute{
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"communication_method": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"sender_address": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
						},
					},
					"push": schema.SingleNestedAttribute{
						Required:    true,
						Description: "PUSH communication for the Template Group",
						Attributes: map[string]schema.Attribute{
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"communication_method": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
					"sms": schema.SingleNestedAttribute{
						Required:    true,
						Description: "SMS communication for the Template Group",
						Attributes: map[string]schema.Attribute{
							"communication_method": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "The provider UUID used to send the notifications",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
							"sender_address": schema.StringAttribute{
								Required:    true,
								Description: "",
							},
						},
					},
				},
			},
			"default_locale": schema.StringAttribute{
				Required:    true,
				Description: "Default locale for the Template Group",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"tg_type": schema.StringAttribute{
				Computed:    true,
				Default:     stringdefault.StaticString("cidaas"),
				Description: "",
			},
		},
	}
}

func (r templateGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan TemplateGroup

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(plan.ID.ValueString()) == 0 {
		resp.Diagnostics.AddError("Attempting to create template group without ID", "Template group needs an ID")
		return
	}

	existingGroup, _ := r.provider.client.GetTemplateGroup(plan.ID.ValueString())

	if existingGroup != nil && existingGroup.ID == plan.ID.ValueString() {
		resp.Diagnostics.AddWarning("Attempting to create a template group with and existing ID", "Operation skipped. ID: "+plan.ID.ValueString())
		r.UpdateStateWithNewValues(existingGroup, &plan)

		diags = resp.State.Set(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		return
	}

	createGroupRequest := client.CreateTemplateGroupRequest{
		ID:            plan.ID.ValueString(),
		Description:   plan.Description.ValueString(),
		DefaultLocale: plan.DefaultLocale.ValueString(),
		TgType:        plan.TgType.ValueString(),
		Copy: client.CreateTemplateGroupCopy{
			// @FIXME: Currently we are not allowing to send this property from the clients
			FromGroupId: "default",
			Locale:      []client.CreateTemplateGroupCopyLocale{},
		},
	}
	createGroupRequest.Copy.Locale = append(createGroupRequest.Copy.Locale, client.CreateTemplateGroupCopyLocale{
		From: plan.DefaultLocale.ValueString(),
		To:   plan.DefaultLocale.ValueString(),
	})

	createGroupResponse, err := r.provider.client.CreateTemplateGroup(createGroupRequest)

	if err != nil {
		if strings.Contains(err.Error(), "already templates found for the locales") {
			resp.Diagnostics.AddWarning("Template group already exists", "Skipping operation")
			diags = resp.State.Set(ctx, &plan)
			resp.Diagnostics.Append(diags...)
			return
		}
		resp.Diagnostics.AddError(
			"Error creating template group",
			"Template group ID: "+plan.ID.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	r.UpdateStateWithNewValues(createGroupResponse, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r templateGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateGroup
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := state.ID.ValueString()

	if len(groupId) == 0 {
		resp.Diagnostics.AddError("Attempting to read group without ID", "Operation skipped and resource removed from state")
		return
	}

	templateGroup, err := r.provider.client.GetTemplateGroup(groupId)
	if err != nil {
		if err.Error() == "resource not found" {
			resp.Diagnostics.AddWarning("Template group not found", "Could not find template group with id "+groupId)
			req.State.RemoveResource(ctx)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Template Group",
			"Could not read template group with id "+groupId+": "+err.Error(),
		)
		return
	}

	r.UpdateStateWithNewValues(templateGroup, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r templateGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan TemplateGroup

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateGroupReq := client.TemplateGroup{
		ID:            plan.ID.ValueString(),
		Description:   plan.Description.ValueString(),
		DefaultLocale: plan.DefaultLocale.ValueString(),
		TgType:        "cidaas",
		CommSettings: client.TemplateGroupComSettings{
			Email: client.EmailSenderConfig{
				SenderAddress:       plan.CommSettings.Email.SenderAddress.ValueString(),
				ServiceSetupId:      plan.CommSettings.Email.ServiceSetupId.ValueString(),
				SenderName:          plan.CommSettings.Email.SenderName.ValueString(),
				CommunicationMethod: plan.CommSettings.Email.CommunicationMethod.ValueString(),
			},
			SMS: client.SmsSenderConfig{
				SenderAddress:       plan.CommSettings.SMS.SenderAddress.ValueString(),
				ServiceSetupId:      plan.CommSettings.SMS.ServiceSetupId.ValueString(),
				SenderName:          plan.CommSettings.SMS.SenderName.ValueString(),
				CommunicationMethod: plan.CommSettings.SMS.CommunicationMethod.ValueString(),
			},
			IVR: client.IVRSenderConfig{
				SenderAddress:       plan.CommSettings.IVR.SenderAddress.ValueString(),
				ServiceSetupId:      plan.CommSettings.IVR.ServiceSetupId.ValueString(),
				SenderName:          plan.CommSettings.IVR.SenderName.ValueString(),
				CommunicationMethod: plan.CommSettings.IVR.CommunicationMethod.ValueString(),
			},
			Push: client.PushSenderConfig{
				CommunicationMethod: plan.CommSettings.Push.CommunicationMethod.ValueString(),
				ServiceSetupId:      plan.CommSettings.Push.ServiceSetupId.ValueString(),
				SenderName:          plan.CommSettings.Push.SenderName.ValueString(),
			},
		},
	}

	if len(updateGroupReq.ID) == 0 {
		resp.Diagnostics.AddError("Unable to update template group with empty ID", "Make sure the template ID is set")
		return
	}

	groupResponse, err := r.provider.client.UpdateTemplateGroup(updateGroupReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Template Group",
			"Could not update group, unexpected error: "+err.Error(),
		)
		return
	}
	if groupResponse == nil {
		resp.Diagnostics.AddError(
			"Error updating Template Group",
			"Empty result after update",
		)
		return
	}

	r.UpdateStateWithNewValues(groupResponse, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r templateGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state TemplateGroup

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(state.ID.ValueString()) == 0 {
		resp.Diagnostics.AddWarning("Attempting to remove template group without ID", "Operation skipped")
		resp.State.RemoveResource(ctx)
		return
	}

	err := r.provider.client.DeleteTemplateGroup(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Template Group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r templateGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	group, err := r.provider.client.GetTemplateGroup(req.ID)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing Template Group",
			"Could not read template group with id "+req.ID+": "+err.Error(),
		)
		return
	}

	var state TemplateGroup

	r.UpdateStateWithNewValues(group, &state)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r templateGroupResource) UpdateStateWithNewValues(group *client.TemplateGroup, state *TemplateGroup) {
	state.ID = types.StringValue(group.ID)
	state.DefaultLocale = types.StringValue(group.DefaultLocale)
	state.TgType = types.StringValue(group.TgType)
	state.Description = types.StringValue(group.Description)
	state.CommSettings = TemplateGroupComSettings{
		Email: EmailSenderConfig{
			CommunicationMethod: types.StringValue(group.CommSettings.Email.CommunicationMethod),
			ServiceSetupId:      types.StringValue(group.CommSettings.Email.ServiceSetupId),
			SenderName:          types.StringValue(group.CommSettings.Email.SenderName),
			SenderAddress:       types.StringValue(group.CommSettings.Email.SenderAddress),
		},
		SMS: SmsSenderConfig{
			CommunicationMethod: types.StringValue(group.CommSettings.SMS.CommunicationMethod),
			ServiceSetupId:      types.StringValue(group.CommSettings.SMS.ServiceSetupId),
			SenderName:          types.StringValue(group.CommSettings.SMS.SenderName),
			SenderAddress:       types.StringValue(group.CommSettings.SMS.SenderAddress),
		},
		IVR: IVRSenderConfig{
			CommunicationMethod: types.StringValue(group.CommSettings.IVR.CommunicationMethod),
			ServiceSetupId:      types.StringValue(group.CommSettings.IVR.ServiceSetupId),
			SenderName:          types.StringValue(group.CommSettings.IVR.SenderName),
			SenderAddress:       types.StringValue(group.CommSettings.IVR.SenderAddress),
		},
		Push: PushSenderConfig{
			CommunicationMethod: types.StringValue(group.CommSettings.Push.CommunicationMethod),
			ServiceSetupId:      types.StringValue(group.CommSettings.Push.ServiceSetupId),
			SenderName:          types.StringValue(group.CommSettings.Push.SenderName),
		},
	}
}
