package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type templateRequestCreationCopyLocale struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type templateCreationRequestCopyFrom struct {
	FromGroupId string                            `json:"fromGroupId"`
	Locale      templateRequestCreationCopyLocale `json:"locale"`
}

// @TODO: Probably this also needs to be adjusted to work with the new endpoint
type templateGroupCreationRequest struct {
	Id            string                          `json:"id"`
	Description   string                          `json:"description"`
	DefaultLocale string                          `json:"defaultLocale"`
	Copy          templateCreationRequestCopyFrom `json:"copy"`
}

type templateGroupResponse struct {
	Data TemplateGroup `json:"data"`
}

func (c *client) CreateTemplateGroup(groupId string) (*TemplateGroup, error) {
	rb, err := json.Marshal(templateGroupCreationRequest{Id: groupId})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/notifications-srv/templategroups/%s", c.HostUrl, groupId),
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

func (c *client) UpdateTemplateGroup(group *TemplateGroup) error {
	rb, err := json.Marshal(group)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/templates-srv/groups/%s", c.HostUrl, group.Id),
		strings.NewReader(string(rb)),
	)

	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")

	resp, err := c.doRequest(req)

	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, group)

	if err != nil {
		return err
	}

	return nil
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
