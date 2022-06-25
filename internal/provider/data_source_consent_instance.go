package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func (c computeConsentInstanceDataSourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return computeConsentInstanceDataSource{
		client: p.(*provider).client,
	}, nil
}

func (c computeConsentInstanceDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var name string
	var state ConsentInstance

	diags := req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("consent_name"), &name)

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
