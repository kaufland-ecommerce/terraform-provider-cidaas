package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (c *tenantInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Information about the connected tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_name": schema.StringAttribute{
				Computed:    true,
				Description: "Public visible name of the tenant",
			},
			"tenant_key": schema.StringAttribute{
				Computed:    true,
				Description: "(internal) id of the tenant",
			},
			"custom_field_flatten": schema.BoolAttribute{
				Computed: true,
			},
			"version_info": schema.StringAttribute{
				Computed:    true,
				Description: "Currently deployed version",
			},
		},
	}
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

	state.CustomFieldFlatten = types.BoolValue(info.CustomFieldFlatten)
	state.VersionInfo = types.StringValue(info.VersionInfo)

	state.TenantName = types.StringValue(info.TenantName)
	state.TenantKey = types.StringValue(info.TenantKey)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
