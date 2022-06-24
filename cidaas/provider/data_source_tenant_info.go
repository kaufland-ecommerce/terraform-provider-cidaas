package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/cidaas/client"
)

type computeTenantInfoDataSourceType struct{}
type computeTenantInfoDataSource struct {
	client client.Client
}

func (c computeTenantInfoDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"tenant_name": {
				Type:     types.StringType,
				Computed: true,
			},
			"tenant_key": {
				Type:     types.StringType,
				Computed: true,
			},
			"custom_field_flatten": {
				Type:     types.BoolType,
				Computed: true,
			},
			"version_info": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (c computeTenantInfoDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return computeTenantInfoDataSource{
		client: p.(*provider).client,
	}, nil
}

func (c computeTenantInfoDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
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
