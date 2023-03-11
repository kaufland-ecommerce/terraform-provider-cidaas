package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type passwordPolicyDataSource struct {
	provider *cidaasProvider
}

var _ datasource.DataSource = (*passwordPolicyDataSource)(nil)

func NewPasswordPolicyDataSource() datasource.DataSource {
	return &passwordPolicyDataSource{}
}

func (d *passwordPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"policy_name": schema.StringAttribute{
				Required: true,
			},
			"lower_and_upper_case": schema.BoolAttribute{
				Computed: true,
			},
			"minimum_length": schema.Int64Attribute{
				Computed: true,
			},
			"no_of_digits": schema.Int64Attribute{
				Computed: true,
			},
			"no_of_special_chars": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *passwordPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (d *passwordPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (d passwordPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var name string
	var state PasswordPolicy

	diags := req.Config.GetAttribute(ctx, path.Root("policy_name"), &name)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.provider.client.GetPasswordPolicyByName(name)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch social provider",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(policy.ID)
	state.PolicyName = types.StringValue(policy.PolicyName)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
