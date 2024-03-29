package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type passwordPolicyResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*passwordPolicyResource)(nil)

func NewPasswordPolicyResource() resource.Resource {
	return &passwordPolicyResource{}
}

func (r *passwordPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (r *passwordPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *passwordPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_password_policy` controls the password policies in the tenant",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Unique identifier of the policy",
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the policy",
			},
			"lower_and_upper_case": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates if passwords are required to have lower and upper case letters",
			},
			"minimum_length": schema.Int64Attribute{
				Required:    true,
				Description: "Minimum length of the passwords",
			},
			"no_of_digits": schema.Int64Attribute{
				Required:    true,
				Description: "Number of digits that need to be included in the password",
			},
			"no_of_special_chars": schema.Int64Attribute{
				Required:    true,
				Description: "Number of special chars that need to be included in the password",
			},
		},
	}
}

func (r *passwordPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan PasswordPolicy

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedPolicy := client.PasswordPolicy{
		PolicyName:        plan.PolicyName.ValueString(),
		MinimumLength:     plan.MinimumLength.ValueInt64(),
		NoOfDigits:        plan.NoOfDigits.ValueInt64(),
		LowerAndUpperCase: plan.LowerAndUpperCase.ValueBool(),
		NoOfSpecialChars:  plan.NoOfSpecialChars.ValueInt64(),
	}

	policy, err := r.provider.client.UpdatePasswordPolicy(plannedPolicy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating password policy",
			"Could not create policy, unexpected error: "+err.Error(),
		)
		return
	}

	result := PasswordPolicy{
		ID:                types.StringValue(policy.ID),
		PolicyName:        types.StringValue(policy.PolicyName),
		MinimumLength:     types.Int64Value(policy.MinimumLength),
		NoOfDigits:        types.Int64Value(policy.NoOfDigits),
		LowerAndUpperCase: types.BoolValue(policy.LowerAndUpperCase),
		NoOfSpecialChars:  types.Int64Value(policy.NoOfSpecialChars),
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r passwordPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PasswordPolicy
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := state.ID.ValueString()

	policy, err := r.provider.client.GetPasswordPolicy(policyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading password policy",
			"Could not read policy with id "+policyID+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(policy.ID)
	state.PolicyName = types.StringValue(policy.PolicyName)
	state.LowerAndUpperCase = types.BoolValue(policy.LowerAndUpperCase)
	state.MinimumLength = types.Int64Value(policy.MinimumLength)
	state.NoOfDigits = types.Int64Value(policy.NoOfDigits)
	state.NoOfSpecialChars = types.Int64Value(policy.NoOfSpecialChars)

	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r passwordPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan PasswordPolicy
	var state PasswordPolicy

	req.State.Get(ctx, &state)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedPolicy := client.PasswordPolicy{
		ID:                state.ID.ValueString(),
		PolicyName:        plan.PolicyName.ValueString(),
		MinimumLength:     plan.MinimumLength.ValueInt64(),
		NoOfDigits:        plan.NoOfDigits.ValueInt64(),
		LowerAndUpperCase: plan.LowerAndUpperCase.ValueBool(),
		NoOfSpecialChars:  plan.NoOfSpecialChars.ValueInt64(),
	}

	policy, err := r.provider.client.UpdatePasswordPolicy(plannedPolicy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating password policy",
			"Could not update policy, unexpected error: "+err.Error(),
		)
		return
	}

	result := PasswordPolicy{
		ID:                state.ID,
		PolicyName:        types.StringValue(policy.PolicyName),
		MinimumLength:     types.Int64Value(policy.MinimumLength),
		NoOfDigits:        types.Int64Value(policy.NoOfDigits),
		LowerAndUpperCase: types.BoolValue(policy.LowerAndUpperCase),
		NoOfSpecialChars:  types.Int64Value(policy.NoOfSpecialChars),
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r passwordPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state PasswordPolicy

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.DeletePasswordPolicy(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting password policy",
			"Could not delete policy, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
