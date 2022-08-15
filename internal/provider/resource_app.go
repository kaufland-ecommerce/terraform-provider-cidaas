package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
	"golang.org/x/exp/slices"
)

type resourceAppType struct{}
type resourceApp struct {
	p cidaasProvider
}

func (r resourceAppType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},

			// App Details
			"client_name": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"client_display_name": {
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtLeast(1),
				},
			},
			// TODO: App Logo URL
			// TODO: App Background URL
			"primary_color": {
				Type:     types.StringType,
				Optional: true,
			},
			"accent_color": {
				Type:     types.StringType,
				Optional: true,
			},
			"client_type": {
				Required: true,
				Type:     types.StringType,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(
						"SINGLE_PAGE", "ANDROID", "IOS", "REGULAR_WEB", "NON_INTERACTIVE",
					),
				},
			},

			// App Settings
			"client_id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"client_secret": {
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"allowed_scopes": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},
			"redirect_uris": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},
			"allowed_logout_urls": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},

			// Company Details
			"company_name": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"company_address": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"company_website": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthAtLeast(1),
				},
			},

			// OAuth Settings
			"response_types": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},
			"grant_types": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},
			"allowed_web_origins": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},
			"allowed_origins": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},

			// Token Settings
			"additional_access_token_payload": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},
			"token_lifetime_in_seconds": {
				Type:       types.Int64Type,
				Required:   true,
				Validators: []tfsdk.AttributeValidator{
					// validators.AtLeast(0),
				},
			},
			"id_token_lifetime_in_seconds": {
				Type:     types.Int64Type,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					int64validator.AtLeast(0),
				},
			},
			"refresh_token_lifetime_in_seconds": {
				Type:     types.Int64Type,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					int64validator.AtLeast(0),
				},
			},

			// Consent management
			"consent_refs": {
				Type:     types.ListType{ElemType: types.StringType},
				Required: true,
			},

			// Login Provider
			"social_providers": {
				Required: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"social_id": {
						Type:     types.StringType,
						Required: true,
					},
					"provider_name": {
						Type:     types.StringType,
						Required: true,
					},
					"provider_type": {
						Type:     types.StringType,
						Computed: true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							resource.UseStateForUnknown(),
						},
					},
				}),
			},

			// Guest Login
			"allow_guest_login": {
				Type:     types.BoolType,
				Required: true,
			},

			// TODO: Guest login groups

			// Registration Fields
			"allowed_fields": {
				Type:     types.ListType{ElemType: types.StringType},
				Optional: true,
			},
			"required_fields": {
				Type:     types.ListType{ElemType: types.StringType},
				Optional: true,
			},
			"email_verification_required": {
				Type:     types.BoolType,
				Required: true,
			},
			"mobile_number_verification_required": {
				Type:     types.BoolType,
				Required: true,
			},

			// Captcha
			// TODO

			// Password Rules
			"password_policy": {
				Type:     types.StringType,
				Optional: true,
				Required: false,
			},

			// Template Group
			"hosted_page_group": {
				Type:     types.StringType,
				Required: true,
			},

			// Bot Detection
			"enable_bot_detection": {
				Type:     types.BoolType,
				Required: true,
			},

			// Authentication
			"always_ask_mfa": {
				Type:     types.BoolType,
				Required: true,
			},
			"allowed_mfa": {
				Type:     types.ListType{ElemType: types.StringType},
				Optional: true,
			},

			// Remember Me
			"is_remember_me_selected": {
				Type:     types.BoolType,
				Required: true,
			},

			// Success Page
			"is_login_success_page_enabled": {
				Type:     types.BoolType,
				Required: true,
			},

			// Groupes & Roles
			// TODO

			// Encryption Settings
			"jwe_enabled": {
				Type:     types.BoolType,
				Required: true,
			},

			// Certificates
			"app_key": {
				Type: types.ObjectType{AttrTypes: map[string]attr.Type{
					"id":          types.StringType,
					"private_key": types.StringType,
					"public_key":  types.StringType,
				}},
				Computed:  true,
				Sensitive: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},

			// Flow Settings
			"auto_login_after_register": {
				Type:        types.BoolType,
				Required:    true,
				Description: "If set, customers will be logged in directly after registrtion",
			},
			"allow_login_with": {
				Type:        types.ListType{ElemType: types.StringType},
				Required:    true,
				Description: "Profile information that can be used to login",
			},
			"fds_enabled": {
				Type:     types.BoolType,
				Required: true,
			},
			"enable_passwordless_auth": {
				Type:     types.BoolType,
				Required: true,
			},
			"enable_deduplication": {
				Type:     types.BoolType,
				Required: true,
			},
			"allow_disposable_email": {
				Type:        types.BoolType,
				Required:    true,
				Description: "If set, emails generated by throwaway email providers can be used for signup",
			},
		},
	}, nil
}

func (r resourceAppType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return resourceApp{
		p: *(p.(*cidaasProvider)),
	}, nil
}

func (r resourceApp) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var allowedFields []string
	var requiredFields []string

	req.Config.GetAttribute(ctx, path.Root("allowed_fields"), &allowedFields)
	req.Config.GetAttribute(ctx, path.Root("required_fields"), &requiredFields)

	for _, el := range requiredFields {
		if !slices.Contains(allowedFields, el) {
			resp.Diagnostics.AddError(
				"Required field not in list of allowed fields",
				fmt.Sprintf("%s is not in the list of allowed fileds and can therefore not be required", el),
			)
		}
	}
}

func (r resourceApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan App

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedApp, _ := planToApp(ctx, &plan, &plan)

	app, err := r.p.client.CreateApp(plannedApp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create app",
			err.Error(),
		)
		return
	}

	var state App

	err = applyAppToState(ctx, &state, app)

	if err != nil {
		resp.Diagnostics.AddError("Error Updating app", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state App
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ClientId.Value

	app, err := r.p.client.GetApp(appID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading app",
			err.Error(),
		)
		return
	}

	if app == nil {
		return
	}

	err = applyAppToState(ctx, &state, app)

	if err != nil {
		resp.Diagnostics.AddError(
			"error reading app",
			err.Error(),
		)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceApp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan App
	var state App

	req.State.Get(ctx, &state)
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)

	plannedApp, _ := planToApp(ctx, &plan, &state)

	app, err := r.p.client.UpdateApp(*plannedApp)

	if err != nil {
		resp.Diagnostics.AddError("Error Updating app", err.Error())
		return
	}

	err = applyAppToState(ctx, &state, app)

	if err != nil {
		resp.Diagnostics.AddError("Error Updating app", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state App

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.DeleteApp(state.ClientId.Value)

	if err != nil {
		resp.Diagnostics.AddError("Error deleting app", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

func (r resourceApp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state App

	tflog.Trace(ctx, "fetching app")

	app, err := r.p.client.GetApp(req.ID)

	if err != nil {
		resp.Diagnostics.AddError("Error importing App", err.Error())
		return
	}

	err = applyAppToState(ctx, &state, app)

	if err != nil {
		resp.Diagnostics.AddError("Error importing app", err.Error())
		return
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func applyAppToState(ctx context.Context, state *App, app *client.App) error {
	state.ID.Value = app.ID
	state.ClientId.Value = app.ClientId
	state.ClientSecret.Value = app.ClientSecret
	state.ClientName.Value = app.ClientName
	state.ClientDisplayName.Value = app.ClientDisplayName
	state.IsRememberMeSelected.Value = app.IsRememberMeSelected
	state.ClientType.Value = app.ClientType
	state.AllowDisposableEmail.Value = app.AllowDisposableEmail
	state.FdsEnabled.Value = app.FdsEnabled
	state.EnablePasswordlessAuth.Value = app.EnablePasswordlessAuth
	state.EnableDeduplication.Value = app.EnableDeduplication
	state.MobileNumberVerificationRequired.Value = app.MobileNumberVerificationRequired
	state.HostedPageGroup.Value = app.HostedPageGroup
	state.PrimaryColor.Value = app.PrimaryColor
	state.AccentColor.Value = app.AccentColor
	state.AutoLoginAfterRegister.Value = app.AutoLoginAfterRegister
	state.CompanyName.Value = app.CompanyName
	state.CompanyAddress.Value = app.CompanyAddress
	state.CompanyWebsite.Value = app.CompanyWebsite
	state.TokenLifetimeInSeconds.Value = app.TokenLifetimeInSeconds
	state.IdTokenLifetimeInSeconds.Value = app.IdTokenLifetimeInSeconds
	state.RefreshTokenLifetimeInSeconds.Value = app.RefreshTokenLifetimeInSeconds
	state.EmailVerificationRequired.Value = app.EmailVerificationRequired
	state.EnableBotDetection.Value = app.EnableBotDetection
	state.IsLoginSuccessPageEnabled.Value = app.IsLoginSuccessPageEnabled
	state.JweEnabled.Value = app.JweEnabled
	state.AlwaysAskMfa.Value = app.AlwaysAskMfa

	tfsdk.ValueFrom(ctx, app.PasswordPolicy, types.StringType, &state.PasswordPolicy)

	state.Scopes = app.AllowedScopes
	state.RedirectUris = app.RedirectUris
	state.AllowedLogoutUrls = app.AllowedLogoutUrls
	state.AdditionalAccessTokenPayload = app.AdditionalAccessTokenPayload
	state.AllowLoginWith = app.AllowLoginWith
	state.AllowedFields = app.AllowedFields
	state.RequiredFields = app.RequiredFields
	state.ConsentRefs = app.ConsentRefs
	state.ResponseTypes = app.ResponseTypes
	state.GrantTypes = app.GrantTypes
	state.AllowedWebOrigins = app.AllowedWebOrigins
	state.AllowedOrigins = app.AllowedOrigins
	state.AllowedMfa = app.AllowedMfa

	state.SocialProviders = []SocialProvider{}
	for _, item := range app.SocialProviders {
		state.SocialProviders = append(state.SocialProviders, SocialProvider{
			SocialId:     types.String{Value: item.SocialId},
			ProviderName: types.String{Value: item.ProviderName},
		})
	}

	state.AppKey.AttrTypes = map[string]attr.Type{
		"id":          types.StringType,
		"private_key": types.StringType,
		"public_key":  types.StringType,
	}

	state.AppKey.Attrs = map[string]attr.Value{
		"id":          types.String{Value: app.AppKey.ID},
		"private_key": types.String{Value: app.AppKey.PrivateKey},
		"public_key":  types.String{Value: app.AppKey.PublicKey},
	}

	return nil
}

func planToApp(ctx context.Context, plan *App, state *App) (*client.App, error) {
	plannedApp := client.App{
		ID:                               state.ID.Value,
		ClientSecret:                     state.ClientSecret.Value,
		ClientId:                         state.ClientId.Value,
		ClientDisplayName:                plan.ClientDisplayName.Value,
		ClientName:                       plan.ClientName.Value,
		ClientType:                       plan.ClientType.Value,
		IsRememberMeSelected:             plan.IsRememberMeSelected.Value,
		AllowDisposableEmail:             plan.AllowDisposableEmail.Value,
		AutoLoginAfterRegister:           plan.AutoLoginAfterRegister.Value,
		FdsEnabled:                       plan.FdsEnabled.Value,
		EnablePasswordlessAuth:           plan.EnablePasswordlessAuth.Value,
		EnableDeduplication:              plan.EnableDeduplication.Value,
		MobileNumberVerificationRequired: plan.MobileNumberVerificationRequired.Value,
		HostedPageGroup:                  plan.HostedPageGroup.Value,
		PrimaryColor:                     plan.PrimaryColor.Value,
		AccentColor:                      plan.AccentColor.Value,
		CompanyName:                      plan.CompanyName.Value,
		CompanyWebsite:                   plan.CompanyWebsite.Value,
		CompanyAddress:                   plan.CompanyAddress.Value,
		TokenLifetimeInSeconds:           plan.TokenLifetimeInSeconds.Value,
		IdTokenLifetimeInSeconds:         plan.IdTokenLifetimeInSeconds.Value,
		RefreshTokenLifetimeInSeconds:    plan.RefreshTokenLifetimeInSeconds.Value,
		EmailVerificationRequired:        plan.EmailVerificationRequired.Value,
		EnableBotDetection:               plan.EnableBotDetection.Value,
		IsLoginSuccessPageEnabled:        plan.IsLoginSuccessPageEnabled.Value,
		JweEnabled:                       plan.JweEnabled.Value,
		AlwaysAskMfa:                     plan.AlwaysAskMfa.Value,

		AllowLoginWith:               plan.AllowLoginWith,
		RedirectUris:                 plan.RedirectUris,
		AllowedLogoutUrls:            plan.AllowedLogoutUrls,
		AllowedScopes:                plan.Scopes,
		AdditionalAccessTokenPayload: plan.AdditionalAccessTokenPayload,
		AllowedFields:                plan.AllowedFields,
		RequiredFields:               plan.RequiredFields,
		ConsentRefs:                  plan.ConsentRefs,
		ResponseTypes:                plan.ResponseTypes,
		GrantTypes:                   plan.GrantTypes,
		AllowedWebOrigins:            plan.AllowedWebOrigins,
		AllowedOrigins:               plan.AllowedOrigins,
		AllowedMfa:                   plan.AllowedMfa,

		SocialProviders: []client.SocialProvider{},
	}

	for _, socialProvider := range plan.SocialProviders {
		plannedApp.SocialProviders = append(
			plannedApp.SocialProviders,
			client.SocialProvider{
				SocialId:     socialProvider.SocialId.Value,
				ProviderName: socialProvider.ProviderName.Value,
			},
		)
	}

	tfsdk.ValueAs(ctx, plan.PasswordPolicy, &plannedApp.PasswordPolicy)

	return &plannedApp, nil
}
