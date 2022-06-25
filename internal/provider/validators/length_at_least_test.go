package validators_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/real-digital/terraform-provider-cidaas/internal/provider/validators"
	"testing"
)

func TestLengthAtLeastValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         attr.Value
		length      int64
		expectError bool
	}

	tests := map[string]testCase{
		"not a String": {
			val:         types.Bool{Value: true},
			expectError: true,
		},
		"unknown String": {
			val: types.String{Unknown: true},
		},
		"null String": {
			val: types.String{Null: true},
		},
		"valid string equal": {
			val:    types.String{Value: "test"},
			length: 4,
		},
		"valid string longer": {
			val:    types.String{Value: "long-test"},
			length: 1,
		},
		"valid string to short": {
			val:         types.String{Value: "test"},
			length:      100,
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			request := tfsdk.ValidateAttributeRequest{
				AttributePath:   tftypes.NewAttributePath().WithAttributeName("test"),
				AttributeConfig: test.val,
			}

			response := tfsdk.ValidateAttributeResponse{}
			validators.LengthAtLeast(test.length).Validate(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
