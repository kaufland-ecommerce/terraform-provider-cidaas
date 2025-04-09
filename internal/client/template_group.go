package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type templateGroupResponse struct {
	Data TemplateGroup `json:"data"`
}

func (c *client) CreateTemplateGroup(group CreateTemplateGroupRequest) (*TemplateGroup, error) {
	rb, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/notifications-srv/templategroups", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var templateGroup TemplateGroup
	err = json.Unmarshal(resp, &templateGroup)

	if err != nil {
		return nil, err
	}

	return &templateGroup, nil
}

func (c *client) UpdateTemplateGroup(group TemplateGroup) (*TemplateGroup, error) {
	groupID := group.ID
	group.ID = ""
	rb, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/notifications-srv/templategroups/%s", c.HostUrl, groupID),
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
	//return nil, fmt.Errorf("Update template group response:" + string(resp))

	var groupResponse templateGroupResponse
	err = json.Unmarshal(resp, &groupResponse)

	if err != nil {
		return nil, err
	}

	return &groupResponse.Data, nil
}

func (c *client) GetTemplateGroup(groupId string) (*TemplateGroup, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/notifications-srv/templategroups/%s", c.HostUrl, groupId),
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

	var group templateGroupResponse
	err = json.Unmarshal(resp, &group)

	if err != nil {
		return nil, err
	}

	return &group.Data, nil
}

func (c *client) DeleteTemplateGroup(groupId string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/notifications-srv/templategroups/%s", c.HostUrl, groupId),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)

	return err
}
