package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

var _ provider.Provider = (*cidaasProvider)(nil)
var _ provider.ProviderWithMetadata = (*cidaasProvider)(nil)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cidaasProvider{
			version: version,
		}
	}
}

type cidaasProvider struct {
	configured bool
	client     client.Client
	version    string
}

func (p *cidaasProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cidaas"
}

func (p *cidaasProvider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"client_id": {
				Type:      types.StringType,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"client_secret": {
				Type:      types.StringType,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
		},
	}, nil
}

type providerData struct {
	Host         types.String `tfsdk:"host"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *cidaasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, res *provider.ConfigureResponse) {
	var config providerData

	res.ResourceData = p
	res.DataSourceData = p

	diags := req.Config.Get(ctx, &config)
	res.Diagnostics.Append(diags...)

	if res.Diagnostics.HasError() {
		return
	}

	var host string
	if config.Host.IsUnknown() {
		res.Diagnostics.AddWarning(
			"unable to create client",
			"Cannot use unknown value as host",
		)
		return
	}

	if config.Host.IsNull() {
		host = os.Getenv("CIDAAS_HOST")
	} else {
		host = config.Host.ValueString()
	}

	var clientId string
	if config.ClientId.IsUnknown() {
		res.Diagnostics.AddWarning(
			"unable to create client",
			"Cannot use unknown value as client_id",
		)
		return
	}

	if config.ClientId.IsNull() {
		clientId = os.Getenv("CIDAAS_CLIENT_ID")
	} else {
		clientId = config.ClientId.ValueString()
	}

	var clientSecret string
	if config.ClientSecret.IsUnknown() {
		res.Diagnostics.AddWarning(
			"unable to create client",
			"Cannot use unknown value as client_secret",
		)
		return
	}

	if config.ClientSecret.IsNull() {
		clientSecret = os.Getenv("CIDAAS_CLIENT_SECRET")
	} else {
		clientSecret = config.ClientSecret.ValueString()
	}

	c, err := client.NewClient(&host, &clientId, &clientSecret)

	if err != nil {
		res.Diagnostics.AddError(
			"unable to create client",
			"unable to create internal client: \n\n"+err.Error(),
		)
	}

	p.client = c
	p.configured = true
}

func (p *cidaasProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAppResource,
		NewHookResource,
		NewHostedPageGroupResource,
		NewPasswordPolicyResource,
		NewRegistrationFieldResource,
	}
}

func (p *cidaasProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewConsentInstanceDataSource,
		NewPasswordPolicyDataSource,
		NewSocialProviderDataSource,
		NewTenantInfoDataSource,
	}
}

// toProvider can be used to cast a generic provider.Provider reference to this specific provider.
// This is ideally used in DataSourceType.NewDataSource and ResourceType.NewResource calls.
func toProvider(in any) (*cidaasProvider, diag.Diagnostics) {
	if in == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	p, ok := in.(*cidaasProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. "+
				"This is always a bug in the provider code and should be reported to the provider developers.", in,
			),
		)
		return nil, diags
	}

	return p, diags
}
