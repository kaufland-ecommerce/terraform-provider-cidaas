package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type templateResponse struct {
	Data Template
}

func (c *client) GetTemplate(template Template) (*Template, error) {
	rb, err := json.Marshal(template)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/templates-srv/template/find", c.HostUrl),
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

	var templateResponse templateResponse
	err = json.Unmarshal(resp, &templateResponse)
	if err != nil {
		return nil, err
	}

	return &templateResponse.Data, nil
}

func (c *client) UpdateTemplate(template Template) (*Template, error) {
	rb, err := json.Marshal(template)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/templates-srv/template", c.HostUrl),
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

	var templateResponse templateResponse
	err = json.Unmarshal(resp, &templateResponse)
	if err != nil {
		return nil, err
	}

	return &templateResponse.Data, nil
}
