package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type consentInstanceDataSource struct {
	provider *cidaasProvider
}

var _ datasource.DataSource = (*consentInstanceDataSource)(nil)

func NewConsentInstanceDataSource() datasource.DataSource {
	return &consentInstanceDataSource{}
}

func (d *consentInstanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consent_instance"
}

func (d *consentInstanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"consent_name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *consentInstanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (c consentInstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var name string
	var state ConsentInstance

	diags := req.Config.GetAttribute(ctx, path.Root("consent_name"), &name)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	consent, err := c.provider.client.GetConsentInstance(name)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch consent instance",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(consent.ID)
	state.ConsentName = types.StringValue(consent.ConsentName)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
