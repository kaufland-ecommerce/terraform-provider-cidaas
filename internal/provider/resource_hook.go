package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type hookResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*hookResource)(nil)

func NewHookResource() resource.Resource {
	return &hookResource{}
}

func (r *hookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hook"
}

func (r *hookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *hookResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "`cidaas_hook` manages webhooks in the tenant.\n\n" +
			"Webhooks are triggered depending on the configured events.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Description: "Unique identifier of the hook",
			},
			"last_updated": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Time of the last update of the hook",
			},
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The URL corresponding to the client's Webhook receiver",
			},
			"auth_type": {
				Type:        types.StringType,
				Required:    true,
				Description: "Authentication method that is used for the hook",
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(
						"APIKEY", "TOTP", "CIDAAS_OAUTH2",
					),
				},
			},
			"events": {
				Type:        types.ListType{ElemType: types.StringType},
				Required:    true,
				Description: "One or more hook events which will trigger the hook",
			},
			"apikey_details": {
				Required: false,
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"apikey": {
						Type:        types.StringType,
						Required:    true,
						Description: "apikey to measure and restrict access to the hook",
					},
					"apikey_placeholder": {
						Type:        types.StringType,
						Required:    true,
						Description: "name of the attribute in which the apikey is to be provided",
					},
					"apikey_placement": {
						Type:        types.StringType,
						Required:    true,
						Description: "pass apikey as query param or header param",
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf(
								"query", "header",
							),
						},
					},
				}),
			},
		},
	}, nil
}

func (r hookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan Hook

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedHook := client.Hook{
		AuthType:    plan.AuthType,
		Events:      plan.Events,
		URL:         plan.Url,
		CreatedTime: "",
		UpdatedTime: "",
		ApiKeyDetails: client.HookApiKeyDetails{
			APIKey:            plan.APIKeyDetails.APIKey,
			APIKeyPlaceholder: plan.APIKeyDetails.APIKeyPlaceholder,
			APIKeyPlacement:   plan.APIKeyDetails.APIKeyPlacement,
		},
	}

	hook, err := r.provider.client.UpsertHook(plannedHook)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating hook",
			"Could not create hook, unexpected error: "+err.Error(),
		)
		return
	}

	result := Hook{
		ID:         types.String{Value: hook.Id},
		LastUpdate: types.String{Value: hook.UpdatedTime},
		Url:        hook.URL,
		AuthType:   hook.AuthType,
		Events:     hook.Events,
		APIKeyDetails: HookAPIKeyDetails{
			APIKey:            hook.ApiKeyDetails.APIKey,
			APIKeyPlaceholder: hook.ApiKeyDetails.APIKeyPlaceholder,
			APIKeyPlacement:   hook.ApiKeyDetails.APIKeyPlacement,
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r hookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Hook
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hookID := state.ID.Value

	hook, err := r.provider.client.GetHook(hookID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading hook",
			"Could not read hookID "+hookID+": "+err.Error(),
		)
		return
	}

	state.Url = hook.URL
	state.Events = hook.Events
	state.LastUpdate.Value = hook.UpdatedTime
	state.AuthType = hook.AuthType

	state.APIKeyDetails.APIKeyPlaceholder = hook.ApiKeyDetails.APIKeyPlaceholder
	state.APIKeyDetails.APIKeyPlacement = hook.ApiKeyDetails.APIKeyPlacement
	state.APIKeyDetails.APIKey = hook.ApiKeyDetails.APIKey

	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)
}

func (r hookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan Hook
	var state Hook

	req.State.Get(ctx, &state)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedHook := client.Hook{
		Id:       state.ID.Value,
		AuthType: plan.AuthType,
		Events:   plan.Events,
		URL:      plan.Url,
		ApiKeyDetails: client.HookApiKeyDetails{
			APIKey:            plan.APIKeyDetails.APIKey,
			APIKeyPlaceholder: plan.APIKeyDetails.APIKeyPlaceholder,
			APIKeyPlacement:   plan.APIKeyDetails.APIKeyPlacement,
		},
	}

	hook, err := r.provider.client.UpsertHook(plannedHook)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating hook",
			"Could not update hook, unexpected error: "+err.Error(),
		)
		return
	}

	result := Hook{
		ID:         types.String{Value: hook.Id},
		LastUpdate: types.String{Value: hook.UpdatedTime},
		Url:        hook.URL,
		AuthType:   hook.AuthType,
		Events:     hook.Events,
		APIKeyDetails: HookAPIKeyDetails{
			APIKey:            hook.ApiKeyDetails.APIKey,
			APIKeyPlaceholder: hook.ApiKeyDetails.APIKeyPlaceholder,
			APIKeyPlacement:   hook.ApiKeyDetails.APIKeyPlacement,
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r hookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state Hook

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.DeleteHook(state.ID.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting hook",
			"Could not delete hook, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
