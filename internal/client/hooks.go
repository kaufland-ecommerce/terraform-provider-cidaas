package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type hookResponse struct {
	Status int  `json:"status"`
	Data   Hook `json:"data"`
}

type hooksResponse struct {
	Status int    `json:"status"`
	Data   []Hook `json:"data"`
}

func (c *client) GetHooks() ([]*Hook, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/webhooks-srv/webhook/list", c.HostUrl), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response hooksResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	hooks := make([]*Hook, len(response.Data))

	for i, h := range response.Data {
		var err error
		hooks[i], err = c.GetHook(h.Id)

		if err != nil {
			return nil, err
		}
	}

	return hooks, nil
}

func (c *client) GetHook(ID string) (*Hook, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/webhook-srv/webhook?id=%s", c.HostUrl, ID),
		nil,
	)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response hookResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) UpsertHook(hook Hook) (*Hook, error) {
	rb, err := json.Marshal(hook)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/webhook-srv/webhook", c.HostUrl),
		bytes.NewReader(rb),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response hookResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) DeleteHook(ID string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/webhook-srv/webhook/%s", c.HostUrl, ID),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)

	if err != nil {
		return err
	}

	return nil
}
