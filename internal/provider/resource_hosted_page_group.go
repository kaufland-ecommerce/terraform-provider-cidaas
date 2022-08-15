package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type resourceHostedPageGroupType struct{}
type resourceHostedPageGroup struct {
	p cidaasProvider
}

func (r resourceHostedPageGroupType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
				Description: "Unique identifier of the hook",
			},
			"pages": {
				Type:     types.MapType{ElemType: types.StringType},
				Required: true,
			},
		},
	}, nil
}

func (r resourceHostedPageGroupType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceHostedPageGroup{
		p: *(p.(*cidaasProvider)),
	}, nil
}

func (r resourceHostedPageGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
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

	plannedGroup.Name = plan.Name.Value

	diags = plan.Pages.ElementsAs(ctx, &plannedGroup.Pages, true)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.CreateHostedPagesGroup(plannedGroup)
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
}

func (r resourceHostedPageGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var groupName string

	diags := req.State.GetAttribute(ctx, path.Root("name"), &groupName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.p.client.GetHostedPagesGroup(groupName)

	tflog.Trace(ctx, "Done fetching hosted pages group", map[string]interface{}{
		"group": group.Name,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading hosted pages group",
			"Unexpected error fetching group: "+err.Error(),
		)
		return
	}

	state := HostedPageGroup{
		Name:  types.String{Value: group.Name},
		Pages: types.Map{Elems: map[string]attr.Value{}, ElemType: types.StringType},
	}

	for key, val := range group.Pages {
		state.Pages.Elems[key] = types.String{Value: val}
	}

	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)

}

func (r resourceHostedPageGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
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

	plannedGroup.Name = plan.Name.Value

	diags = plan.Pages.ElementsAs(ctx, &plannedGroup.Pages, true)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.UpdateHostedPagesGroup(plannedGroup)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating hosted pages group",
			"Could not update group, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceHostedPageGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
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

	err := r.p.client.DeleteHostedPagesGroup(state.Name.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Hosted Pages Group",
			"Could not delete group, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
