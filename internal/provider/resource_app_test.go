package provider

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/real-digital/terraform-provider-cidaas/internal/client"
)

// fakeAppClient implements client.Client by embedding it (nil) so only the methods under
// test need to be overridden; calling any other method panics on the nil embedded interface,
// which is fine since these tests never exercise them.
type fakeAppClient struct {
	client.Client
	createAppFn    func(app *client.App) (*client.App, error)
	getAppFn       func(clientId string) (*client.App, error)
	getAppByNameFn func(clientName string) (*client.App, error)
	updateAppFn    func(app client.App) (*client.App, error)
}

func (f *fakeAppClient) CreateApp(app *client.App) (*client.App, error) {
	return f.createAppFn(app)
}

func (f *fakeAppClient) GetApp(clientId string) (*client.App, error) {
	return f.getAppFn(clientId)
}

func (f *fakeAppClient) GetAppByName(clientName string) (*client.App, error) {
	return f.getAppByNameFn(clientName)
}

func (f *fakeAppClient) UpdateApp(app client.App) (*client.App, error) {
	return f.updateAppFn(app)
}

func TestCreateOrUpsertApp_BuildsStateFromGetAppNotCreateResponse(t *testing.T) {
	// Regression test: neither the raw POST /apps-srv/clients response nor the follow-up
	// UpdateApp response are guaranteed to reflect the full persisted object (confirmed:
	// redirect_uris/allowed_logout_urls lose their only element on a bare create), so the
	// resource must always re-fetch by ID before building state.
	complete := &client.App{
		ClientId:     "new-id",
		AccentColor:  "#123456",
		RedirectUris: []string{"https://example.com/cb"},
	}
	incomplete := &client.App{
		ClientId:     "new-id",
		AccentColor:  "",
		RedirectUris: []string{},
	}

	var updateCalledWith client.App
	var getAppCalledWith string
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
			return incomplete, nil
		},
		updateAppFn: func(app client.App) (*client.App, error) {
			updateCalledWith = app
			return incomplete, nil
		},
		getAppFn: func(clientId string) (*client.App, error) {
			getAppCalledWith = clientId
			return complete, nil
		},
	}

	plannedApp := &client.App{ClientName: "my-app"}
	got, warning, err := createOrUpsertApp(fake, plannedApp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if warning != "" {
		t.Fatalf("expected no warning, got %q", warning)
	}
	if updateCalledWith.ClientId != "new-id" {
		t.Fatalf("expected UpdateApp to be called with the created app's client id, got %+v", updateCalledWith)
	}
	if getAppCalledWith != "new-id" {
		t.Fatalf("expected GetApp to be called with %q, got %q", "new-id", getAppCalledWith)
	}
	if !reflect.DeepEqual(got, complete) {
		t.Fatalf("expected state to be built from GetApp result %+v, got %+v", complete, got)
	}
}

func TestCreateOrUpsertApp_AlreadyExistsFallsBackToUpdateThenRefetches(t *testing.T) {
	complete := &client.App{ClientId: "existing-id", AccentColor: "#abcdef"}
	incompleteUpdateResult := &client.App{ClientId: "existing-id", AccentColor: ""}
	existingApp := &client.App{ID: "internal-id", ClientId: "existing-id"}

	var updateCalledWith client.App
	var getAppCalledWith string
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
			return nil, errors.New("app already exists")
		},
		getAppByNameFn: func(clientName string) (*client.App, error) {
			return existingApp, nil
		},
		updateAppFn: func(app client.App) (*client.App, error) {
			updateCalledWith = app
			return incompleteUpdateResult, nil
		},
		getAppFn: func(clientId string) (*client.App, error) {
			getAppCalledWith = clientId
			return complete, nil
		},
	}

	plannedApp := &client.App{ClientName: "my-app"}
	got, warning, err := createOrUpsertApp(fake, plannedApp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if warning == "" {
		t.Fatal("expected a warning when falling back to update")
	}
	if updateCalledWith.ClientId != "existing-id" || updateCalledWith.ID != "internal-id" {
		t.Fatalf("expected UpdateApp to be called with the resolved existing app id, got %+v", updateCalledWith)
	}
	if getAppCalledWith != "existing-id" {
		t.Fatalf("expected GetApp to be called with %q, got %q", "existing-id", getAppCalledWith)
	}
	if !reflect.DeepEqual(got, complete) {
		t.Fatalf("expected state to be built from GetApp result %+v, got %+v", complete, got)
	}
}

func TestCreateOrUpsertApp_CreateErrorPropagates(t *testing.T) {
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
			return nil, errors.New("boom")
		},
	}

	_, _, err := createOrUpsertApp(fake, &client.App{})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestCreateOrUpsertApp_UpdateAfterCreateErrorPropagates(t *testing.T) {
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
			return &client.App{ClientId: "new-id"}, nil
		},
		updateAppFn: func(app client.App) (*client.App, error) {
			return nil, errors.New("boom")
		},
	}

	_, _, err := createOrUpsertApp(fake, &client.App{})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestCreateOrUpsertApp_GetAppNilErrorsInsteadOfPanicking(t *testing.T) {
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
			return &client.App{ClientId: "new-id"}, nil
		},
		updateAppFn: func(app client.App) (*client.App, error) {
			return &client.App{ClientId: "new-id"}, nil
		},
		getAppFn: func(clientId string) (*client.App, error) {
			return nil, nil
		},
	}

	_, _, err := createOrUpsertApp(fake, &client.App{})
	if err == nil {
		t.Fatal("expected an error when GetApp returns nil right after creation")
	}
}

func TestUpdateAndRefetchApp_BuildsStateFromGetAppAndPreservesAllowedOrigins(t *testing.T) {
	// Regression test for the same class of bug on Update, plus the pre-existing
	// AllowedOrigins workaround which must survive on top of the refetched state.
	incompleteUpdateResult := &client.App{ClientId: "client-id", AccentColor: ""}
	complete := &client.App{ClientId: "client-id", AccentColor: "#654321", AllowedOrigins: []string{"https://stale.example.com"}}

	var getAppCalledWith string
	fake := &fakeAppClient{
		updateAppFn: func(app client.App) (*client.App, error) {
			return incompleteUpdateResult, nil
		},
		getAppFn: func(clientId string) (*client.App, error) {
			getAppCalledWith = clientId
			return complete, nil
		},
	}

	plannedApp := &client.App{ClientId: "client-id", AllowedOrigins: []string{"https://fresh.example.com"}}
	got, err := updateAndRefetchApp(fake, plannedApp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if getAppCalledWith != "client-id" {
		t.Fatalf("expected GetApp to be called with %q, got %q", "client-id", getAppCalledWith)
	}
	if got.AccentColor != "#654321" {
		t.Fatalf("expected state to be built from GetApp result, got AccentColor=%q", got.AccentColor)
	}
	if !reflect.DeepEqual(got.AllowedOrigins, plannedApp.AllowedOrigins) {
		t.Fatalf("expected AllowedOrigins to be overridden with the planned value %v, got %v", plannedApp.AllowedOrigins, got.AllowedOrigins)
	}
}

func TestUpdateAndRefetchApp_UpdateErrorPropagates(t *testing.T) {
	fake := &fakeAppClient{
		updateAppFn: func(app client.App) (*client.App, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := updateAndRefetchApp(fake, &client.App{})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestNonInteractiveFieldViolations_OnlyAppliesToNonInteractive(t *testing.T) {
	allSet := map[string]bool{
		"accent_color":                      true,
		"primary_color":                     true,
		"hosted_page_group":                 true,
		"communication_medium_verification": true,
		"is_remember_me_selected":           true,
		"allowed_logout_urls":               true,
	}

	if got := nonInteractiveFieldViolations("SINGLE_PAGE", allSet); got != nil {
		t.Fatalf("expected no violations for an interactive client_type, got %v", got)
	}

	got := nonInteractiveFieldViolations("NON_INTERACTIVE", allSet)
	if len(got) != len(nonInteractiveUnsupportedFields) {
		t.Fatalf("expected all %d unsupported fields to be flagged, got %v", len(nonInteractiveUnsupportedFields), got)
	}
}

func TestNonInteractiveFieldViolations_OnlyFlagsFieldsThatAreSet(t *testing.T) {
	got := nonInteractiveFieldViolations("NON_INTERACTIVE", map[string]bool{
		"accent_color": true,
	})
	if len(got) != 1 || got[0] != "accent_color" {
		t.Fatalf("expected only accent_color to be flagged, got %v", got)
	}
}

func TestApplyAppToState_NonInteractiveAppsGetNullForUnsupportedFields(t *testing.T) {
	// Regression test: cidaas never persists these fields for NON_INTERACTIVE apps, so state
	// must report them as null (matching what ValidateConfig requires the config to leave
	// unset) rather than whatever stale/zero value happens to come back from the API.
	app := &client.App{
		ClientType:                      "NON_INTERACTIVE",
		AccentColor:                     "#leftover",
		PrimaryColor:                    "#leftover",
		HostedPageGroup:                 "leftover",
		CommunicationMediumVerification: "leftover",
		IsRememberMeSelected:            true,
		AllowedLogoutUrls:               []string{"https://leftover.example.com"},
		AppKey:                          &client.AppKey{},
	}

	var state App
	diags := applyAppToState(context.Background(), &state, app, true)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.AccentColor.IsNull() {
		t.Errorf("expected AccentColor to be null, got %v", state.AccentColor)
	}
	if !state.PrimaryColor.IsNull() {
		t.Errorf("expected PrimaryColor to be null, got %v", state.PrimaryColor)
	}
	if !state.HostedPageGroup.IsNull() {
		t.Errorf("expected HostedPageGroup to be null, got %v", state.HostedPageGroup)
	}
	if !state.CommunicationMediumVerification.IsNull() {
		t.Errorf("expected CommunicationMediumVerification to be null, got %v", state.CommunicationMediumVerification)
	}
	if !state.IsRememberMeSelected.IsNull() {
		t.Errorf("expected IsRememberMeSelected to be null, got %v", state.IsRememberMeSelected)
	}
	if !state.AllowedLogoutUrls.IsNull() {
		t.Errorf("expected AllowedLogoutUrls to be null, got %v", state.AllowedLogoutUrls)
	}
}

func TestApplyAppToState_InteractiveAppsKeepRealValues(t *testing.T) {
	app := &client.App{
		ClientType:                      "SINGLE_PAGE",
		AccentColor:                     "#ef4923",
		PrimaryColor:                    "#f7941d",
		HostedPageGroup:                 "default",
		CommunicationMediumVerification: "mobile_and_email_verification_required",
		IsRememberMeSelected:            true,
		AllowedLogoutUrls:               []string{"https://example.com/logout"},
		AppKey:                          &client.AppKey{},
	}

	var state App
	diags := applyAppToState(context.Background(), &state, app, true)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.AccentColor.ValueString() != "#ef4923" {
		t.Errorf("expected AccentColor to be preserved, got %v", state.AccentColor)
	}
	if state.HostedPageGroup.ValueString() != "default" {
		t.Errorf("expected HostedPageGroup to be preserved, got %v", state.HostedPageGroup)
	}
	if !state.IsRememberMeSelected.ValueBool() {
		t.Errorf("expected IsRememberMeSelected to be preserved as true, got %v", state.IsRememberMeSelected)
	}
	var gotLogoutUrls []string
	if diags := state.AllowedLogoutUrls.ElementsAs(context.Background(), &gotLogoutUrls, false); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(gotLogoutUrls) != 1 || gotLogoutUrls[0] != "https://example.com/logout" {
		t.Errorf("expected AllowedLogoutUrls to be preserved, got %v", gotLogoutUrls)
	}
}

func TestPlanToApp_UnknownAllowedLogoutUrlsDoesNotError(t *testing.T) {
	// Regression test: on a brand-new NON_INTERACTIVE resource with allowed_logout_urls left
	// unset (as ValidateConfig requires), the plan value is unknown, not null - there's no
	// prior state for UseStateForUnknown to carry forward. []string can't represent unknown,
	// so planToApp must not attempt the conversion in that case.
	groupObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_id":      types.StringType,
			"roles":         types.ListType{ElemType: types.StringType},
			"default_roles": types.ListType{ElemType: types.StringType},
		},
	}

	plan := &App{
		ClientType:              types.StringValue("NON_INTERACTIVE"),
		AllowedLogoutUrls:       types.ListUnknown(types.StringType),
		AllowedGroups:           types.ListNull(groupObjectType),
		OperationsAllowedGroups: types.ListNull(groupObjectType),
		TemplateGroupId:         types.StringNull(),
		PasswordPolicy:          types.StringNull(),
	}

	plannedApp, diags := planToApp(context.Background(), plan, plan, true)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if plannedApp.AllowedLogoutUrls != nil {
		t.Errorf("expected AllowedLogoutUrls to stay nil when the plan value is unknown, got %v", plannedApp.AllowedLogoutUrls)
	}
}
