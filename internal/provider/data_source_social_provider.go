package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type computeSocialProviderDataSourceType struct{}
type computeSocialProviderDataSource struct {
	client client.Client
}

func (c computeSocialProviderDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (c computeSocialProviderDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return computeSocialProviderDataSource{
		client: p.(*provider).client,
	}, nil
}

func (c computeSocialProviderDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var name string
	var state SocialProvider

	diags := req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("provider_name"), &name)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := c.client.GetSocialProvider(name)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch social provider",
			err.Error(),
		)
		return
	}

	state.SocialId.Value = provider.Id
	state.ProviderName.Value = provider.ProviderName
	state.ProviderType.Value = provider.ProviderType

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
