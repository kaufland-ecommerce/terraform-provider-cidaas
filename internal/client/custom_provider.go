package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type customProvidersResponse struct {
	Status int            `json:"status"`
	Data   CustomProvider `json:"data"`
}

func (c *client) GetCustomProvider(providerName string) (*CustomProvider, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/providers-srv/custom/%s", c.HostUrl, providerName), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response customProvidersResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil

}
