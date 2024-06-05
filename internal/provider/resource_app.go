package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"app_owner": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"bot_provider": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// App Details
			"client_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"client_display_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"primary_color": schema.StringAttribute{
				Optional: true,
			},
			"accent_color": schema.StringAttribute{
				Optional: true,
			},
			"client_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"SINGLE_PAGE", "ANDROID", "IOS", "REGULAR_WEB", "NON_INTERACTIVE",
					),
				},
			},

			// App Settings
			"client_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_scopes": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"redirect_uris": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"allowed_logout_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Company Details
			"company_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"company_address": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"company_website": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},

			// OAuth Settings
			"response_types": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"grant_types": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"allowed_web_origins": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"allowed_origins": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},

			// Token Settings
			"additional_access_token_payload": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"token_lifetime_in_seconds": schema.Int64Attribute{
				Required:   true,
				Validators: []validator.Int64{
					// validators.AtLeast(0),
				},
			},
			"id_token_lifetime_in_seconds": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"refresh_token_lifetime_in_seconds": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},

			// Consent management
			"consent_refs": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},

			// Template Group ID
			"template_group_id": schema.StringAttribute{
				Required: true,
			},

			"custom_providers": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{
							Required: true,
						},
						"provider_name": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},

			// Login Provider
			"social_providers": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"social_id": schema.StringAttribute{
							Required: true,
						},
						"provider_name": schema.StringAttribute{
							Required: true,
						},
						"provider_type": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},

			// Guest Login
			"allow_guest_login": schema.BoolAttribute{
				Required: true,
			},

			// TODO: Guest login groups

			// Registration Fields
			"allowed_fields": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"required_fields": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"email_verification_required": schema.BoolAttribute{
				Required: true,
			},
			"mobile_number_verification_required": schema.BoolAttribute{
				Required: true,
			},

			// Captcha
			// TODO

			// Password Rules
			"password_policy": schema.StringAttribute{
				Optional: true,
			},

			// Template Group
			"hosted_page_group": schema.StringAttribute{
				Required: true,
			},

			// Bot Detection
			"enable_bot_detection": schema.BoolAttribute{
				Required: true,
			},

			// Authentication
			"always_ask_mfa": schema.BoolAttribute{
				Required: true,
			},
			"allowed_mfa": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},

			// Remember Me
			"is_remember_me_selected": schema.BoolAttribute{
				Required: true,
			},

			// Success Page
			"is_login_success_page_enabled": schema.BoolAttribute{
				Required: true,
			},

			// Groups & Roles
			"allowed_groups": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Required: true,
						},
						"roles": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
						"default_roles": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"accept_roles_in_the_registration": schema.BoolAttribute{
				Required: true,
			},
			"operations_allowed_groups": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Required: true,
						},
						"roles": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
						"default_roles": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			// Encryption Settings
			"jwe_enabled": schema.BoolAttribute{
				Required: true,
			},

			// Certificates
			"app_key": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"id":          types.StringType,
					"private_key": types.StringType,
					"public_key":  types.StringType,
				},
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},

			// Flow Settings
			"auto_login_after_register": schema.BoolAttribute{
				Required:    true,
				Description: "If set, customers will be logged in directly after registrtion",
			},
			"allow_login_with": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Profile information that can be used to login",
			},
			"register_with_login_information": schema.BoolAttribute{
				Required:    true,
				Description: "Should a login with social lead to account creation if not existing",
			},
			"fds_enabled": schema.BoolAttribute{
				Required: true,
			},
			"enable_passwordless_auth": schema.BoolAttribute{
				Required: true,
			},
			"enable_deduplication": schema.BoolAttribute{
				Required: true,
			},
			"allow_disposable_email": schema.BoolAttribute{
				Required:    true,
				Description: "If set, emails generated by throwaway email providers can be used for signup",
			},
		},
	}
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

	plannedApp, diags := planToApp(ctx, &plan, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

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

	plannedApp, diags := planToApp(ctx, &plan, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	state.BotProvider = types.StringValue(app.BotProvider)
	state.AppOwner = types.StringValue(app.AppOwner)
	state.ClientId = types.StringValue(app.ClientId)
	state.ClientSecret = types.StringValue(app.ClientSecret)
	state.ClientName = types.StringValue(app.ClientName)
	state.ClientDisplayName = types.StringValue(app.ClientDisplayName)
	state.IsRememberMeSelected = types.BoolValue(app.IsRememberMeSelected)
	state.ClientType = types.StringValue(app.ClientType)
	state.AllowDisposableEmail = types.BoolValue(app.AllowDisposableEmail)
	state.AllowGuestLogin = types.BoolValue(app.AllowGuestLogin)
	state.FdsEnabled = types.BoolValue(app.FdsEnabled)
	state.EnablePasswordlessAuth = types.BoolValue(app.EnablePasswordlessAuth)
	state.EnableDeduplication = types.BoolValue(app.EnableDeduplication)
	state.MobileNumberVerificationRequired = types.BoolValue(app.MobileNumberVerificationRequired)
	state.HostedPageGroup = types.StringValue(app.HostedPageGroup)
	state.PrimaryColor = types.StringValue(app.PrimaryColor)
	state.AccentColor = types.StringValue(app.AccentColor)
	state.AutoLoginAfterRegister = types.BoolValue(app.AutoLoginAfterRegister)
	state.AcceptRolesInTheRegistration = types.BoolValue(app.AcceptRolesInTheRegistration)
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

	state.AllowedGroups, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_id":      types.StringType,
			"roles":         types.ListType{ElemType: types.StringType},
			"default_roles": types.ListType{ElemType: types.StringType},
		},
	}, app.AllowedGroups)

	ret.Append(diags...)

	state.OperationsAllowedGroups, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_id":      types.StringType,
			"roles":         types.ListType{ElemType: types.StringType},
			"default_roles": types.ListType{ElemType: types.StringType},
		},
	}, app.OperationsAllowedGroups)

	ret.Append(diags...)

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

	state.CustomProviders = []CustomProvider{}
	for _, item := range app.CustomProviders {
		state.CustomProviders = append(state.CustomProviders, CustomProvider{
			DisplayName:  types.StringValue(item.DisplayName),
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

func planToApp(ctx context.Context, plan *App, state *App) (*client.App, diag.Diagnostics) {
	ret := diag.Diagnostics{}

	var diags diag.Diagnostics
	plannedApp := client.App{
		ID:                               state.ID.ValueString(),
		AppOwner:                         state.AppOwner.ValueString(),
		BotProvider:                      state.BotProvider.ValueString(),
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
		AcceptRolesInTheRegistration:     plan.AcceptRolesInTheRegistration.ValueBool(),

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
		CustomProviders: []client.CustomProvider{},
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

	for _, customProvider := range plan.CustomProviders {
		plannedApp.CustomProviders = append(
			plannedApp.CustomProviders,
			client.CustomProvider{
				DisplayName:  customProvider.DisplayName.ValueString(),
				ProviderName: customProvider.ProviderName.ValueString(),
			},
		)
	}

	diags = tfsdk.ValueAs(ctx, plan.AllowedGroups, &plannedApp.AllowedGroups)
	ret.Append(diags...)

	diags = tfsdk.ValueAs(ctx, plan.OperationsAllowedGroups, &plannedApp.OperationsAllowedGroups)
	ret.Append(diags...)

	diags = tfsdk.ValueAs(ctx, plan.TemplateGroupId, &plannedApp.TemplateGroupId)
	ret.Append(diags...)

	diags = tfsdk.ValueAs(ctx, plan.PasswordPolicy, &plannedApp.PasswordPolicy)
	ret.Append(diags...)

	return &plannedApp, ret
}
