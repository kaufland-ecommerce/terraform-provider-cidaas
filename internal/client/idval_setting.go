package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type idValSettingResponse struct {
	Success bool
	Status  int
	Data    IdValSetting
}

func (c *client) UpsertIdValSetting(setting IdValSetting) (*IdValSetting, error) {
	rb, err := json.Marshal(setting)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/id-val-srv/idvalsettings", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response idValSettingResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) GetIdValSetting(id string) (*IdValSetting, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/id-val-srv/idvalsettings/%s", c.HostUrl, id),
		nil,
	)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)

	if err != nil {
		return nil, err
	}

	var response idValSettingResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) DeleteIdValSetting(id string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/id-val-srv/idvalsettings/%s", c.HostUrl, id),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)

	return err
}
