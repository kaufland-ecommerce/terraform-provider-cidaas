package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type resourceRegistrationFieldType struct{}
type resourceRegistrationField struct {
	p cidaasProvider
}

func (r resourceRegistrationFieldType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "`cidaas_hook` manages webhooks in the tenant.\n\n" +
			"Webhooks are triggered depending on the configured events.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Description: "Unique identifier of the hook",
			},
			"required": {
				Type:        types.BoolType,
				Required:    true,
				Description: "",
			},
			"enabled": {
				Type:        types.BoolType,
				Required:    true,
				Description: "",
			},
			"claimable": {
				Type:        types.BoolType,
				Required:    true,
				Description: "",
			},
			"read_only": {
				Type:        types.BoolType,
				Required:    true,
				Description: "",
			},
			"parent_group_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "Group the registration field belongs to",
			},
			"field_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "",
			},
			"consent_refs": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: false,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					listvalidator.SizeAtLeast(1),
				},
			},
			"data_type": {
				Type:        types.StringType,
				Required:    true,
				Description: "",
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(
						"CONSENT",
					),
				},
			},
			"order": {
				Type:        types.Int64Type,
				Required:    true,
				Description: "",
			},
		},
	}, nil
}

func (r resourceRegistrationFieldType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceRegistrationField{
		p: *(p.(*cidaasProvider)),
	}, nil
}

func (r resourceRegistrationField) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan RegistrationField
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedField := client.RegistrationField{
		FieldKey:      plan.FieldKey.Value,
		DataType:      plan.DataType.Value,
		Required:      plan.Required.Value,
		Enabled:       plan.Enabled.Value,
		ReadOnly:      plan.ReadOnly.Value,
		Claimable:     plan.Claimable.Value,
		ParentGroupID: plan.ParentGroupId.Value,
		Order:         plan.Order.Value,
	}

	tfsdk.ValueAs(ctx, plan.ConsentRefs, &plannedField.ConsentRefs)

	err := r.p.client.UpsertRegistrationField(&plannedField)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating registration field",
			"Could not update field, unexpected error: "+err.Error(),
		)
		return
	}

	var result RegistrationField
	diags = result.FromClient(ctx, &plannedField)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r resourceRegistrationField) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state RegistrationField
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fieldKey := state.FieldKey.Value

	field, err := r.p.client.GetRegistrationField(fieldKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading registration field",
			"Could not read registration field "+fieldKey+": "+err.Error(),
		)
		return
	}

	diags = state.FromClient(ctx, field)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceRegistrationField) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan RegistrationField
	var state RegistrationField

	req.State.Get(ctx, &state)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedField := client.RegistrationField{
		FieldKey:      plan.FieldKey.Value,
		DataType:      plan.DataType.Value,
		Required:      plan.Required.Value,
		Enabled:       plan.Enabled.Value,
		ReadOnly:      plan.ReadOnly.Value,
		Claimable:     plan.Claimable.Value,
		ParentGroupID: plan.ParentGroupId.Value,
		Order:         plan.Order.Value,
	}

	tfsdk.ValueAs(ctx, plan.ID, &plannedField.ID)
	tfsdk.ValueAs(ctx, plan.ConsentRefs, &plannedField.ConsentRefs)

	err := r.p.client.UpsertRegistrationField(&plannedField)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Registration Field",
			err.Error(),
		)
		return
	}

	var result RegistrationField
	diags = result.FromClient(ctx, &plannedField)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r resourceRegistrationField) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state RegistrationField
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.DeleteRegistrationField(state.FieldKey.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Registration Field",
			err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (field *RegistrationField) FromClient(ctx context.Context, crf *client.RegistrationField) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(tfsdk.ValueFrom(ctx, crf.ID, types.StringType, &field.ID)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.DataType, types.StringType, &field.DataType)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.FieldKey, types.StringType, &field.FieldKey)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.Required, types.BoolType, &field.Required)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.Enabled, types.BoolType, &field.Enabled)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.ReadOnly, types.BoolType, &field.ReadOnly)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.Claimable, types.BoolType, &field.Claimable)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.ParentGroupID, types.StringType, &field.ParentGroupId)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.Order, types.Int64Type, &field.Order)...)
	diags.Append(tfsdk.ValueFrom(ctx, crf.ConsentRefs, types.ListType{ElemType: types.StringType}, &field.ConsentRefs)...)

	return diags
}
