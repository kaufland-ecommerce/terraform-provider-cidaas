package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

func mustStringSet(t *testing.T, values []string) types.Set {
	t.Helper()

	set, diags := types.SetValueFrom(context.Background(), types.StringType, values)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building set: %v", diags)
	}

	return set
}

func TestBuildNotificationServiceSetup_MapsPlanToClientStruct(t *testing.T) {
	plan := NotificationServiceSetup{
		ID:                   types.StringValue("existing-id"),
		Name:                 types.StringValue("SES Email"),
		ServiceId:            types.StringValue("custom-ses-email"),
		CommunicationMethods: mustStringSet(t, []string{"email"}),
		Description:          types.StringValue("Marketplace transactional email"),
		HasRemoteTemplates:   types.BoolValue(false),
		ParentServiceSetupId: types.StringNull(),
		ServiceCategory:      types.StringValue("comm_prov"),
	}

	setup, diags := buildNotificationServiceSetup(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if setup.Name != "SES Email" {
		t.Errorf("Name = %q, want %q", setup.Name, "SES Email")
	}
	if setup.ServiceId != "custom-ses-email" {
		t.Errorf("ServiceId = %q, want %q", setup.ServiceId, "custom-ses-email")
	}
	if len(setup.CommunicationMethods) != 1 || setup.CommunicationMethods[0] != "email" {
		t.Errorf("CommunicationMethods = %v, want [email]", setup.CommunicationMethods)
	}
	if setup.ServiceCategory != "comm_prov" {
		t.Errorf("ServiceCategory = %q, want %q", setup.ServiceCategory, "comm_prov")
	}
}

func TestNotificationServiceSetupToState_MapsApiResponseToState(t *testing.T) {
	setup := &client.NotificationServiceSetup{
		ID:                   "service-setup-id",
		Name:                 "SES Email",
		ServiceId:            "custom-ses-email",
		CommunicationMethods: []string{"email"},
		Description:          "Marketplace transactional email",
		HasRemoteTemplates:   false,
		ServiceCategory:      "comm_prov",
		Status:               "in-progress",
	}

	state, diags := notificationServiceSetupToState(context.Background(), setup)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != "service-setup-id" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "service-setup-id")
	}
	if state.Status.ValueString() != "in-progress" {
		t.Errorf("Status = %q, want %q", state.Status.ValueString(), "in-progress")
	}

	var methods []string
	elemDiags := state.CommunicationMethods.ElementsAs(context.Background(), &methods, false)
	if elemDiags.HasError() {
		t.Fatalf("unexpected diagnostics decoding communication_methods: %v", elemDiags)
	}
	if len(methods) != 1 || methods[0] != "email" {
		t.Errorf("communication_methods round-tripped incorrectly: %v", methods)
	}
}
