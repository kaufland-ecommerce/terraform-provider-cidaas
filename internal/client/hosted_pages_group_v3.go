package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type hpgroupResponse struct {
	Status int
	Data   HostedPageGroupV3
}

func (c *client) UpsertHostedPagesGroupV3(group HostedPageGroupV3) (*HostedPageGroupV3, error) {
	rb, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/hostedpages-srv/hpgroup", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response hpgroupResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) GetHostedPagesGroupV3(id string) (*HostedPageGroupV3, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/hostedpages-srv/hpgroup/%s", c.HostUrl, id),
		nil,
	)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	var response hpgroupResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) DeleteHostedPagesGroupV3(id string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/hostedpages-srv/hpgroup/%s", c.HostUrl, id),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)

	return err
}
