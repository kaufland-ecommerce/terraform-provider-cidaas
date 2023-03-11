package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type resourceRegistrationField struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*resourceRegistrationField)(nil)

func NewRegistrationFieldResource() resource.Resource {
	return &resourceRegistrationField{}
}

func (r *resourceRegistrationField) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registration_field"
}

func (r *resourceRegistrationField) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *resourceRegistrationField) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_registration_field` manages registration fields in the tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Unique identifier of the registration field",
			},
			"required": schema.StringAttribute{
				Required:    true,
				Description: "",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "",
			},
			"claimable": schema.BoolAttribute{
				Required:    true,
				Description: "",
			},
			"read_only": schema.BoolAttribute{
				Required:    true,
				Description: "",
			},
			"parent_group_id": schema.StringAttribute{
				Required:    true,
				Description: "Group the registration field belongs to",
			},
			"field_key": schema.StringAttribute{
				Required:    true,
				Description: "",
			},
			"consent_refs": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"data_type": schema.StringAttribute{
				Required:    true,
				Description: "",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"CONSENT",
					),
				},
			},
			"order": schema.Int64Attribute{
				Required:    true,
				Description: "",
			},
		},
	}
}

func (r resourceRegistrationField) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
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
		FieldKey:      plan.FieldKey.ValueString(),
		DataType:      plan.DataType.ValueString(),
		Required:      plan.Required.ValueBool(),
		Enabled:       plan.Enabled.ValueBool(),
		ReadOnly:      plan.ReadOnly.ValueBool(),
		Claimable:     plan.Claimable.ValueBool(),
		ParentGroupID: plan.ParentGroupId.ValueString(),
		Order:         plan.Order.ValueInt64(),
	}

	tfsdk.ValueAs(ctx, plan.ConsentRefs, &plannedField.ConsentRefs)

	err := r.provider.client.UpsertRegistrationField(&plannedField)
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
	if !r.provider.configured {
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

	fieldKey := state.FieldKey.ValueString()

	field, err := r.provider.client.GetRegistrationField(fieldKey)
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
	if !r.provider.configured {
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
		FieldKey:      plan.FieldKey.ValueString(),
		DataType:      plan.DataType.ValueString(),
		Required:      plan.Required.ValueBool(),
		Enabled:       plan.Enabled.ValueBool(),
		ReadOnly:      plan.ReadOnly.ValueBool(),
		Claimable:     plan.Claimable.ValueBool(),
		ParentGroupID: plan.ParentGroupId.ValueString(),
		Order:         plan.Order.ValueInt64(),
	}

	tfsdk.ValueAs(ctx, plan.ID, &plannedField.ID)
	tfsdk.ValueAs(ctx, plan.ConsentRefs, &plannedField.ConsentRefs)

	err := r.provider.client.UpsertRegistrationField(&plannedField)

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
	if !r.provider.configured {
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

	err := r.provider.client.DeleteRegistrationField(state.FieldKey.ValueString())
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
