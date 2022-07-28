package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type tenantInfoResponse struct {
	Status int        `json:"status"`
	Data   TenantInfo `json:"data"`
}

func (c *client) GetTenantInfo() (*TenantInfo, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/public-srv/tenantinfo/basic", c.HostUrl), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)

	if err != nil {
		return nil, err
	}

	var response tenantInfoResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}
