package validators_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/provider/validators"
	"testing"
)

func TestOneOfValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         attr.Value
		list        []string
		expectError bool
	}

	tests := map[string]testCase{
		"not a String": {
			val:         types.Bool{Value: true},
			expectError: true,
		},
		"unknown String": {
			val:  types.String{Unknown: true},
			list: []string{},
		},
		"null String": {
			val:  types.String{Null: true},
			list: []string{},
		},
		"valid string in list": {
			val:  types.String{Value: "test"},
			list: []string{"foo", "test"},
		},
		"valid string not in list": {
			val:         types.String{Value: "not-test"},
			list:        []string{"test", "example"},
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			request := tfsdk.ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: test.val,
			}

			response := tfsdk.ValidateAttributeResponse{}
			validators.OneOf(test.list).Validate(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
