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

func (c *client) GetSocialProvider(name string) (*SocialProvider, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/providers-srv/multi/providers/list?provider_name=%s&provider_type=system", c.HostUrl, name), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)

	if err != nil {
		return nil, err
	}

	var response socialProvidersResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	if len(response.Data) != 1 {
		return nil, errors.New("could not identify provider")
	}

	return &response.Data[0], nil

}
