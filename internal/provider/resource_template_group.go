package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
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
		Description: "`cidaas_template_group` manages Template Groups in the tenant.\n\n",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique Name of the Template Group",
				//PlanModifiers: []planmodifier.String{
				//	stringplanmodifier.RequiresReplace(),
				//},
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
				},
			},
			"default_locale": schema.StringAttribute{
				Required:    true,
				Description: "Default locale for the Template Group",
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

	templateGroup, err := r.provider.client.CreateTemplateGroup(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template group",
			"Could not create hook, unexpected error: "+err.Error(),
		)
		return
	}

	result := TemplateGroup{
		ID: types.StringValue(templateGroup.Id),
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r templateGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateGroup
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := state.ID.ValueString()

	templateGroup, err := r.provider.client.GetTemplateGroup(groupId)

	if err != nil {
		// @FIXME: Does it make sense to skip if template group not found?
		if err.Error() == "resource not found" {
			resp.Diagnostics.AddWarning("Skipped templated group not found", "Could not find template group with id "+groupId)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Template Group",
			"Could not read template group with id "+groupId+": "+err.Error(),
		)
		return
	}

	r.ModelToState(ctx, templateGroup, &state)

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

	group := client.TemplateGroup{
		Id:            plan.ID.ValueString(),
		DefaultLocale: plan.DefaultLocale.ValueString(),
	}

	diags = plan.CommSettings.As(ctx, &group.CommSettings, struct {
		UnhandledNullAsEmpty    bool
		UnhandledUnknownAsEmpty bool
	}{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.UpdateTemplateGroup(&group)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Template Group",
			"Could not update group, unexpected error: "+err.Error(),
		)
		return
	}

	var state TemplateGroup
	r.ModelToState(ctx, &group, &state)

	diags = resp.State.Set(ctx, &state)
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
			"Error reading Template Group",
			"Could not read template group with id "+req.ID+": "+err.Error(),
		)
		return
	}

	var state TemplateGroup

	r.ModelToState(ctx, group, &state)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r templateGroupResource) ModelToState(ctx context.Context, group *client.TemplateGroup, state *TemplateGroup) {
	state.ID = types.StringValue(group.Id)
	state.DefaultLocale = types.StringValue(group.DefaultLocale)
	state.CommSettings, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"email": types.ObjectType{},
		"sms":   types.ObjectType{},
		"ivr":   types.ObjectType{},
		"push":  types.ObjectType{},
	}, group.CommSettings)

	//state.EmailSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//	"service_setup_id":     types.StringType,
	//	"sender_name":          types.StringType,
	//	"sender_address":       types.StringType,
	//	"communication_method": types.StringType,
	//}, group.CommSettings.Email)
	//
	//state.SmsSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//	"service_setup_id":     types.StringType,
	//	"communication_method": types.StringType,
	//	"sender_name":          types.StringType,
	//	"sender_address":       types.StringType,
	//}, group.CommSettings.SMS)
	//
	//state.PushSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//	"service_setup_id":     types.StringType,
	//	"communication_method": types.StringType,
	//}, group.CommSettings.Push)
	//
	//state.IVRSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//	"service_setup_id":     types.StringType,
	//	"communication_method": types.StringType,
	//	"sender_address":       types.StringType,
	//}, group.CommSettings.IVR)
}
