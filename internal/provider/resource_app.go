package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
				Optional:    true,
				Computed:    true,
				Description: "Not supported for client_type = \"NON_INTERACTIVE\" apps; must be left unset there.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"accent_color": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Not supported for client_type = \"NON_INTERACTIVE\" apps; must be left unset there.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"oauth_standard": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Oauth standard used, default is OAuth2.1",
				Default:     stringdefault.StaticString("OAuth2.1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
				Optional:    true,
				Computed:    true,
				Description: "Not supported for client_type = \"NON_INTERACTIVE\" apps; must be left unset there.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
						// @deprecated
						//"provider_type": schema.StringAttribute{
						//	Computed: true,
						//	PlanModifiers: []planmodifier.String{
						//		stringplanmodifier.UseStateForUnknown(),
						//	},
						//},
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
			"communication_medium_verification": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Not supported for client_type = \"NON_INTERACTIVE\" apps; must be left unset there.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Captcha
			// TODO

			// Password Rules
			"password_policy": schema.StringAttribute{
				Optional: true,
			},

			// Template Group
			"hosted_page_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Not supported for client_type = \"NON_INTERACTIVE\" apps; must be left unset there.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional:    true,
				Computed:    true,
				Description: "Not supported for client_type = \"NON_INTERACTIVE\" apps; must be left unset there.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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

// nonInteractiveUnsupportedFields are fields cidaas silently drops for client_type =
// "NON_INTERACTIVE" apps, causing a permanent "inconsistent result after apply" error if set.
// Verified against cidaas's admin UI: accent_color, primary_color, hosted_page_group, and
// is_remember_me_selected are absent from its save requests; communication_medium_verification
// is sent as "none"; allowed_logout_urls has no UI field at all.
var nonInteractiveUnsupportedFields = []string{
	"accent_color",
	"primary_color",
	"hosted_page_group",
	"communication_medium_verification",
	"is_remember_me_selected",
	// NON_INTERACTIVE apps have no browser session to redirect after logout. UI shows
	// "Backchannel Logout URI" instead, no "Allowed Logout URLs" field.
	"allowed_logout_urls",
}

// nonInteractiveFieldViolations returns the subset of nonInteractiveUnsupportedFields that are
// set (true) in setFields, or nil if clientType isn't "NON_INTERACTIVE".
func nonInteractiveFieldViolations(clientType string, setFields map[string]bool) []string {
	if clientType != "NON_INTERACTIVE" {
		return nil
	}

	var violations []string
	for _, name := range nonInteractiveUnsupportedFields {
		if setFields[name] {
			violations = append(violations, name)
		}
	}
	return violations
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

	var clientType types.String
	var accentColor, primaryColor, hostedPageGroup, communicationMediumVerification types.String
	var isRememberMeSelected types.Bool
	var allowedLogoutUrls []string

	req.Config.GetAttribute(ctx, path.Root("client_type"), &clientType)
	req.Config.GetAttribute(ctx, path.Root("accent_color"), &accentColor)
	req.Config.GetAttribute(ctx, path.Root("primary_color"), &primaryColor)
	req.Config.GetAttribute(ctx, path.Root("hosted_page_group"), &hostedPageGroup)
	req.Config.GetAttribute(ctx, path.Root("communication_medium_verification"), &communicationMediumVerification)
	req.Config.GetAttribute(ctx, path.Root("is_remember_me_selected"), &isRememberMeSelected)
	req.Config.GetAttribute(ctx, path.Root("allowed_logout_urls"), &allowedLogoutUrls)

	setFields := map[string]bool{
		"accent_color":                      !accentColor.IsNull(),
		"primary_color":                     !primaryColor.IsNull(),
		"hosted_page_group":                 !hostedPageGroup.IsNull(),
		"communication_medium_verification": !communicationMediumVerification.IsNull(),
		"is_remember_me_selected":           !isRememberMeSelected.IsNull(),
		"allowed_logout_urls":               allowedLogoutUrls != nil,
	}

	for _, name := range nonInteractiveFieldViolations(clientType.ValueString(), setFields) {
		resp.Diagnostics.AddAttributeError(
			path.Root(name),
			"Field not supported for NON_INTERACTIVE apps",
			fmt.Sprintf("cidaas does not persist %q for client_type = \"NON_INTERACTIVE\" apps (machine-to-machine clients have no login/hosted-page UI) — remove it from this resource's configuration.", name),
		)
	}
}

var _ resource.ResourceWithModifyPlan = (*appResource)(nil)

// ModifyPlan forces the six NON_INTERACTIVE-unsupported fields to null in the plan whenever the
// planned client_type is "NON_INTERACTIVE". Without this, UseStateForUnknown carries forward
// their prior known values on a client_type change away from an interactive type (they weren't
// touched in config), while applyAppToState nulls them after apply - the same inconsistent-result
// mismatch this resource otherwise guards against.
func (r appResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var clientType types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("client_type"), &clientType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if clientType.ValueString() != "NON_INTERACTIVE" {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("accent_color"), types.StringNull())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("primary_color"), types.StringNull())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("hosted_page_group"), types.StringNull())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("communication_medium_verification"), types.StringNull())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("is_remember_me_selected"), types.BoolNull())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("allowed_logout_urls"), types.ListNull(types.StringType))...)
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

	app, warning, err := createOrUpsertApp(r.provider.client, plannedApp)
	if warning != "" {
		resp.Diagnostics.AddWarning("Attempting to create an app that already exists", warning)
	}
	if err != nil {
		resp.Diagnostics.AddError("Could not create app", err.Error())
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

	app, err := updateAndRefetchApp(r.provider.client, plannedApp)
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

// createOrUpsertApp creates the app (falling back to an update if it already exists), then
// re-fetches it by ID via a real GET. The create/update response body is not guaranteed to
// reflect the full persisted object (see the prepareResponse comment in internal/client/app.go),
// so state must never be built directly from it. The returned warning, if non-empty, should be
// surfaced as a Terraform diagnostics warning by the caller.
func createOrUpsertApp(c client.Client, plannedApp *client.App) (app *client.App, warning string, err error) {
	app, err = c.CreateApp(plannedApp)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return nil, "", err
		}

		warning = "Applying update instead"

		existingApp, findErr := c.GetAppByName(plannedApp.ClientName)
		if findErr != nil {
			return nil, "", fmt.Errorf("could not find existing app: %w", findErr)
		}
		plannedApp.ID = existingApp.ID
		plannedApp.ClientId = existingApp.ClientId
	} else {
		plannedApp.ID = app.ID
		plannedApp.ClientId = app.ClientId
	}

	// CreateApp's POST doesn't reliably persist redirect_uris; a follow-up UpdateApp fixes
	// it, matching the "already exists" fallback above. Does not fix allowed_logout_urls
	// (see nonInteractiveUnsupportedFields).
	_, err = c.UpdateApp(*plannedApp)
	if err != nil {
		return nil, warning, err
	}

	app, err = c.GetApp(plannedApp.ClientId)
	if err != nil {
		return nil, warning, err
	}
	if app == nil {
		return nil, warning, fmt.Errorf("app not found immediately after creation")
	}

	return app, warning, nil
}

// updateAndRefetchApp updates the app, then re-fetches it by ID via a real GET for the same
// reason createOrUpsertApp does. AllowedOrigins is re-applied from the plan afterwards because
// the API silently ignores it on update (see the long-standing @FIXME below).
func updateAndRefetchApp(c client.Client, plannedApp *client.App) (*client.App, error) {
	_, err := c.UpdateApp(*plannedApp)
	if err != nil {
		return nil, err
	}

	app, err := c.GetApp(plannedApp.ClientId)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("app not found immediately after update")
	}

	//@FIXME: The property is being ignored from the update, setting to planned value will fix the inconsistencies
	app.AllowedOrigins = plannedApp.AllowedOrigins

	return app, nil
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
	state.ClientType = types.StringValue(app.ClientType)

	// cidaas doesn't persist these for NON_INTERACTIVE apps (see nonInteractiveUnsupportedFields) -
	// report null rather than whatever zero/neutral value the API returns, matching what
	// ValidateConfig requires the config to leave unset.
	nonInteractive := app.ClientType == "NON_INTERACTIVE"
	if nonInteractive {
		state.IsRememberMeSelected = types.BoolNull()
		state.HostedPageGroup = types.StringNull()
		state.PrimaryColor = types.StringNull()
		state.AccentColor = types.StringNull()
	} else {
		state.IsRememberMeSelected = types.BoolValue(app.IsRememberMeSelected)
		state.HostedPageGroup = types.StringValue(app.HostedPageGroup)
		state.PrimaryColor = types.StringValue(app.PrimaryColor)
		state.AccentColor = types.StringValue(app.AccentColor)
	}

	state.AllowDisposableEmail = types.BoolValue(app.AllowDisposableEmail)
	state.AllowGuestLogin = types.BoolValue(app.AllowGuestLogin)
	state.FdsEnabled = types.BoolValue(app.FdsEnabled)
	state.EnablePasswordlessAuth = types.BoolValue(app.EnablePasswordlessAuth)
	state.EnableDeduplication = types.BoolValue(app.EnableDeduplication)
	state.AutoLoginAfterRegister = types.BoolValue(app.AutoLoginAfterRegister)
	state.AcceptRolesInTheRegistration = types.BoolValue(app.AcceptRolesInTheRegistration)
	state.CompanyName = types.StringValue(app.CompanyName)
	state.CompanyAddress = types.StringValue(app.CompanyAddress)
	state.CompanyWebsite = types.StringValue(app.CompanyWebsite)
	state.TokenLifetimeInSeconds = types.Int64Value(app.TokenLifetimeInSeconds)
	state.IdTokenLifetimeInSeconds = types.Int64Value(app.IdTokenLifetimeInSeconds)
	state.RefreshTokenLifetimeInSeconds = types.Int64Value(app.RefreshTokenLifetimeInSeconds)
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
	tfsdk.ValueFrom(ctx, app.ID, types.StringType, &state.ID)

	state.Scopes = app.AllowedScopes
	state.RedirectUris = app.RedirectUris
	if nonInteractive {
		state.AllowedLogoutUrls = types.ListNull(types.StringType)
	} else {
		state.AllowedLogoutUrls, diags = types.ListValueFrom(ctx, types.StringType, app.AllowedLogoutUrls)
		ret.Append(diags...)
	}
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
	state.OauthStandard = types.StringValue(app.OauthStandard)
	if nonInteractive {
		state.CommunicationMediumVerification = types.StringNull()
	} else {
		state.CommunicationMediumVerification = types.StringValue(app.CommunicationMediumVerification)
	}

	ret.Append(diags...)

	return ret
}

func planToApp(ctx context.Context, plan *App, state *App) (*client.App, diag.Diagnostics) {
	ret := diag.Diagnostics{}

	var diags diag.Diagnostics
	plannedApp := client.App{
		ID:                            state.ID.ValueString(),
		AppOwner:                      state.AppOwner.ValueString(),
		BotProvider:                   state.BotProvider.ValueString(),
		ClientSecret:                  state.ClientSecret.ValueString(),
		ClientId:                      state.ClientId.ValueString(),
		ClientDisplayName:             plan.ClientDisplayName.ValueString(),
		ClientName:                    plan.ClientName.ValueString(),
		ClientType:                    plan.ClientType.ValueString(),
		IsRememberMeSelected:          plan.IsRememberMeSelected.ValueBool(),
		AllowDisposableEmail:          plan.AllowDisposableEmail.ValueBool(),
		AutoLoginAfterRegister:        plan.AutoLoginAfterRegister.ValueBool(),
		FdsEnabled:                    plan.FdsEnabled.ValueBool(),
		EnablePasswordlessAuth:        plan.EnablePasswordlessAuth.ValueBool(),
		EnableDeduplication:           plan.EnableDeduplication.ValueBool(),
		HostedPageGroup:               plan.HostedPageGroup.ValueString(),
		PrimaryColor:                  plan.PrimaryColor.ValueString(),
		AccentColor:                   plan.AccentColor.ValueString(),
		CompanyName:                   plan.CompanyName.ValueString(),
		CompanyWebsite:                plan.CompanyWebsite.ValueString(),
		CompanyAddress:                plan.CompanyAddress.ValueString(),
		TokenLifetimeInSeconds:        plan.TokenLifetimeInSeconds.ValueInt64(),
		IdTokenLifetimeInSeconds:      plan.IdTokenLifetimeInSeconds.ValueInt64(),
		RefreshTokenLifetimeInSeconds: plan.RefreshTokenLifetimeInSeconds.ValueInt64(),
		EnableBotDetection:            plan.EnableBotDetection.ValueBool(),
		IsLoginSuccessPageEnabled:     plan.IsLoginSuccessPageEnabled.ValueBool(),
		JweEnabled:                    plan.JweEnabled.ValueBool(),
		AlwaysAskMfa:                  plan.AlwaysAskMfa.ValueBool(),
		RegisterWithLoginInformation:  plan.RegisterWithLoginInformation.ValueBool(),
		AcceptRolesInTheRegistration:  plan.AcceptRolesInTheRegistration.ValueBool(),

		AllowLoginWith:               plan.AllowLoginWith,
		RedirectUris:                 plan.RedirectUris,
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

	plannedApp.OauthStandard = plan.OauthStandard.ValueString()
	plannedApp.CommunicationMediumVerification = plan.CommunicationMediumVerification.ValueString()

	// On a brand-new NON_INTERACTIVE resource with allowed_logout_urls left unset (as
	// ValidateConfig requires), the plan value is legitimately unknown - there's no prior
	// state for UseStateForUnknown to carry forward. []string can't represent unknown, so
	// leave it at its zero value (nil) rather than attempting (and failing) the conversion.
	if !plan.AllowedLogoutUrls.IsUnknown() {
		diags = tfsdk.ValueAs(ctx, plan.AllowedLogoutUrls, &plannedApp.AllowedLogoutUrls)
		ret.Append(diags...)
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
