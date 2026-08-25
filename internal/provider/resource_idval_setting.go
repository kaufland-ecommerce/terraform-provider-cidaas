package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type idValSettingResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*idValSettingResource)(nil)

func NewIdValSettingResource() resource.Resource {
	return &idValSettingResource{}
}

func (r *idValSettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_idval_setting"
}

func (r *idValSettingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

var idValConsentAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"url":          types.StringType,
	"mandatory":    types.BoolType,
	"localization": types.MapType{ElemType: types.StringType},
}

func (r *idValSettingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_idval_setting` manages a cidaas ID Validator settings instance, e.g. an age verification configuration. Prevalidation, document data matching and the ID document filter are not supported by this resource and are always sent disabled. **Deletion is not supported via Terraform** - cidaas appears to authorize deleting this resource by role/group membership on a human identity rather than by OAuth scope, so a client_credentials app can create, read and update it but not delete it. Delete manually via the cidaas Admin UI, then run `terraform state rm` to drop it from state (or just remove the resource from your config and run `terraform apply -refresh-only`, since a manually-deleted setting is also detected on the next `Read` and dropped from state automatically).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Unique identifier of the ID Validator setting",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the setting",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description of the setting",
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "ID validation mode",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"IdentCard", "IdentPhoto", "IdentLight", "AgeCheckCard", "AgeCheckLight", "AgeCheckEssential", "OnboardingLight", "OnboardingEssential",
					),
				},
			},
			"theme": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Theme applied to the ID Validator UI",
			},
			"allowed_redirect_uris": schema.StringAttribute{
				Required:    true,
				Description: "Space-separated list of URLs the end-user may be redirected to after completion",
			},
			"created_time": schema.StringAttribute{
				Computed:    true,
				Description: "Time the setting was created",
			},
			"updated_time": schema.StringAttribute{
				Computed:    true,
				Description: "Time the setting was last updated",
			},
			"consent_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Consent configuration shown during the ID validation process",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Required:    true,
						Description: "Whether consent collection is enabled",
					},
					"consents": schema.ListNestedAttribute{
						Required:    true,
						Description: "List of consents to display to the end-user",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Internal reference name of the consent",
								},
								"url": schema.StringAttribute{
									Required:    true,
									Description: "Publicly accessible consent document URL",
								},
								"mandatory": schema.BoolAttribute{
									Required:    true,
									Description: "Whether the consent must be accepted to continue",
								},
								"localization": schema.MapAttribute{
									Required:    true,
									ElementType: types.StringType,
									Description: "Localized consent text keyed by language code",
								},
							},
						},
					},
				},
			},
		},
	}
}

func idValConsentConfigToClient(ctx context.Context, config IdValConsentConfig) (client.IdValConsentConfig, diag.Diagnostics) {
	var consents []client.IdValConsent
	diags := config.Consents.ElementsAs(ctx, &consents, true)

	return client.IdValConsentConfig{
		Enabled:  config.Enabled.ValueBool(),
		Consents: consents,
	}, diags
}

func idValConsentConfigFromClient(ctx context.Context, config client.IdValConsentConfig) (IdValConsentConfig, diag.Diagnostics) {
	consents, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: idValConsentAttrTypes}, config.Consents)

	return IdValConsentConfig{
		Enabled:  types.BoolValue(config.Enabled),
		Consents: consents,
	}, diags
}

// buildClientSetting maps a Terraform plan into the API-facing struct. The three
// disabled sub-configs are always the fixed client.Disabled*Config values - never read
// from plan/state, since this resource doesn't expose prevalidation, document data
// matching or the ID document filter.
func buildClientSetting(ctx context.Context, plan IdValSetting) (client.IdValSetting, diag.Diagnostics) {
	consentConfig, diags := idValConsentConfigToClient(ctx, plan.ConsentConfig)
	if diags.HasError() {
		return client.IdValSetting{}, diags
	}

	setting := client.IdValSetting{
		ID:                         plan.ID.ValueString(),
		Name:                       plan.Name.ValueString(),
		Description:                plan.Description.ValueString(),
		Mode:                       plan.Mode.ValueString(),
		Theme:                      plan.Theme.ValueString(),
		AllowedRedirectUris:        plan.AllowedRedirectUris.ValueString(),
		ConsentConfig:              consentConfig,
		PrevalidationConfig:        client.DisabledPrevalidationConfig,
		DocumentDataMatchingConfig: client.DisabledDocumentMatchingConfig,
		IdDocumentFilterConfig:     client.DisabledDocumentFilterConfig,
	}

	return setting, diags
}

// settingToState maps an API response into Terraform state.
func settingToState(ctx context.Context, setting *client.IdValSetting) (IdValSetting, diag.Diagnostics) {
	consentConfig, diags := idValConsentConfigFromClient(ctx, setting.ConsentConfig)

	state := IdValSetting{
		ID:                  types.StringValue(setting.ID),
		Name:                types.StringValue(setting.Name),
		Description:         types.StringValue(setting.Description),
		Mode:                types.StringValue(setting.Mode),
		Theme:               types.StringValue(setting.Theme),
		AllowedRedirectUris: types.StringValue(setting.AllowedRedirectUris),
		CreatedTime:         types.StringValue(setting.CreatedTime),
		UpdatedTime:         types.StringValue(setting.UpdatedTime),
		ConsentConfig:       consentConfig,
	}

	return state, diags
}

func (r *idValSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan IdValSetting

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedSetting, diags := buildClientSetting(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setting, err := r.provider.client.UpsertIdValSetting(plannedSetting)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ID Validator setting",
			"Could not create setting, unexpected error: "+err.Error(),
		)
		return
	}

	state, diags := settingToState(ctx, setting)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *idValSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var settingId string

	diags := req.State.GetAttribute(ctx, path.Root("id"), &settingId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setting, err := r.provider.client.GetIdValSetting(settingId)
	if err != nil {
		// id-val-srv returns 401 (not 404, unlike other cidaas resources) for a GET on
		// an id that simply doesn't exist - confirmed by observation, not documented.
		// Since this resource also can't be deleted via Terraform (see Delete), a
		// missing/never-created setting is a real scenario on refresh, not just a
		// theoretical one - drop it from state without erroring so the next plan
		// offers to create it again, instead of failing forever on a stale state
		// entry. Note this does mean a genuinely broken client_credentials token would
		// also look like "not found" here rather than surfacing as an auth error - but
		// Create/Update still hard-error on 401, so a real credentials problem still
		// surfaces loudly during the same apply for any resource that needs creating.
		if err.Error() == "resource not found" || strings.Contains(err.Error(), "status 401") {
			req.State.RemoveResource(ctx)
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading ID Validator setting",
			"Unexpected error fetching setting: "+err.Error(),
		)
		return
	}

	state, diags := settingToState(ctx, setting)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *idValSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan IdValSetting

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedSetting, diags := buildClientSetting(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setting, err := r.provider.client.UpsertIdValSetting(plannedSetting)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ID Validator setting",
			"Could not update setting, unexpected error: "+err.Error(),
		)
		return
	}

	state, diags := settingToState(ctx, setting)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *idValSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var state IdValSetting

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.DeleteIdValSetting(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "status 401") {
			resp.Diagnostics.AddError(
				"Cannot delete ID Validator setting via Terraform",
				"cidaas rejected the delete call with 401. Deleting an ID Validator setting appears to require a human "+
					"identity holding the IDVAL_ACCOUNTANT role (confirmed by inspecting a working delete from the Admin UI: "+
					"it authorized via role/group membership, not via any cidaas:idval_settings_* OAuth scope) - a "+
					"NON_INTERACTIVE client_credentials token has no user identity to hold that role, so this may not be "+
					"possible from Terraform regardless of granted scopes. Delete \""+state.ID.ValueString()+"\" manually "+
					"via the cidaas Admin UI (ID validator > ID validation Settings), then remove it from Terraform state "+
					"with `terraform state rm`.",
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting ID Validator setting",
			"Could not delete setting, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *idValSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	setting, err := r.provider.client.GetIdValSetting(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing ID Validator setting", err.Error())
		return
	}

	state, diags := settingToState(ctx, setting)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
