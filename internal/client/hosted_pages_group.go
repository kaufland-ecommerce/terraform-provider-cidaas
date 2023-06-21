package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type hpgroupResponse struct {
	Status int
	Data   HostedPageGroup
}

func (c *client) UpsertHostedPagesGroup(group HostedPageGroup) (*HostedPageGroup, error) {
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

func (c *client) GetHostedPagesGroup(id string) (*HostedPageGroup, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/hostedpages-srv/hpgroup/%s", c.HostUrl, id),
		nil,
	)

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

func (c *client) DeleteHostedPagesGroup(id string) error {
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
