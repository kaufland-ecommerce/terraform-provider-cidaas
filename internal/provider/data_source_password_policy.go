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
			"password_policy": schema.SingleNestedAttribute{
				Optional:    true,
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
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "It will be used to fetch the password policy",
			},
			"lower_and_upper_case": schema.BoolAttribute{
				Optional:           true,
				DeprecationMessage: "This attribute is deprecated and will be removed in a future version. Use 'password_policy' instead.",
			},
			"minimum_length": schema.Int64Attribute{
				Optional:           true,
				DeprecationMessage: "This attribute is deprecated and will be removed in a future version. Use 'password_policy' instead.",
			},
			"no_of_digits": schema.Int64Attribute{
				Optional:           true,
				DeprecationMessage: "This attribute is deprecated and will be removed in a future version. Use 'password_policy' instead.",
			},
			"no_of_special_chars": schema.Int64Attribute{
				Optional:           true,
				DeprecationMessage: "This attribute is deprecated and will be removed in a future version. Use 'password_policy' instead.",
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
		resp.Diagnostics.AddError("Could not fetch policy: "+name,
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(policy.ID)
	state.PolicyName = types.StringValue(policy.PolicyName)
	state.PolicyProperties = PolicyProperties{
		BlockCompromised: types.BoolValue(policy.PolicyProperties.BlockCompromised),
		DenyUsageCount:   types.Int64Value(policy.PolicyProperties.DenyUsageCount),
		ChangeEnforcement: PolicyChangeEnforcement{
			ExpirationInDays:       types.Int64Value(policy.PolicyProperties.ChangeEnforcement.ExpirationInDays),
			NotifyUserBeforeInDays: types.Int64Value(policy.PolicyProperties.ChangeEnforcement.NotifyUserBeforeInDays),
		},
		StrengthRegexes: []types.String{},
	}
	for _, policyRegex := range policy.PolicyProperties.StrengthRegexes {
		state.PolicyProperties.StrengthRegexes = append(state.PolicyProperties.StrengthRegexes, types.StringValue(policyRegex))
	}

	state.MinimumLength = types.Int64Null()
	state.LowerAndUpperCase = types.BoolNull()
	state.NoOfDigits = types.Int64Null()
	state.NoOfSpecialChars = types.Int64Null()

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
