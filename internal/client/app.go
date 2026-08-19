package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type appResponse struct {
	Status int `json:"status"`
	Data   App `json:"data"`
}

type appListResponse struct {
	Success bool  `json:"success"`
	Status  int   `json:"status"`
	Data    []App `json:"data"`
}

// basicSettingsFor builds the nested basic_settings sub-object the cidaas admin UI always
// sends alongside the top-level fields it mirrors. allowed_logout_urls in particular seems to
// be authoritative from here: writes that only set the top-level field were silently never
// persisted, confirmed both via our own request/response captures and via a real UI PUT
// captured from the browser.
func basicSettingsFor(app App) *BasicSettings {
	return &BasicSettings{
		ClientId:                app.ClientId,
		TokenEndpointAuthMethod: "client_secret_post",
		RedirectUris:            app.RedirectUris,
		AllowedLogoutUrls:       app.AllowedLogoutUrls,
		AppOwner:                app.AppOwner,
		AllowedScopes:           app.AllowedScopes,
		ClientSecrets:           []string{},
	}
}

func (c *client) CreateApp(app *App) (*App, error) {
	toSend := *app

	// cidaas's create endpoint rejects the request with "missing primaryColor or accentColor"
	// (APP10015) if both are empty, even for NON_INTERACTIVE apps - even though it never
	// actually persists either field for that client_type (confirmed via a follow-up GetApp).
	// Fill in a harmless placeholder here rather than exposing this as something practitioners
	// need to configure; it has no visible effect since it's never stored or read back.
	// "#000000" alone did not satisfy the validation (still rejected as "missing"), so this
	// uses a real value that's already accepted across the org's existing apps.
	if toSend.ClientType == "NON_INTERACTIVE" && toSend.AccentColor == "" && toSend.PrimaryColor == "" {
		toSend.AccentColor = "#ef4923"
		toSend.PrimaryColor = "#f7941d"
	}

	toSend.BasicSettings = basicSettingsFor(toSend)

	rb, err := json.Marshal(toSend)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/apps-srv/clients", c.HostUrl),
		bytes.NewReader(rb),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response appResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	err = c.prepareResponse(&response.Data)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) GetApp(clientId string) (*App, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/apps-srv/clients/%s", c.HostUrl, clientId), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, nil
	}

	var response appResponse
	_ = json.Unmarshal(body, &response)

	err = c.prepareResponse(&response.Data)

	if err != nil {
		return nil, err
	}

	return &response.Data, err
}

func (c *client) GetAppByName(clientName string) (*App, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/apps-srv/clients/list", c.HostUrl), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, nil
	}

	var response appListResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	// Search for the app with matching client_name
	for _, app := range response.Data {
		if app.ClientName == clientName {
			// Found matching app
			appCopy := app // Create a copy to avoid returning a reference to a loop variable
			err = c.prepareResponse(&appCopy)
			if err != nil {
				return nil, err
			}
			return &appCopy, nil
		}
	}

	// No matching app found
	return nil, fmt.Errorf("no app found with client_name: %s", clientName)
}

func (c *client) UpdateApp(app App) (*App, error) {
	app.BasicSettings = basicSettingsFor(app)

	rb, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/apps-srv/clients", c.HostUrl),
		bytes.NewReader(rb),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response appResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	err = c.prepareResponse(&response.Data)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) DeleteApp(clientId string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/apps-srv/clients/%s", c.HostUrl, clientId), nil)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)

	return err
}

func (c *client) prepareResponse(app *App) error {

	if app.PasswordPolicy == nil || *app.PasswordPolicy == "" {
		app.PasswordPolicy = nil
	}

	// App creation does not return those if they are empty on the initial creation
	if app.OperationsAllowedGroups == nil {
		app.OperationsAllowedGroups = []AllowedGroup{}
	} else {
		for i := range app.OperationsAllowedGroups {
			if app.OperationsAllowedGroups[i].DefaultRoles == nil {
				app.OperationsAllowedGroups[i].DefaultRoles = []string{}
			}
		}
	}

	if app.AllowedGroups == nil {
		app.AllowedGroups = []AllowedGroup{}
	}

	if app.AllowedMfa == nil {
		app.AllowedMfa = []string{}
	}

	if app.ConsentRefs == nil {
		app.ConsentRefs = []string{}
	}

	if app.AllowedOrigins == nil {
		app.AllowedOrigins = []string{}
	}

	if app.AllowedFields == nil {
		app.AllowedFields = []string{}
	}

	if app.AllowedWebOrigins == nil {
		app.AllowedWebOrigins = []string{}
	}

	if app.RequiredFields == nil {
		app.RequiredFields = []string{}
	}

	if app.AdditionalAccessTokenPayload == nil {
		app.AdditionalAccessTokenPayload = []string{}
	}

	if app.RedirectUris == nil {
		app.RedirectUris = []string{}
	}

	if app.AllowedLogoutUrls == nil {
		app.AllowedLogoutUrls = []string{}
	}

	return nil
}
