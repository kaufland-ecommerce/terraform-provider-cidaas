package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (r *templateGroupResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "`cidaas_template_group` manages Template Groups in the tenant.\n\n",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Description: "Cidaas UUID of the Template Group",
			},
			"group_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "Unique Name of the Templat Group",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
		},
	}, nil
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

	templateGroup, err := r.provider.client.CreateTemplateGroup(plan.GroupId.Value)

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
			"Could not read hookID "+groupId+": "+err.Error(),
		)
		return
	}

	state.GroupId = types.StringValue(templateGroup.GroupId)

	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)
}

func (r templateGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("Not implemented")
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
