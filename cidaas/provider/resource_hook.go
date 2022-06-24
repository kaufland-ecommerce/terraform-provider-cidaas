package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/cidaas/client"
	"github.com/real-digital/terraform-provider-cidaas/cidaas/provider/validators"
)

type resourceHookType struct{}
type resourceHook struct {
	p provider
}

func (r resourceHookType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Description: "Unique identifier of the webhook",
			},
			"last_updated": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Time of the last update of the webhook",
			},
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The URL corresponding to the client's Webhook receiver",
			},
			"auth_type": {
				Type:        types.StringType,
				Required:    true,
				Description: "Authentication method that is used for the webhook",
				Validators: []tfsdk.AttributeValidator{
					validators.ValueInList([]string{
						"APIKEY", "TOTP", "CIDAAS_OAUTH2",
					}),
				},
			},
			"events": {
				Type:        types.ListType{ElemType: types.StringType},
				Required:    true,
				Description: "One or more webhook events which will trigger the webhook",
			},
			"apikey_details": {
				Required: false,
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"apikey": {
						Type:        types.StringType,
						Required:    true,
						Description: "apikey to measure and restrict access to the webhook",
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
							validators.ValueInList([]string{
								"query", "header",
							}),
						},
					},
				}),
			},
		},
	}, nil
}

func (r resourceHookType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceHook{
		p: *(p.(*provider)),
	}, nil
}

func (r resourceHook) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
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

	hook, err := r.p.client.UpsertHook(plannedHook)
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
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceHook) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state Hook
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hookID := state.ID.Value

	hook, err := r.p.client.GetHook(hookID)
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

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceHook) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if !r.p.configured {
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

	hook, err := r.p.client.UpsertHook(plannedHook)
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
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceHook) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	if !r.p.configured {
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

	err := r.p.client.DeleteHook(state.ID.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting hook",
			"Could not delete hook, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
