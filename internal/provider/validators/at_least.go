package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GreaterOrEqualValidtor struct {
	Min int64
}

func AtLeast(min int64) GreaterOrEqualValidtor {
	return GreaterOrEqualValidtor{
		Min: min,
	}
}

func (v GreaterOrEqualValidtor) Description(ctx context.Context) string {
	return fmt.Sprintf("Value needs to be equal or greater %d", v.Min)
}

func (v GreaterOrEqualValidtor) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Value needs to be equal or greater %d", v.Min)
}

func (v GreaterOrEqualValidtor) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var intAttr types.Int64
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &intAttr)

	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if intAttr.Unknown || intAttr.Null {
		return
	}

	if intAttr.Value < v.Min {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid int value",
			fmt.Sprintf("Int value needs to be greater or equal to %d", v.Min),
		)
	}
}
