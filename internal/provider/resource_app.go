package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
	"golang.org/x/exp/slices"
)

type appResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*appResource)(nil)

func NewAppResource() resource.Resource {
	return &appResource{}
}

func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *appResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *appResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

			// Template Group ID
			"template_group_id": {
				Type:     types.StringType,
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
			"register_with_login_information": {
				Type:        types.BoolType,
				Required:    true,
				Description: "Should a login with social lead to account creation if not existing",
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

func (r appResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
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

func (r appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
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

	plannedApp := planToApp(ctx, &plan, &plan)

	app, err := r.provider.client.CreateApp(plannedApp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create app",
			err.Error(),
		)
		return
	}

	var state App

	diags = applyAppToState(ctx, &state, app)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.provider.configured {
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

	appID := state.ClientId.ValueString()

	app, err := r.provider.client.GetApp(appID)
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

	diags = applyAppToState(ctx, &state, app)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
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

	plannedApp := planToApp(ctx, &plan, &state)

	app, err := r.provider.client.UpdateApp(*plannedApp)

	if err != nil {
		resp.Diagnostics.AddError("Error Updating app", err.Error())
		return
	}

	diags = applyAppToState(ctx, &state, app)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
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

	err := r.provider.client.DeleteApp(state.ClientId.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error deleting app", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

func (r appResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state App

	tflog.Trace(ctx, "fetching app")

	app, err := r.provider.client.GetApp(req.ID)

	if err != nil {
		resp.Diagnostics.AddError("Error importing App", err.Error())
		return
	}

	applyAppToState(ctx, &state, app)

	if err != nil {
		resp.Diagnostics.AddError("Error importing app", err.Error())
		return
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func applyAppToState(ctx context.Context, state *App, app *client.App) diag.Diagnostics {
	ret := diag.Diagnostics{}

	var diags diag.Diagnostics

	state.ID = types.StringValue(app.ID)
	state.ClientId = types.StringValue(app.ClientId)
	state.ClientSecret = types.StringValue(app.ClientSecret)
	state.ClientName = types.StringValue(app.ClientName)
	state.ClientDisplayName = types.StringValue(app.ClientDisplayName)
	state.IsRememberMeSelected = types.BoolValue(app.IsRememberMeSelected)
	state.ClientType = types.StringValue(app.ClientType)
	state.AllowDisposableEmail = types.BoolValue(app.AllowDisposableEmail)
	state.FdsEnabled = types.BoolValue(app.FdsEnabled)
	state.EnablePasswordlessAuth = types.BoolValue(app.EnablePasswordlessAuth)
	state.EnableDeduplication = types.BoolValue(app.EnableDeduplication)
	state.MobileNumberVerificationRequired = types.BoolValue(app.MobileNumberVerificationRequired)
	state.HostedPageGroup = types.StringValue(app.HostedPageGroup)
	state.PrimaryColor = types.StringValue(app.PrimaryColor)
	state.AccentColor = types.StringValue(app.AccentColor)
	state.AutoLoginAfterRegister = types.BoolValue(app.AutoLoginAfterRegister)
	state.CompanyName = types.StringValue(app.CompanyName)
	state.CompanyAddress = types.StringValue(app.CompanyAddress)
	state.CompanyWebsite = types.StringValue(app.CompanyWebsite)
	state.TokenLifetimeInSeconds = types.Int64Value(app.TokenLifetimeInSeconds)
	state.IdTokenLifetimeInSeconds = types.Int64Value(app.IdTokenLifetimeInSeconds)
	state.RefreshTokenLifetimeInSeconds = types.Int64Value(app.RefreshTokenLifetimeInSeconds)
	state.EmailVerificationRequired = types.BoolValue(app.EmailVerificationRequired)
	state.EnableBotDetection = types.BoolValue(app.EnableBotDetection)
	state.IsLoginSuccessPageEnabled = types.BoolValue(app.IsLoginSuccessPageEnabled)
	state.JweEnabled = types.BoolValue(app.JweEnabled)
	state.AlwaysAskMfa = types.BoolValue(app.AlwaysAskMfa)

	tfsdk.ValueFrom(ctx, app.RegisterWithLoginInformation, types.BoolType, &state.RegisterWithLoginInformation)
	tfsdk.ValueFrom(ctx, app.PasswordPolicy, types.StringType, &state.PasswordPolicy)
	tfsdk.ValueFrom(ctx, app.TemplateGroupId, types.StringType, &state.TemplateGroupId)

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
			SocialId:     types.StringValue(item.SocialId),
			ProviderName: types.StringValue(item.ProviderName),
		})
	}

	state.AppKey, diags = types.ObjectValue(
		map[string]attr.Type{
			"id":          types.StringType,
			"private_key": types.StringType,
			"public_key":  types.StringType,
		},
		map[string]attr.Value{
			"id":          types.StringValue(app.AppKey.ID),
			"private_key": types.StringValue(app.AppKey.PrivateKey),
			"public_key":  types.StringValue(app.AppKey.PublicKey),
		},
	)

	ret.Append(diags...)

	return ret
}

func planToApp(ctx context.Context, plan *App, state *App) *client.App {
	plannedApp := client.App{
		ID:                               state.ID.ValueString(),
		ClientSecret:                     state.ClientSecret.ValueString(),
		ClientId:                         state.ClientId.ValueString(),
		ClientDisplayName:                plan.ClientDisplayName.ValueString(),
		ClientName:                       plan.ClientName.ValueString(),
		ClientType:                       plan.ClientType.ValueString(),
		IsRememberMeSelected:             plan.IsRememberMeSelected.ValueBool(),
		AllowDisposableEmail:             plan.AllowDisposableEmail.ValueBool(),
		AutoLoginAfterRegister:           plan.AutoLoginAfterRegister.ValueBool(),
		FdsEnabled:                       plan.FdsEnabled.ValueBool(),
		EnablePasswordlessAuth:           plan.EnablePasswordlessAuth.ValueBool(),
		EnableDeduplication:              plan.EnableDeduplication.ValueBool(),
		MobileNumberVerificationRequired: plan.MobileNumberVerificationRequired.ValueBool(),
		HostedPageGroup:                  plan.HostedPageGroup.ValueString(),
		PrimaryColor:                     plan.PrimaryColor.ValueString(),
		AccentColor:                      plan.AccentColor.ValueString(),
		CompanyName:                      plan.CompanyName.ValueString(),
		CompanyWebsite:                   plan.CompanyWebsite.ValueString(),
		CompanyAddress:                   plan.CompanyAddress.ValueString(),
		TokenLifetimeInSeconds:           plan.TokenLifetimeInSeconds.ValueInt64(),
		IdTokenLifetimeInSeconds:         plan.IdTokenLifetimeInSeconds.ValueInt64(),
		RefreshTokenLifetimeInSeconds:    plan.RefreshTokenLifetimeInSeconds.ValueInt64(),
		EmailVerificationRequired:        plan.EmailVerificationRequired.ValueBool(),
		EnableBotDetection:               plan.EnableBotDetection.ValueBool(),
		IsLoginSuccessPageEnabled:        plan.IsLoginSuccessPageEnabled.ValueBool(),
		JweEnabled:                       plan.JweEnabled.ValueBool(),
		AlwaysAskMfa:                     plan.AlwaysAskMfa.ValueBool(),
		RegisterWithLoginInformation:     plan.RegisterWithLoginInformation.ValueBool(),

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
				SocialId:     socialProvider.SocialId.ValueString(),
				ProviderName: socialProvider.ProviderName.ValueString(),
			},
		)
	}

	tfsdk.ValueAs(ctx, plan.TemplateGroupId, &plannedApp.TemplateGroupId)
	tfsdk.ValueAs(ctx, plan.PasswordPolicy, &plannedApp.PasswordPolicy)

	return &plannedApp
}
