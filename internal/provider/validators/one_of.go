package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

type ValueInListValidator struct {
	List []string
}

func OneOf(list []string) ValueInListValidator {
	return ValueInListValidator{
		List: list,
	}
}

func (v ValueInListValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value needs to be in list %+v", v.List)
}

func (v ValueInListValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Value needs to be in list %+v", v.List)
}

func (v ValueInListValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)

	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	if str.Null || str.Unknown {
		return
	}

	if !slices.Contains(v.List, str.Value) {
		resp.Diagnostics.AddError(
			"Value not in List",
			fmt.Sprintf("'%s' is not in the list of possible values: %+v", str.Value, v.List),
		)
	}
}
