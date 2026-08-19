package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &client{
		HTTPClient: server.Client(),
		HostUrl:    server.URL,
		Token:      "test-token",
	}
}

func writeAppResponse(t *testing.T, w http.ResponseWriter, app App) {
	t.Helper()
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(appResponse{Status: 200, Data: app})
}

func TestCreateApp_FillsPlaceholderColorForNonInteractiveAppsWithNoColorSet(t *testing.T) {
	// Regression test: cidaas's create endpoint rejects the request with "missing
	// primaryColor or accentColor" (APP10015) when both are empty, even for NON_INTERACTIVE
	// apps that never persist either field - the caller (plannedApp) must not need to set one
	// just to satisfy this quirk.
	var receivedBody map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		writeAppResponse(t, w, App{ClientId: "new-client-id"})
	})

	original := &App{ClientName: "my-app", ClientType: "NON_INTERACTIVE"}
	_, err := c.CreateApp(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v, _ := receivedBody["accentColor"].(string); v == "" {
		t.Fatalf("expected a non-empty placeholder accentColor to be sent, got %v", receivedBody["accentColor"])
	}
	if original.AccentColor != "" {
		t.Fatalf("expected the caller's App not to be mutated, got AccentColor=%q", original.AccentColor)
	}
}

func TestCreateApp_SendsBasicSettingsMirroringTopLevelFields(t *testing.T) {
	// Regression test: allowed_logout_urls is silently never persisted unless basic_settings
	// carries the same value too - confirmed via captured request/response pairs and a real UI
	// PUT captured from the browser, both showing basic_settings.allowed_logout_urls kept in
	// sync with the top-level field on every write.
	var receivedBody map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		writeAppResponse(t, w, App{ClientId: "new-client-id"})
	})

	_, err := c.CreateApp(&App{
		ClientName:        "my-app",
		ClientId:          "new-client-id",
		AllowedLogoutUrls: []string{"https://example.com/logout"},
		RedirectUris:      []string{"https://example.com/cb"},
		AllowedScopes:     []string{"openid"},
		AppOwner:          "CLIENT",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	basicSettings, ok := receivedBody["basic_settings"].(map[string]any)
	if !ok {
		t.Fatalf("expected basic_settings to be sent, got %v", receivedBody["basic_settings"])
	}
	if basicSettings["client_id"] != "new-client-id" {
		t.Errorf("expected basic_settings.client_id to be set, got %v", basicSettings["client_id"])
	}
	logoutUrls, _ := basicSettings["allowed_logout_urls"].([]any)
	if len(logoutUrls) != 1 || logoutUrls[0] != "https://example.com/logout" {
		t.Errorf("expected basic_settings.allowed_logout_urls to mirror the top-level value, got %v", basicSettings["allowed_logout_urls"])
	}
}

func TestCreateApp_DoesNotOverrideAnExplicitColor(t *testing.T) {
	var receivedBody map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		writeAppResponse(t, w, App{ClientId: "new-client-id"})
	})

	_, err := c.CreateApp(&App{ClientName: "my-app", ClientType: "SINGLE_PAGE", AccentColor: "#123456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["accentColor"] != "#123456" {
		t.Fatalf("expected the explicit accentColor to be preserved, got %v", receivedBody["accentColor"])
	}
}

func TestCreateApp_ReturnsExactlyWhatTheServerSent(t *testing.T) {
	// CreateApp must not silently invent data beyond prepareResponse's documented
	// nil->[] coercion; any incompleteness in the API's create response has to be
	// handled by the caller doing a follow-up GetApp, not papered over here.
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeAppResponse(t, w, App{
			ClientId:    "new-client-id",
			AccentColor: "",
		})
	})

	got, err := c.CreateApp(&App{ClientName: "my-app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ClientId != "new-client-id" {
		t.Fatalf("expected client id to be passed through, got %q", got.ClientId)
	}
	if got.AccentColor != "" {
		t.Fatalf("expected AccentColor to remain empty as sent by the server, got %q", got.AccentColor)
	}
}

func TestGetApp_ReturnsFullyPopulatedApp(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeAppResponse(t, w, App{
			ClientId:     "existing-client-id",
			AccentColor:  "#123456",
			RedirectUris: []string{"https://example.com/cb"},
		})
	})

	got, err := c.GetApp("existing-client-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccentColor != "#123456" {
		t.Fatalf("expected AccentColor to be populated, got %q", got.AccentColor)
	}
	if len(got.RedirectUris) != 1 || got.RedirectUris[0] != "https://example.com/cb" {
		t.Fatalf("expected RedirectUris to be populated, got %v", got.RedirectUris)
	}
}

func TestPrepareResponse_CoercesNilListsToEmpty(t *testing.T) {
	app := &App{}

	if err := (&client{}).prepareResponse(app); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nilCoercedListFields := map[string]any{
		"OperationsAllowedGroups":      app.OperationsAllowedGroups,
		"AllowedGroups":                app.AllowedGroups,
		"AllowedMfa":                   app.AllowedMfa,
		"ConsentRefs":                  app.ConsentRefs,
		"AllowedOrigins":               app.AllowedOrigins,
		"AllowedFields":                app.AllowedFields,
		"AllowedWebOrigins":            app.AllowedWebOrigins,
		"RequiredFields":               app.RequiredFields,
		"AdditionalAccessTokenPayload": app.AdditionalAccessTokenPayload,
		"RedirectUris":                 app.RedirectUris,
		"AllowedLogoutUrls":            app.AllowedLogoutUrls,
	}

	for name, field := range nilCoercedListFields {
		if field == nil {
			t.Errorf("expected %s to be coerced from nil to an empty slice", name)
		}
	}
}

func TestPrepareResponse_EmptyPasswordPolicyBecomesNil(t *testing.T) {
	empty := ""
	app := &App{PasswordPolicy: &empty}

	if err := (&client{}).prepareResponse(app); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app.PasswordPolicy != nil {
		t.Fatalf("expected empty PasswordPolicy to become nil, got %v", *app.PasswordPolicy)
	}
}
