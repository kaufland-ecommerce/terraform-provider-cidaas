package provider

import (
	"errors"
	"reflect"
	"testing"

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
	// Regression test: the raw POST /apps-srv/clients response is not guaranteed to reflect
	// the full persisted object, so the resource must re-fetch by ID before building state.
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

	var getAppCalledWith string
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
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

func TestCreateOrUpsertApp_GetAppNilErrorsInsteadOfPanicking(t *testing.T) {
	fake := &fakeAppClient{
		createAppFn: func(app *client.App) (*client.App, error) {
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
