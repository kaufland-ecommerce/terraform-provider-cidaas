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
