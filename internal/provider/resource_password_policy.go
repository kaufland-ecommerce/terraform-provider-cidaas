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
			"password_policy": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Password policy settings",
				Attributes: map[string]schema.Attribute{
					"block_compromised": schema.BoolAttribute{
						Required:    true,
						Description: "Block compromised passwords",
					},
					"strength_regexes": schema.ListAttribute{
						Required:    true,
						Description: "List of regexes to validate password strength",
						ElementType: types.StringType,
					},
					"deny_usage_count": schema.Int64Attribute{
						Required:    true,
						Description: "Number of times a password can be used before it is blocked, 0 means no limit",
					},
					"change_enforcement": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Change enforcement settings",
						Attributes: map[string]schema.Attribute{
							"expiration_in_days": schema.Int64Attribute{
								Required:    true,
								Description: "Number of days after which the password must be changed, 0 means no expiration",
							},
							"notify_user_before_in_days": schema.Int64Attribute{
								Required:    true,
								Description: "Number of days before expiration to notify the user, 0 means no notification",
							},
						},
					},
				},
			},
			"lower_and_upper_case": schema.BoolAttribute{
				Optional:           true,
				DeprecationMessage: "Deprecated, use `password_policy` instead",
				Description:        "Indicates if passwords are required to have lower and upper case letters",
			},
			"minimum_length": schema.Int64Attribute{
				Optional:           true,
				DeprecationMessage: "Deprecated, use `password_policy` instead",
				Description:        "Minimum length of the passwords",
			},
			"no_of_digits": schema.Int64Attribute{
				Optional:           true,
				DeprecationMessage: "Deprecated, use `password_policy` instead",
				Description:        "Number of digits that need to be included in the password",
			},
			"no_of_special_chars": schema.Int64Attribute{
				Optional:           true,
				DeprecationMessage: "Deprecated, use `password_policy` instead",
				Description:        "Number of special chars that need to be included in the password",
			},
		},
	}
}

func (r *passwordPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan PasswordPolicy

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedPolicy := client.CreatePolicyRequest{
		PolicyName: plan.PolicyName.ValueString(),
		PolicyProperties: client.PolicyProperties{
			BlockCompromised: plan.PolicyProperties.BlockCompromised.ValueBool(),
			DenyUsageCount:   plan.PolicyProperties.DenyUsageCount.ValueInt64(),
			ChangeEnforcement: client.PolicyChangeEnforcement{
				ExpirationInDays:       plan.PolicyProperties.ChangeEnforcement.ExpirationInDays.ValueInt64(),
				NotifyUserBeforeInDays: plan.PolicyProperties.ChangeEnforcement.NotifyUserBeforeInDays.ValueInt64(),
			},
		},
	}
	for i := range plan.PolicyProperties.StrengthRegexes {
		plannedPolicy.PolicyProperties.StrengthRegexes[i] = plan.PolicyProperties.StrengthRegexes[i].ValueString()
	}

	policy, err := r.provider.client.CreatePasswordPolicy(plannedPolicy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating password policy",
			"Could not create policy, unexpected error: "+err.Error(),
		)
		return
	}

	var state PasswordPolicy
	r.ResultToState(policy, &state)

	diags = resp.State.Set(ctx, state)
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
	if len(policyID) != 0 {
		policy, err := r.provider.client.GetPasswordPolicy(policyID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading password policy",
				"Could not read policy with id "+policyID+": "+err.Error(),
			)
			return
		}

		r.ResultToState(policy, &state)
	} else {
		policy, err := r.provider.client.GetPasswordPolicyByName(state.PolicyName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading password policy",
				"Could not read policy with name "+state.PolicyName.ValueString()+": "+err.Error(),
			)
			return
		}

		r.ResultToState(policy, &state)
	}

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
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
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
		ID:         state.ID.ValueString(),
		PolicyName: plan.PolicyName.ValueString(),
		PolicyProperties: client.PolicyProperties{
			BlockCompromised: plan.PolicyProperties.BlockCompromised.ValueBool(),
			DenyUsageCount:   plan.PolicyProperties.DenyUsageCount.ValueInt64(),
			ChangeEnforcement: client.PolicyChangeEnforcement{
				ExpirationInDays:       plan.PolicyProperties.ChangeEnforcement.ExpirationInDays.ValueInt64(),
				NotifyUserBeforeInDays: plan.PolicyProperties.ChangeEnforcement.NotifyUserBeforeInDays.ValueInt64(),
			},
		},
		// @Deprecated Remove in future
		MinimumLength:     plan.MinimumLength.ValueInt64(),
		NoOfDigits:        plan.NoOfDigits.ValueInt64(),
		LowerAndUpperCase: plan.LowerAndUpperCase.ValueBool(),
		NoOfSpecialChars:  plan.NoOfSpecialChars.ValueInt64(),
	}
	if len(plan.PolicyProperties.StrengthRegexes) > 0 {
		var policyStrengthRegexes []string
		for _, policy := range plan.PolicyProperties.StrengthRegexes {
			policyStrengthRegexes = append(policyStrengthRegexes, policy.ValueString())
		}
		plannedPolicy.PolicyProperties.StrengthRegexes = policyStrengthRegexes
	}

	policy, err := r.provider.client.UpdatePasswordPolicy(plannedPolicy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating password policy",
			"Could not update policy, unexpected error: "+err.Error(),
		)
		return
	}

	var result PasswordPolicy
	r.ResultToState(policy, &result)

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
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
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

func (r passwordPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policy, err := r.provider.client.GetPasswordPolicy(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch password policy",
			err.Error(),
		)
		return
	}
	var state PasswordPolicy

	r.ResultToState(policy, &state)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *passwordPolicyResource) ResultToState(policy *client.PasswordPolicy, state *PasswordPolicy) {
	state.ID = types.StringValue(policy.ID)
	state.PolicyName = types.StringValue(policy.PolicyName)

	r.setPolicyProperties(&policy.PolicyProperties, &state.PolicyProperties)
}

func (r *passwordPolicyResource) setPolicyProperties(clientProps *client.PolicyProperties, stateProps *PolicyProperties) {
	stateProps.BlockCompromised = types.BoolValue(clientProps.BlockCompromised)
	stateProps.DenyUsageCount = types.Int64Value(clientProps.DenyUsageCount)

	r.setChangeEnforcement(&clientProps.ChangeEnforcement, &stateProps.ChangeEnforcement)

	if len(clientProps.StrengthRegexes) > 0 {
		var policyStrengthRegexes []types.String
		for _, policy := range clientProps.StrengthRegexes {
			policyStrengthRegexes = append(policyStrengthRegexes, types.StringValue(policy))
		}
		stateProps.StrengthRegexes = policyStrengthRegexes

	}
}

func (r *passwordPolicyResource) setChangeEnforcement(clientCE *client.PolicyChangeEnforcement, stateCE *PolicyChangeEnforcement) {
	stateCE.ExpirationInDays = types.Int64Value(clientCE.ExpirationInDays)
	stateCE.NotifyUserBeforeInDays = types.Int64Value(clientCE.NotifyUserBeforeInDays)
}
