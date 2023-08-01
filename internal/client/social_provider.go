package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type socialProvidersResponse struct {
	Status int              `json:"status"`
	Data   []SocialProvider `json:"data"`
}

func (c *client) GetSocialProvider(providerName string, name string) (*SocialProvider, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/providers-srv/multi/providers/list?provider_name=%s&provider_type=system", c.HostUrl, providerName), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response socialProvidersResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	for _, provider := range response.Data {
		if provider.Name == name {
			return &provider, nil
		}
	}

	if len(response.Data) != 1 {
		return nil, errors.New("could not identify provider")
	}

	return &response.Data[0], nil

}
