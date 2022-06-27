package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type appResponse struct {
	Status int `json:"status"`
	Data   App `json:"data"`
}

func (c *client) GetApp(clientId string) (*App, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/apps-srv/clients/%s", c.HostUrl, clientId), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)

	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, nil
	}

	var response appResponse

	err = json.Unmarshal(body, &response)

	return &response.Data, err
}

func (c *client) UpdateApp(app App) (*App, error) {
	rb, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/apps-srv/clients", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)

	if err != nil {
		return nil, err
	}

	var response appResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	if *response.Data.PasswordPolicy == "" {
		response.Data.PasswordPolicy = nil
	}

	return &response.Data, nil
}

func (c *client) DeleteApp(clientId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/apps-srv/clients/%s", c.HostUrl, clientId), nil)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req, nil)

	return err
}
