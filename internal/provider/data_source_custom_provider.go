package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type customProviderDataSource struct {
	provider *cidaasProvider
}

var _ datasource.DataSource = (*customProviderDataSource)(nil)

func NewCustomProviderDataSource() datasource.DataSource {
	return &customProviderDataSource{}
}

func (d *customProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Allows reading custom login providers that are configured",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Computed: true,
			},
			"provider_name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *customProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider"
}

func (d *customProviderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (d customProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var providerName string

	var state CustomProvider

	diags := req.Config.GetAttribute(ctx, path.Root("provider_name"), &providerName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	customProvider, err := d.provider.client.GetCustomProvider(providerName)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch custom provider",
			err.Error(),
		)
		return
	}

	state.ProviderName = types.StringValue(customProvider.ProviderName)
	state.DisplayName = types.StringValue(customProvider.DisplayName)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
