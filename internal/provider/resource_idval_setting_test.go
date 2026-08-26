package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

func mustConsentsList(t *testing.T, consents []client.IdValConsent) types.List {
	t.Helper()

	list, diags := types.ListValueFrom(context.Background(), types.ObjectType{AttrTypes: idValConsentAttrTypes}, consents)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building consents list: %v", diags)
	}

	return list
}

func TestBuildClientSetting_HardcodesDisabledSubConfigsRegardlessOfPlan(t *testing.T) {
	plan := IdValSetting{
		ID:                  types.StringValue("b4ba52ae-87b3-4967-93b6-8e6ae8f39922"),
		Name:                types.StringValue("Kaufland Marketplace Age Check DE"),
		Description:         types.StringValue("Kaufland Marketplace Age Check DE"),
		Mode:                types.StringValue("AgeCheckEssential"),
		Theme:               types.StringNull(),
		AllowedRedirectUris: types.StringValue("https://account.kaufland.com https://www.kaufland.de"),
		ConsentConfig: IdValConsentConfig{
			Enabled: types.BoolValue(true),
			Consents: mustConsentsList(t, []client.IdValConsent{
				{
					Name:         "Consent",
					URL:          "https://www.kaufland.de/i/rechtliches/datenschutz",
					Mandatory:    true,
					Localization: map[string]string{"de": "Text", "en": "Text"},
				},
			}),
		},
	}

	setting, diags := buildClientSetting(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if setting.ID != "b4ba52ae-87b3-4967-93b6-8e6ae8f39922" {
		t.Errorf("ID = %q, want the plan's id", setting.ID)
	}
	if !setting.ConsentConfig.Enabled || len(setting.ConsentConfig.Consents) != 1 {
		t.Fatalf("ConsentConfig not mapped from plan: %+v", setting.ConsentConfig)
	}
	if setting.ConsentConfig.Consents[0].Name != "Consent" {
		t.Errorf("Consents[0].Name = %q, want %q", setting.ConsentConfig.Consents[0].Name, "Consent")
	}

	// The three unsupported sub-configs must always be the fixed disabled defaults,
	// never derived from the plan (the plan has no fields for them at all).
	if got, want := setting.PrevalidationConfig, client.DisabledPrevalidationConfig; !prevalidationConfigsEqual(got, want) {
		t.Errorf("PrevalidationConfig = %+v, want the disabled default %+v", got, want)
	}
	if got, want := setting.DocumentDataMatchingConfig.Enabled, client.DisabledDocumentMatchingConfig.Enabled; got != want {
		t.Errorf("DocumentDataMatchingConfig.Enabled = %v, want %v", got, want)
	}
	if len(setting.DocumentDataMatchingConfig.Fields) != len(client.DisabledDocumentMatchingConfig.Fields) {
		t.Errorf("DocumentDataMatchingConfig.Fields has %d entries, want %d", len(setting.DocumentDataMatchingConfig.Fields), len(client.DisabledDocumentMatchingConfig.Fields))
	}
	if got, want := setting.IdDocumentFilterConfig, client.DisabledDocumentFilterConfig; got.Enabled != want.Enabled || got.FilterMode != want.FilterMode {
		t.Errorf("IdDocumentFilterConfig = %+v, want the disabled default %+v", got, want)
	}
}

func prevalidationConfigsEqual(a, b client.IdValPrevalidationConfig) bool {
	if a.Enabled != b.Enabled || len(a.Fields) != len(b.Fields) || len(a.Description) != len(b.Description) {
		return false
	}
	for k, v := range a.Description {
		if b.Description[k] != v {
			return false
		}
	}
	return true
}

func TestSettingToState_MapsConsentConfigFromApiResponse(t *testing.T) {
	setting := &client.IdValSetting{
		ID:                  "b4ba52ae-87b3-4967-93b6-8e6ae8f39922",
		Name:                "Kaufland Marketplace Age Check DE",
		Description:         "Kaufland Marketplace Age Check DE",
		Mode:                "AgeCheckEssential",
		AllowedRedirectUris: "https://account.kaufland.com https://www.kaufland.de",
		CreatedTime:         "2026-08-25T09:16:31.208949203Z",
		UpdatedTime:         "2026-08-25T09:16:31.208949203Z",
		ConsentConfig: client.IdValConsentConfig{
			Enabled: true,
			Consents: []client.IdValConsent{
				{
					Name:         "Consent",
					URL:          "https://www.kaufland.de/i/rechtliches/datenschutz",
					Mandatory:    true,
					Localization: map[string]string{"de": "Text", "en": "Text"},
				},
			},
		},
		PrevalidationConfig:        client.DisabledPrevalidationConfig,
		DocumentDataMatchingConfig: client.DisabledDocumentMatchingConfig,
		IdDocumentFilterConfig:     client.DisabledDocumentFilterConfig,
	}

	state, diags := settingToState(context.Background(), setting)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != setting.ID {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), setting.ID)
	}
	if state.CreatedTime.ValueString() != setting.CreatedTime {
		t.Errorf("CreatedTime = %q, want %q", state.CreatedTime.ValueString(), setting.CreatedTime)
	}

	var consents []client.IdValConsent
	diags = state.ConsentConfig.Consents.ElementsAs(context.Background(), &consents, true)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics decoding consents: %v", diags)
	}
	if len(consents) != 1 || consents[0].Name != "Consent" || consents[0].Localization["de"] != "Text" {
		t.Errorf("consents round-tripped incorrectly: %+v", consents)
	}
}
