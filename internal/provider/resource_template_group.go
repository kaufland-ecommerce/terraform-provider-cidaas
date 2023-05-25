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
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Cidaas UUID of the Template Group",
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "Unique Name of the Templat Group",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email_sender_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"from_name": schema.StringAttribute{
						Required:    true,
						Description: "Sender name for E-Mails",
					},
					"from_email": schema.StringAttribute{
						Required:    true,
						Description: "Sender address for E-Mails",
					},
					"provider": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "List of providers that should be used",
					},
				},
			},
			"sms_sender_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "SMS related sender settings",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"from_name": schema.StringAttribute{
						Required:    true,
						Description: "From name for SMS",
					},
					"provider": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "List of providers that should be used for sms communication",
					},
				},
			},
			"ivr_sender_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "IVR related settings",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"provider": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "List of providers that should be used for IVR",
					},
				},
			},
			"push_sender_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Push message related settings",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"provider": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "List of providers that should be used for Push",
					},
				},
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

	templateGroup, err := r.provider.client.CreateTemplateGroup(plan.GroupId.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template group",
			"Could not create hook, unexpected error: "+err.Error(),
		)
		return
	}

	result := TemplateGroup{
		GroupId: types.StringValue(templateGroup.GroupId),
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

	groupId := state.GroupId.ValueString()

	templateGroup, err := r.provider.client.GetTemplateGroup(groupId)

	if err != nil {
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
		Id:      plan.ID.ValueString(),
		GroupId: plan.GroupId.ValueString(),
	}

	diags = plan.EmailSenderConfig.As(ctx, &group.EmailSenderConfig, struct {
		UnhandledNullAsEmpty    bool
		UnhandledUnknownAsEmpty bool
	}{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	resp.Diagnostics.Append(diags...)

	diags = plan.SmsSenderConfig.As(ctx, &group.SmsSenderConfig, struct {
		UnhandledNullAsEmpty    bool
		UnhandledUnknownAsEmpty bool
	}{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	resp.Diagnostics.Append(diags...)

	diags = plan.IVRSenderConfig.As(ctx, &group.IVRSenderConfig, struct {
		UnhandledNullAsEmpty    bool
		UnhandledUnknownAsEmpty bool
	}{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	resp.Diagnostics.Append(diags...)

	diags = plan.PushSenderConfig.As(ctx, &group.PushSenderConfig, struct {
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

	err := r.provider.client.DeleteTemplateGroup(state.GroupId.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Template Group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r templateGroupResource) ModelToState(ctx context.Context, group *client.TemplateGroup, state *TemplateGroup) {
	state.ID = types.StringValue(group.Id)
	state.GroupId = types.StringValue(group.GroupId)
	state.EmailSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":         types.StringType,
		"from_name":  types.StringType,
		"from_email": types.StringType,
		"provider":   types.ListType{ElemType: types.StringType},
	}, group.EmailSenderConfig)

	state.SmsSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":        types.StringType,
		"from_name": types.StringType,
		"provider":  types.ListType{ElemType: types.StringType},
	}, group.SmsSenderConfig)

	state.PushSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":       types.StringType,
		"provider": types.ListType{ElemType: types.StringType},
	}, group.PushSenderConfig)

	state.IVRSenderConfig, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":       types.StringType,
		"provider": types.ListType{ElemType: types.StringType},
	}, group.IVRSenderConfig)
}
