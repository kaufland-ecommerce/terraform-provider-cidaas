package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

type themeResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*themeResource)(nil)
var _ resource.ResourceWithImportState = (*themeResource)(nil)

func NewThemeResource() resource.Resource {
	return &themeResource{}
}

func (r *themeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme"
}

func (r *themeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *themeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_theme` manages a Hosted Pages theme (CSS) via the `hostedpages-srv/themes` API. Used e.g. to theme the ID Validator UI referenced from `cidaas_idval_setting.theme` - cidaas requires a theme's name to be prefixed `idval` for it to be selectable as an ID Validator theme.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name of the theme - prefix with `idval` to make it selectable as an ID Validator theme",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"css": schema.StringAttribute{
				Required:    true,
				Description: "Theme CSS content (a `:root{...}` block of `--idval-*`/hosted-pages CSS variables), e.g. loaded via Terraform's `file()` function",
			},
		},
	}
}

func (r *themeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan Theme
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	theme := client.Theme{
		Name: plan.Name.ValueString(),
		Css:  plan.Css.ValueString(),
	}

	if err := r.provider.client.UpsertTheme(theme); err != nil {
		resp.Diagnostics.AddError("Error creating theme", "Could not create theme, unexpected error: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *themeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var state Theme
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	theme, err := r.provider.client.GetTheme(state.Name.ValueString())
	if err != nil {
		if err.Error() == "resource not found" {
			req.State.RemoveResource(ctx)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading theme", "Unexpected error fetching theme: "+err.Error())
		return
	}

	state.Css = types.StringValue(theme.Css)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *themeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var plan Theme
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	theme := client.Theme{
		Name: plan.Name.ValueString(),
		Css:  plan.Css.ValueString(),
	}

	if err := r.provider.client.UpsertTheme(theme); err != nil {
		resp.Diagnostics.AddError("Error updating theme", "Could not update theme, unexpected error: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *themeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider has not been configured. This typically occurs when a resource attribute depends on a value that is unknown until another resource is applied. Apply the dependent resource separately, then retry.",
		)
		return
	}

	var state Theme
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.provider.client.DeleteTheme(state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting theme", "Could not delete theme, unexpected error: "+err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *themeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
