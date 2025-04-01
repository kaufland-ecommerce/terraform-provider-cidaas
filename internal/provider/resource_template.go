package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
	"github.com/real-digital/terraform-provider-cidaas/internal/util"
)

type templateResource struct {
	provider *cidaasProvider
}

var _ resource.Resource = (*templateResource)(nil)

func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.provider, resp.Diagnostics = toProvider(req.ProviderData)
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`cidaas_template_group` manages Template Groups in the tenant.\n\n",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Cidaas UUID of the Template",
			},
			"last_seeded_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "Group of this template",
			},
			"template_key": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the template",
			},
			"communication_method": schema.StringAttribute{
				Required:    true,
				Description: "Which Communication method is this template for",
			},
			"processing_type": schema.StringAttribute{
				Optional:    true,
				Description: "Processing Type",
			},
			"locale": schema.StringAttribute{
				Required:    true,
				Description: "Locale",
			},
			"message_format": schema.StringAttribute{
				Required:    true,
				Description: "Language",
			},
			"subject": schema.StringAttribute{
				Optional:    true,
				Description: "Subject of the Template",
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "actual content of the Template",
			},
			"enabled": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Template",
			},
		},
	}
}

func (r templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan Template

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	template := client.Template{
		ID:                  nil,
		LastSeededBy:        nil,
		GroupId:             plan.GroupId.ValueString(),
		TemplateKey:         plan.TemplateKey.ValueString(),
		CommunicationMethod: plan.CommunicationMethod.ValueString(),
		ProcessingType:      plan.ProcessingType.ValueString(),
		Locale:              plan.Locale.ValueString(),
		MessageFormat:       plan.MessageFormat.ValueString(),
		Enabled:             plan.Enabled.ValueBool(),
		Subject:             plan.Subject.ValueString(),
		Content:             plan.Content.ValueString(),
	}

	templateResult, err := r.provider.client.UpdateTemplate(template)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template",
			"Could not create hook, unexpected error: "+err.Error(),
		)
		return
	}

	tfsdk.ValueFrom(ctx, templateResult.ID, types.StringType, &plan.ID)
	tfsdk.ValueFrom(ctx, templateResult.LastSeededBy, types.StringType, &plan.LastSeededBy)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Template
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	template := &client.Template{
		ID:                  nil,
		LastSeededBy:        nil,
		GroupId:             state.GroupId.ValueString(),
		TemplateKey:         state.TemplateKey.ValueString(),
		CommunicationMethod: state.CommunicationMethod.ValueString(),
		ProcessingType:      state.ProcessingType.ValueString(),
		Locale:              state.Locale.ValueString(),
		MessageFormat:       state.MessageFormat.ValueString(),
		Enabled:             state.Enabled.ValueBool(),
	}

	template, err := r.provider.client.GetTemplate(*template)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Template",
			"Could not read template "+state.TemplateKey.ValueString()+": "+err.Error(),
		)
		return
	}

	tfsdk.ValueFrom(ctx, template.LastSeededBy, types.StringType, &state.LastSeededBy)
	tfsdk.ValueFrom(ctx, template.Subject, types.StringType, &state.Subject)
	tfsdk.ValueFrom(ctx, template.Content, types.StringType, &state.Content)

	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)
}

func (r templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan Template

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	template := client.Template{
		ID:                  util.ToStringPointer(plan.ID.ValueString()),
		GroupId:             plan.GroupId.ValueString(),
		TemplateKey:         plan.TemplateKey.ValueString(),
		CommunicationMethod: plan.CommunicationMethod.ValueString(),
		ProcessingType:      plan.ProcessingType.ValueString(),
		Locale:              plan.Locale.ValueString(),
		MessageFormat:       plan.MessageFormat.ValueString(),
		Enabled:             plan.Enabled.ValueBool(),
		Subject:             plan.Subject.ValueString(),
		Content:             plan.Content.ValueString(),
	}

	templateResult, err := r.provider.client.UpdateTemplate(template)

	if err != nil {
		resp.Diagnostics.AddError("Could not update Template", err.Error())
		return
	}

	tfsdk.ValueFrom(ctx, template.LastSeededBy, types.StringType, &templateResult.LastSeededBy)
	tfsdk.ValueFrom(ctx, template.Subject, types.StringType, &templateResult.Subject)
	tfsdk.ValueFrom(ctx, template.Content, types.StringType, &templateResult.Content)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
