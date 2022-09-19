package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type socialProviderDataSource struct {
	provider *cidaasProvider
}

var _ datasource.DataSource = (*socialProviderDataSource)(nil)

func NewSocialProviderDataSource() datasource.DataSource {
	return &socialProviderDataSource{}
}

func (d *socialProviderDataSource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Allows reading social providers that are configured in internal",
		Attributes: map[string]tfsdk.Attribute{
			"social_id": {
				Type:     types.StringType,
				Computed: true,
			},
			"provider_type": {
				Type:     types.StringType,
				Computed: true,
			},
			"provider_name": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (d *socialProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_social_provider"
}

func (d *socialProviderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (d socialProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var name string
	var state SocialProvider

	diags := req.Config.GetAttribute(ctx, path.Root("provider_name"), &name)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	socialProvider, err := d.provider.client.GetSocialProvider(name)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch social socialProvider",
			err.Error(),
		)
		return
	}

	state.SocialId.Value = socialProvider.Id
	state.ProviderName.Value = socialProvider.ProviderName
	state.ProviderType.Value = socialProvider.ProviderType

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
