package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type computeTenantInfoDataSourceType struct{}
type computeTenantInfoDataSource struct {
	client client.Client
}

func (c computeTenantInfoDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Information about the connected tenant.",
		Attributes: map[string]tfsdk.Attribute{
			"tenant_name": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Public visible name of the tenant",
			},
			"tenant_key": {
				Type:        types.StringType,
				Computed:    true,
				Description: "(internal) id of the tenant",
			},
			"custom_field_flatten": {
				Type:     types.BoolType,
				Computed: true,
			},
			"version_info": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Currently deployed version",
			},
		},
	}, nil
}

func (c computeTenantInfoDataSourceType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return computeTenantInfoDataSource{
		client: p.(*cidaasProvider).client,
	}, nil
}

func (c computeTenantInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantInfo

	info, err := c.client.GetTenantInfo()

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch social provider",
			err.Error(),
		)
		return
	}

	state.CustomFieldFlatten.Value = info.CustomFieldFlatten
	state.TenantName.Value = info.TenantName
	state.TenantKey.Value = info.TenantKey
	state.VersionInfo.Value = info.VersionInfo

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
