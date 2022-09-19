package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tenantInfoDataSource struct {
	provider *cidaasProvider
}

var _ datasource.DataSource = (*tenantInfoDataSource)(nil)

func NewTenantInfoDataSource() datasource.DataSource {
	return &tenantInfoDataSource{}
}

func (d *tenantInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_info"
}

func (d *tenantInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (c *tenantInfoDataSource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (c tenantInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantInfo

	info, err := c.provider.client.GetTenantInfo()

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
