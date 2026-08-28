package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type notificationServiceSetupResponse struct {
	Success bool                     `json:"success"`
	Status  int                      `json:"status"`
	Data    NotificationServiceSetup `json:"data"`
}

func (c *client) CreateNotificationServiceSetup(setup NotificationServiceSetup) (*NotificationServiceSetup, error) {
	rb, err := json.Marshal(setup)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/notifications-srv/servicesetups", c.HostUrl),
		strings.NewReader(string(rb)),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var setupResponse notificationServiceSetupResponse
	if err := json.Unmarshal(resp, &setupResponse); err != nil {
		return nil, err
	}

	return &setupResponse.Data, nil
}

func (c *client) GetNotificationServiceSetup(id string) (*NotificationServiceSetup, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/notifications-srv/servicesetups/%s", c.HostUrl, id),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var setupResponse notificationServiceSetupResponse
	if err := json.Unmarshal(resp, &setupResponse); err != nil {
		return nil, err
	}

	return &setupResponse.Data, nil
}

// UpdateNotificationServiceSetup only sends name and description - the API only allows
// those two fields to change on PATCH; everything else (service_id, communication_methods)
// requires replacement.
func (c *client) UpdateNotificationServiceSetup(id string, name string, description string) (*NotificationServiceSetup, error) {
	rb, err := json.Marshal(struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}{Name: name, Description: description})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/notifications-srv/servicesetups/%s", c.HostUrl, id),
		strings.NewReader(string(rb)),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var setupResponse notificationServiceSetupResponse
	if err := json.Unmarshal(resp, &setupResponse); err != nil {
		return nil, err
	}

	return &setupResponse.Data, nil
}

func (c *client) DeleteNotificationServiceSetup(id string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/notifications-srv/servicesetups/%s", c.HostUrl, id),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
