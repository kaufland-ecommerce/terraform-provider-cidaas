package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
	"strings"
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
				Required:    true,
				Description: "Cidaas unique ID of the Template",
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
			"usage_type": schema.StringAttribute{
				Optional:    true,
				Description: "Usage Type",
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
			"enabled": schema.BoolAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Template",
			},
			"verification_type": schema.StringAttribute{
				Optional:    true,
				Description: "Verification Type",
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
		ID:                  plan.ID.ValueString(),
		GroupId:             plan.GroupId.ValueString(),
		TemplateKey:         plan.TemplateKey.ValueString(),
		CommunicationMethod: plan.CommunicationMethod.ValueString(),
		ProcessingType:      plan.ProcessingType.ValueString(),
		UsageType:           plan.UsageType.ValueString(),
		Locale:              plan.Locale.ValueString(),
		MessageFormat:       plan.MessageFormat.ValueString(),
		Owner:               "client",
		Enabled:             plan.Enabled.ValueBool(),
		Subject:             plan.Subject.ValueString(),
		Content:             plan.Content.ValueString(),
		Description:         plan.Description.ValueString(),
		VerificationType:    plan.VerificationType.ValueString(),
	}

	templateResult, err := r.provider.client.CreateTemplate(template)

	if err != nil {
		if strings.Contains(err.Error(), "template already found") {
			updateResponse, errUpdate := r.provider.client.UpdateTemplate(template)
			if errUpdate != nil {
				resp.Diagnostics.AddError(
					"Error creating template",
					"Could not create template ID: "+plan.ID.ValueString()+", unexpected error: "+err.Error(),
				)
				return
			}

			r.resultToState(ctx, &plan, updateResponse)
			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
			return
		}
		resp.Diagnostics.AddError(
			"Error creating template",
			"Could not create template ID: "+plan.ID.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}
	r.resultToState(ctx, &plan, templateResult)

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

	if len(state.ID.ValueString()) == 0 {
		// Resource has no ID, remove it from state without warning
		req.State.RemoveResource(ctx)
		resp.State.RemoveResource(ctx)
		return
	}
	template, err := r.provider.client.GetTemplate(state.ID.ValueString())
	if err != nil {
		if err.Error() == "resource not found" {
			// Resource not found, remove it from state without warning
			req.State.RemoveResource(ctx)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Template",
			"Could not read template "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	r.resultToState(ctx, &state, template)

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
		ID:                  plan.ID.ValueString(),
		GroupId:             plan.GroupId.ValueString(),
		TemplateKey:         plan.TemplateKey.ValueString(),
		CommunicationMethod: plan.CommunicationMethod.ValueString(),
		ProcessingType:      plan.ProcessingType.ValueString(),
		UsageType:           plan.UsageType.ValueString(),
		Locale:              plan.Locale.ValueString(),
		MessageFormat:       plan.MessageFormat.ValueString(),
		Enabled:             plan.Enabled.ValueBool(),
		Subject:             plan.Subject.ValueString(),
		Content:             plan.Content.ValueString(),
		Description:         plan.Description.ValueString(),
		VerificationType:    plan.VerificationType.ValueString(),
	}

	templateResult, err := r.provider.client.UpdateTemplate(template)

	if err != nil {
		resp.Diagnostics.AddError("Could not update Template ID:"+plan.ID.ValueString(), err.Error())
		return
	}

	r.resultToState(ctx, &plan, templateResult)

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

func (r templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	template, err := r.provider.client.GetTemplate(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch template",
			err.Error(),
		)
		return
	}
	var state Template
	r.resultToState(ctx, &state, template)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r templateResource) resultToState(ctx context.Context, state *Template, template *client.Template) {
	state.ID = types.StringValue(template.ID)
	state.GroupId = types.StringValue(template.GroupId)
	state.TemplateKey = types.StringValue(template.TemplateKey)
	state.CommunicationMethod = types.StringValue(template.CommunicationMethod)
	state.ProcessingType = types.StringValue(template.ProcessingType)
	state.UsageType = types.StringValue(template.UsageType)
	state.Locale = types.StringValue(template.Locale)
	state.MessageFormat = types.StringValue(template.MessageFormat)
	state.Enabled = types.BoolValue(template.Enabled)
	state.Description = types.StringValue(template.Description)
	if len(template.VerificationType) != 0 {
		state.VerificationType = types.StringValue(template.VerificationType)
	} else {
		state.VerificationType = types.StringNull()
	}

	state.ProcessingType = types.StringValue(template.ProcessingType)
	if len(template.UsageType) != 0 {
		state.UsageType = types.StringValue(template.UsageType)
	} else {
		state.UsageType = types.StringNull()
	}

	tfsdk.ValueFrom(ctx, template.Subject, types.StringType, &state.Subject)
	tfsdk.ValueFrom(ctx, template.Content, types.StringType, &state.Content)
}
