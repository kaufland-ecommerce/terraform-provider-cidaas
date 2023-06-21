package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type hostedPageGroupResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*hostedPageGroupResource)(nil)

func NewHostedPageGroupResource() resource.Resource {
	return &hostedPageGroupResource{}
}

func (r *hostedPageGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hpgroup"
}

func (r *hostedPageGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *hostedPageGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Unique identifier of the hosted page group",
			},
			"created_time": schema.StringAttribute{
				Computed:    true,
				Description: "Time the hosted page was created",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_time": schema.StringAttribute{
				Computed:    true,
				Description: "Time the hosted page was last updated",
			},
			"default_locale": schema.StringAttribute{
				Required:    true,
				Description: "Default locale of the hosted page group",
			},
			"group_owner": schema.StringAttribute{
				Required:    true,
				Description: "Group owner of the hosted page group",
			},
			"hosted_pages": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of hosted pages",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "Unique identifier of the hosted page",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"consent_preview", "consent_scopes", "error", "login_success", "logout_success", "mfa_required", "password_forgot_init", "password_set_success", "register_additional_info", "register_success", "verification_init", "verification_complete", "password_set", "account_deduplication", "reactivate_verification_method", "device_init_code", "device_success_page", "status", "group_selection", "login", "register",
								),
							},
						},
						"content": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Description: "Content of the hosted page",
						},
						"locale": schema.StringAttribute{
							Required:    true,
							Description: "Locale of the hosted page",
						},
						"url": schema.StringAttribute{
							Optional:    true,
							Description: "URL of the hosted page",
						},
					},
				},
			},
		},
	}
}

func (r hostedPageGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan HostedPageGroup
	var plannedGroup client.HostedPageGroup

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedGroup.ID = plan.ID.ValueString()
	plannedGroup.DefaultLocale = plan.DefaultLocale.ValueString()
	plannedGroup.GroupOwner = plan.GroupOwner.ValueString()

	diags = plan.HostedPages.ElementsAs(ctx, &plannedGroup.HostedPages, true)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.provider.client.UpsertHostedPagesGroup(plannedGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hosted Pages group",
			"Could not create group, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := HostedPageGroup{
		ID:            types.StringValue(group.ID),
		GroupOwner:    types.StringValue(group.GroupOwner),
		CreatedTime:   types.StringValue(group.CreatedTime),
		UpdatedTime:   types.StringValue(group.UpdatedTime),
		DefaultLocale: types.StringValue(group.DefaultLocale),
	}

	state.HostedPages, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"content": types.StringType,
			"locale":  types.StringType,
			"url":     types.StringType,
		},
	}, group.HostedPages)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r hostedPageGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var groupId string

	diags := req.State.GetAttribute(ctx, path.Root("id"), &groupId)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.provider.client.GetHostedPagesGroup(groupId)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading hosted pages group",
			"Unexpected error fetching group: "+err.Error(),
		)
		return
	}

	state := HostedPageGroup{
		ID:            types.StringValue(group.ID),
		GroupOwner:    types.StringValue(group.GroupOwner),
		CreatedTime:   types.StringValue(group.CreatedTime),
		UpdatedTime:   types.StringValue(group.UpdatedTime),
		DefaultLocale: types.StringValue(group.DefaultLocale),
	}

	state.HostedPages, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"content": types.StringType,
			"locale":  types.StringType,
			"url":     types.StringType,
		},
	}, group.HostedPages)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)

}

func (r hostedPageGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan HostedPageGroup

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var plannedGroup client.HostedPageGroup

	plannedGroup.ID = plan.ID.ValueString()
	plannedGroup.GroupOwner = plan.GroupOwner.ValueString()
	plannedGroup.DefaultLocale = plan.DefaultLocale.ValueString()

	diags = plan.HostedPages.ElementsAs(ctx, &plannedGroup.HostedPages, true)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.provider.client.UpsertHostedPagesGroup(plannedGroup)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating hosted pages group",
			"Could not update group, unexpected error: "+err.Error(),
		)
		return
	}

	state := HostedPageGroup{
		ID:            types.StringValue(group.ID),
		GroupOwner:    types.StringValue(group.GroupOwner),
		CreatedTime:   types.StringValue(group.CreatedTime),
		UpdatedTime:   types.StringValue(group.UpdatedTime),
		DefaultLocale: types.StringValue(group.DefaultLocale),
	}

	state.HostedPages, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"content": types.StringType,
			"locale":  types.StringType,
			"url":     types.StringType,
		},
	}, group.HostedPages)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r hostedPageGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state HostedPageGroup

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.DeleteHostedPagesGroup(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Hosted Pages Group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
