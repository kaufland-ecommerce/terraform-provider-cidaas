package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
	"github.com/real-digital/terraform-provider-cidaas/internal/util"
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
								Description: "The communication method used for email notifications",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "The service setup ID for email notifications, automatically assigned by the API",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "The name that will appear as the sender of email notifications",
							},
							"sender_address": schema.StringAttribute{
								Required:    true,
								Description: "The email address that will be used as the sender of email notifications",
							},
						},
					},
					"ivr": schema.SingleNestedAttribute{
						Required:    true,
						Description: "IVR communication for the Template Group",
						Attributes: map[string]schema.Attribute{
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "The name that will appear as the sender of IVR notifications",
							},
							"communication_method": schema.StringAttribute{
								Required:    true,
								Description: "The communication method used for IVR notifications",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "The service setup ID for IVR notifications, automatically assigned by the API",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"sender_address": schema.StringAttribute{
								Required:    true,
								Description: "The phone number or address that will be used as the sender of IVR notifications",
							},
						},
					},
					"push": schema.SingleNestedAttribute{
						Required:    true,
						Description: "PUSH communication for the Template Group",
						Attributes: map[string]schema.Attribute{
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "The name that will appear as the sender of PUSH notifications",
							},
							"communication_method": schema.StringAttribute{
								Required:    true,
								Description: "The communication method used for PUSH notifications",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "The service setup ID for PUSH notifications, automatically assigned by the API",
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
								Description: "The communication method used for SMS notifications",
							},
							"service_setup_id": schema.StringAttribute{
								Computed:    true,
								Description: "The service setup ID for SMS notifications, automatically assigned by the API",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"sender_name": schema.StringAttribute{
								Required:    true,
								Description: "The name that will appear as the sender of SMS notifications",
							},
							"sender_address": schema.StringAttribute{
								Required:    true,
								Description: "The phone number or address that will be used as the sender of SMS notifications",
							},
						},
					},
				},
			},
			"default_locale": schema.StringAttribute{
				Required:    true,
				Description: "Default locale for the Template Group (e.g., 'en-US', 'de-DE')",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the Template Group",
			},
			"tg_type": schema.StringAttribute{
				Computed:    true,
				Default:     stringdefault.StaticString("cidaas"),
				Description: "The type of the Template Group, defaults to 'cidaas'",
			},
		},
	}
}

func (r *templateGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
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

	createGroupRequest := client.CreateTemplateGroupRequest{
		ID:            plan.ID.ValueString(),
		Description:   plan.Description.ValueString(),
		DefaultLocale: plan.DefaultLocale.ValueString(),
		TgType:        plan.TgType.ValueString(),
		Copy: client.CreateTemplateGroupCopy{
			// We use "default" as the source template group to copy from
			FromGroupId: "default",
			Locale:      []client.CreateTemplateGroupCopyLocale{},
		},
		Owner: util.ToStringPointer("client"),
	}

	createGroupRequest.Copy.Locale = append(createGroupRequest.Copy.Locale, client.CreateTemplateGroupCopyLocale{
		From: plan.DefaultLocale.ValueString(),
		To:   plan.DefaultLocale.ValueString(),
	})

	if existingGroup != nil {
		resp.Diagnostics.AddWarning(
			"Attempting to create a template group with an existing ID",
			"Applying update instead for template group: "+plan.ID.ValueString(),
		)
		r.doUpsert(ctx, resp, &plan, existingGroup)
		return
	}

	createdGroup, err := r.provider.client.CreateTemplateGroup(createGroupRequest)

	if err != nil {
		reqStr, marshalErr := json.Marshal(createGroupRequest)
		errDetail := "Error: " + err.Error()

		if marshalErr == nil {
			errDetail += "\nRequest: " + string(reqStr)
		}

		resp.Diagnostics.AddError(
			"Error creating template group",
			errDetail,
		)
		return
	}

	if createdGroup == nil {
		resp.Diagnostics.AddError(
			"Error creating template group",
			"No group data returned after creation",
		)
		return
	}

	// As creating a template group uses a different payload in the create request,
	// we need to update the recently created group with the other parameters not sent in the create request
	r.doUpsert(ctx, resp, &plan, createdGroup)
}

// buildEmailConfig creates an EmailSenderConfig from plan and existing values
func buildEmailConfig(planConfig EmailSenderConfig, existingCommMethod, existingSenderName, existingSenderAddress, serviceSetupId string) client.EmailSenderConfig {
	commMethod := existingCommMethod
	if planConfig.CommunicationMethod.ValueString() != "" {
		commMethod = planConfig.CommunicationMethod.ValueString()
	}

	senderName := existingSenderName
	if planConfig.SenderName.ValueString() != "" {
		senderName = planConfig.SenderName.ValueString()
	}

	senderAddress := existingSenderAddress
	if planConfig.SenderAddress.ValueString() != "" {
		senderAddress = planConfig.SenderAddress.ValueString()
	}

	return client.EmailSenderConfig{
		CommunicationMethod: commMethod,
		SenderName:          senderName,
		SenderAddress:       senderAddress,
		ServiceSetupId:      serviceSetupId,
	}
}

// buildSmsConfig creates a SmsSenderConfig from plan and existing values
func buildSmsConfig(planConfig SmsSenderConfig, existingCommMethod, existingSenderName, existingSenderAddress, serviceSetupId string) client.SmsSenderConfig {
	commMethod := existingCommMethod
	if planConfig.CommunicationMethod.ValueString() != "" {
		commMethod = planConfig.CommunicationMethod.ValueString()
	}

	senderName := existingSenderName
	if planConfig.SenderName.ValueString() != "" {
		senderName = planConfig.SenderName.ValueString()
	}

	senderAddress := existingSenderAddress
	if planConfig.SenderAddress.ValueString() != "" {
		senderAddress = planConfig.SenderAddress.ValueString()
	}

	return client.SmsSenderConfig{
		CommunicationMethod: commMethod,
		SenderName:          senderName,
		SenderAddress:       senderAddress,
		ServiceSetupId:      serviceSetupId,
	}
}

// buildIvrConfig creates an IVRSenderConfig from plan and existing values
func buildIvrConfig(planConfig IVRSenderConfig, existingCommMethod, existingSenderName, existingSenderAddress, serviceSetupId string) client.IVRSenderConfig {
	commMethod := existingCommMethod
	if planConfig.CommunicationMethod.ValueString() != "" {
		commMethod = planConfig.CommunicationMethod.ValueString()
	}

	senderName := existingSenderName
	if planConfig.SenderName.ValueString() != "" {
		senderName = planConfig.SenderName.ValueString()
	}

	senderAddress := existingSenderAddress
	if planConfig.SenderAddress.ValueString() != "" {
		senderAddress = planConfig.SenderAddress.ValueString()
	}

	return client.IVRSenderConfig{
		CommunicationMethod: commMethod,
		SenderName:          senderName,
		SenderAddress:       senderAddress,
		ServiceSetupId:      serviceSetupId,
	}
}

// buildPushConfig creates a PushSenderConfig from plan and existing values
func buildPushConfig(planConfig PushSenderConfig, existingCommMethod, existingSenderName, serviceSetupId string) client.PushSenderConfig {
	commMethod := existingCommMethod
	if planConfig.CommunicationMethod.ValueString() != "" {
		commMethod = planConfig.CommunicationMethod.ValueString()
	}

	senderName := existingSenderName
	if planConfig.SenderName.ValueString() != "" {
		senderName = planConfig.SenderName.ValueString()
	}

	return client.PushSenderConfig{
		CommunicationMethod: commMethod,
		SenderName:          senderName,
		ServiceSetupId:      serviceSetupId,
	}
}

// processUpdateResponse handles the update response and updates the state
func (r *templateGroupResource) processUpdateResponse(ctx context.Context, resp *resource.CreateResponse, plan *TemplateGroup, updateResponse *client.TemplateGroup) {
	if updateResponse == nil {
		resp.Diagnostics.AddError(
			"Error updating Template Group",
			"Empty result after update operation",
		)
		return
	}

	r.UpdateStateWithNewValues(updateResponse, plan)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// createUpdateRequest creates a client.TemplateGroup for updating based on the existing group
func createUpdateRequest(
	plan *TemplateGroup,
	id string,
	description string,
	tgType string,
	defaultLocale string,
	emailServiceSetupId string,
	emailCommMethod string,
	emailSenderName string,
	emailSenderAddress string,
	smsServiceSetupId string,
	smsCommMethod string,
	smsSenderName string,
	smsSenderAddress string,
	ivrServiceSetupId string,
	ivrCommMethod string,
	ivrSenderName string,
	ivrSenderAddress string,
	pushServiceSetupId string,
	pushCommMethod string,
	pushSenderName string,
) client.TemplateGroup {
	return client.TemplateGroup{
		ID:            id,
		Description:   description,
		TgType:        tgType,
		DefaultLocale: defaultLocale,
		CommSettings: client.TemplateGroupComSettings{
			Email: buildEmailConfig(
				plan.CommSettings.Email,
				emailCommMethod,
				emailSenderName,
				emailSenderAddress,
				emailServiceSetupId,
			),
			SMS: buildSmsConfig(
				plan.CommSettings.SMS,
				smsCommMethod,
				smsSenderName,
				smsSenderAddress,
				smsServiceSetupId,
			),
			IVR: buildIvrConfig(
				plan.CommSettings.IVR,
				ivrCommMethod,
				ivrSenderName,
				ivrSenderAddress,
				ivrServiceSetupId,
			),
			Push: buildPushConfig(
				plan.CommSettings.Push,
				pushCommMethod,
				pushSenderName,
				pushServiceSetupId,
			),
		},
	}
}

func (r *templateGroupResource) doUpsert(ctx context.Context, resp *resource.CreateResponse, plan *TemplateGroup, existing *client.TemplateGroup) {
	groupToUpdate := createUpdateRequest(
		plan,
		existing.ID,
		existing.Description,
		existing.TgType,
		existing.DefaultLocale,
		existing.CommSettings.Email.ServiceSetupId,
		existing.CommSettings.Email.CommunicationMethod,
		existing.CommSettings.Email.SenderName,
		existing.CommSettings.Email.SenderAddress,
		existing.CommSettings.SMS.ServiceSetupId,
		existing.CommSettings.SMS.CommunicationMethod,
		existing.CommSettings.SMS.SenderName,
		existing.CommSettings.SMS.SenderAddress,
		existing.CommSettings.IVR.ServiceSetupId,
		existing.CommSettings.IVR.CommunicationMethod,
		existing.CommSettings.IVR.SenderName,
		existing.CommSettings.IVR.SenderAddress,
		existing.CommSettings.Push.ServiceSetupId,
		existing.CommSettings.Push.CommunicationMethod,
		existing.CommSettings.Push.SenderName,
	)

	// Update the template group
	updateResponse, errUpdate := r.provider.client.UpdateTemplateGroup(groupToUpdate)
	if errUpdate != nil {
		resp.Diagnostics.AddError(
			"Error updating template group",
			"Template group ID: "+plan.ID.ValueString()+", unexpected error: "+errUpdate.Error(),
		)
		return
	}

	// Process the update response
	r.processUpdateResponse(ctx, resp, plan, updateResponse)
}

func (r *templateGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			// Resource not found, remove it from state without warning
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

func (r *templateGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
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

func (r *templateGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
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

func (r *templateGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *templateGroupResource) UpdateStateWithNewValues(group *client.TemplateGroup, state *TemplateGroup) {
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
