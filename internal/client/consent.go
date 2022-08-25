package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type consentsInstancesResponse struct {
	Status int               `json:"status"`
	Data   []ConsentInstance `json:"data"`
}

func (c *client) GetConsentInstance(name string) (*ConsentInstance, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/consent-management-srv/v2/consent/instance/all/list", c.HostUrl),
		nil,
	)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response consentsInstancesResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, errors.New("could not find any consent instances")
	}

	for _, el := range response.Data {
		if el.ConsentName == name {
			return &el, nil
		}
	}

	return nil, errors.New("consent instance could not be located")
}
