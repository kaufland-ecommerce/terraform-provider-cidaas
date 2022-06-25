package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

var _ tfsdk.Provider = &provider{}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

type provider struct {
	configured bool
	client     client.Client
	version    string
}

func (p *provider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, res *tfsdk.ConfigureProviderResponse) {
	var config providerData

	diags := req.Config.Get(ctx, &config)
	res.Diagnostics.Append(diags...)

	if res.Diagnostics.HasError() {
		return
	}

	var host string
	if config.Host.Unknown {
		res.Diagnostics.AddWarning(
			"unable to create client",
			"Cannot use unknown value as host",
		)
		return
	}

	if config.Host.Null {
		host = os.Getenv("CIDAAS_HOST")
	} else {
		host = config.Host.Value
	}

	var clientId string
	if config.ClientId.Unknown {
		res.Diagnostics.AddWarning(
			"unable to create client",
			"Cannot use unknown value as client_id",
		)
		return
	}

	if config.ClientId.Null {
		clientId = os.Getenv("CIDAAS_CLIENT_ID")
	} else {
		clientId = config.ClientId.Value
	}

	var clientSecret string
	if config.ClientSecret.Unknown {
		res.Diagnostics.AddWarning(
			"unable to create client",
			"Cannot use unknown value as client_secret",
		)
		return
	}

	if config.ClientSecret.Null {
		clientSecret = os.Getenv("CIDAAS_CLIENT_SECRET")
	} else {
		clientSecret = config.ClientSecret.Value
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

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"cidaas_hook":              resourceHookType{},
		"cidaas_app":               resourceAppType{},
		"cidaas_password_policy":   resourcePasswordPolicyType{},
		"cidaas_hosted_page_group": resourceHostedPageGroupType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"cidaas_social_provider":  computeSocialProviderDataSourceType{},
		"cidaas_consent_instance": computeConsentInstanceDataSourceType{},
		"cidaas_password_policy":  computePasswordPolicyDataSourceType{},
		"cidaas_tenant_info":      computeTenantInfoDataSourceType{},
	}, nil
}
