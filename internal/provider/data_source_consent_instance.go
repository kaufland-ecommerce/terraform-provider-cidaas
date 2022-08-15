package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type computeConsentInstanceDataSourceType struct{}
type computeConsentInstanceDataSource struct {
	client client.Client
}

func (c computeConsentInstanceDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"consent_name": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (c computeConsentInstanceDataSourceType) NewDataSource(_ context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	return computeConsentInstanceDataSource{
		client: p.(*cidaasProvider).client,
	}, nil
}

func (c computeConsentInstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var name string
	var state ConsentInstance

	diags := req.Config.GetAttribute(ctx, path.Root("consent_name"), &name)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	consent, err := c.client.GetConsentInstance(name)

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch consent instance",
			err.Error(),
		)
		return
	}

	state.ID.Value = consent.ID
	state.ConsentName.Value = consent.ConsentName

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
