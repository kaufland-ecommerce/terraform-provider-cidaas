package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MinimumStringLengthValidator struct {
	Min int64
}

func MinimumStringLength(min int64) MinimumStringLengthValidator {
	return MinimumStringLengthValidator{
		Min: min,
	}
}

func NonEmptyString() MinimumStringLengthValidator {
	return MinimumStringLengthValidator{
		Min: 1,
	}
}

func (v MinimumStringLengthValidator) Description(context.Context) string {
	return fmt.Sprintf("Value needs to be equal or greater %d", v.Min)
}

func (v MinimumStringLengthValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("Value needs to be equal or greater %d", v.Min)
}

func (v MinimumStringLengthValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)

	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	strLen := int64(len(str.Value))

	if strLen < v.Min {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid string Length",
			fmt.Sprintf("String needs to be longer than %d character(s)", v.Min),
		)
	}
}
