package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type templateGroupCreationRequest struct {
	GroupId string `json:"group_id"`
}

type templateGroupResponse struct {
	Data TemplateGroup `json:"data"`
}

func (c *client) CreateTemplateGroup(groupId string) (*TemplateGroup, error) {
	rb, err := json.Marshal(templateGroupCreationRequest{GroupId: groupId})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/templates-srv/groups", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	_, err = c.doRequest(req)

	if err != nil {
		return nil, err
	}

	return &TemplateGroup{GroupId: groupId}, nil
}

func (c *client) GetTemplateGroup(groupId string) (*TemplateGroup, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/templates-srv/groups/%s", c.HostUrl, groupId),
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
		fmt.Sprintf("%s/templates-srv/groups/%s", c.HostUrl, groupId),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)

	return err
}
