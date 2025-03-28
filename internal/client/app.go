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

func (c *client) CreateApp(app *App) (*App, error) {
	rb, err := json.Marshal(app)
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
	err = json.Unmarshal(body, &response)

	err = c.prepareResponse(&response.Data)

	if err != nil {
		return nil, err
	}

	return &response.Data, err
}

func (c *client) UpdateApp(app App) (*App, error) {
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
