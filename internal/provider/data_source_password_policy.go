package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type computePasswordPolicyDataSourceType struct{}
type computePasswordPolicyDataSource struct {
	client client.Client
}

func (c computePasswordPolicyDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"policy_name": {
				Type:     types.StringType,
				Required: true,
			},
			"lower_and_upper_case": {
				Type:     types.BoolType,
				Computed: true,
			},
			"minimum_length": {
				Type:     types.Int64Type,
				Computed: true,
			},
			"no_of_digits": {
				Type:     types.Int64Type,
				Computed: true,
			},
			"no_of_special_chars": {
				Type:     types.Int64Type,
				Computed: true,
			},
		},
	}, nil
}

func (c computePasswordPolicyDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return computePasswordPolicyDataSource{
		client: p.(*provider).client,
	}, nil
}

func (c computePasswordPolicyDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var name string
	var state PasswordPolicy

	diags := req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("policy_name"), &name)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := c.client.GetPasswordPolicyByName(name)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch social provider",
			err.Error(),
		)
		return
	}

	state.ID.Value = policy.ID
	state.PolicyName.Value = policy.PolicyName

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
