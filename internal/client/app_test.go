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

// v4AppResponse is the shape cidaas v4 returns from GET and PUT /apps-srv/clients:
// basic_settings.client_secrets is a list of objects.
const v4AppResponse = `{
  "success": true,
  "status": 200,
  "data": {
    "_id": "app-id",
    "client_id": "app-client-id",
    "client_secret": "secret-1",
    "client_secret_expires_at": null,
    "client_name": "my-app",
    "client_type": "NON_INTERACTIVE",
    "allow_login_with": [],
    "enable_passwordless_auth": null,
    "allow_disposable_email": null,
    "redirect_uris": [],
    "response_types": [],
    "grant_types": ["client_credentials"],
    "allowed_scopes": ["openid"],
    "password_policy_ref": "default",
    "post_logout_redirect_uris": [],
    "basic_settings": {
      "client_id": "app-client-id",
      "token_endpoint_auth_method": "client_secret_post",
      "redirect_uris": [],
      "allowed_logout_urls": [],
      "app_owner": "CLIENT",
      "allowed_scopes": ["openid"],
      "client_secrets": [
        {"client_secret": "secret-1", "client_secret_expires_at": null, "client_secret_issued_at": 1651481993812}
      ]
    }
  }
}`

func TestCreateAndUpdateApp_DoNotSendBasicSettings(t *testing.T) {
	var receivedBodies []map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		receivedBodies = append(receivedBodies, body)
		writeAppResponse(t, w, App{ClientId: "new-client-id"})
	})

	app := App{
		ClientName:    "my-app",
		ClientId:      "new-client-id",
		RedirectUris:  []string{"https://example.com/cb"},
		AllowedScopes: []string{"openid"},
	}
	if _, err := c.CreateApp(&app); err != nil {
		t.Fatalf("CreateApp: unexpected error: %v", err)
	}
	if _, err := c.UpdateApp(app); err != nil {
		t.Fatalf("UpdateApp: unexpected error: %v", err)
	}

	for i, body := range receivedBodies {
		if _, ok := body["basic_settings"]; ok {
			t.Errorf("request %d: expected no basic_settings, got %v", i, body["basic_settings"])
		}
	}
}

func TestGetAndUpdateApp_DecodeV4Response(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(v4AppResponse))
	})

	got, err := c.GetApp("app-client-id")
	if err != nil {
		t.Fatalf("GetApp: unexpected error: %v", err)
	}
	if got.ClientSecret != "secret-1" {
		t.Errorf("expected client_secret to be read, got %q", got.ClientSecret)
	}
	if got.PasswordPolicy == nil || *got.PasswordPolicy != "default" {
		t.Errorf("expected password_policy_ref \"default\", got %v", got.PasswordPolicy)
	}

	if _, err := c.UpdateApp(*got); err != nil {
		t.Fatalf("UpdateApp: unexpected error: %v", err)
	}
}

func TestGetApp_ReturnsDecodeError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status": 200, "data": {"client_id": "app-client-id", "redirect_uris": "not-a-list"}}`))
	})

	if _, err := c.GetApp("app-client-id"); err == nil {
		t.Fatal("expected a decode error, got nil")
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
