package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type passwordPolicyResponse struct {
	Status int            `json:"status"`
	Data   PasswordPolicy `json:"data"`
}

type passwordPoliciesResponse struct {
	Status int              `json:"status"`
	Data   []PasswordPolicy `json:"data"`
}

func (c *client) GetPasswordPolicy(id string) (*PasswordPolicy, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/password-policy-srv/policy/%s", c.HostUrl, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	var response passwordPolicyResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *client) DeletePasswordPolicy(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/password-policy-srv/policy/%s", c.HostUrl, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, nil)

	if err != nil {
		return err
	}

	return nil
}

func (c *client) UpdatePasswordPolicy(policy PasswordPolicy) (*PasswordPolicy, error) {
	rb, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/password-policy-srv/policy", c.HostUrl),
		bytes.NewReader(rb),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	var response passwordPolicyResponse
	err = json.Unmarshal(body, &response)

	return &response.Data, err
}

func (c *client) GetPasswordPolicyByName(name string) (*PasswordPolicy, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/password-policy-srv/policy/list", c.HostUrl), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)

	if err != nil {
		return nil, err
	}

	var response passwordPoliciesResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, errors.New("could not find any policies")
	}

	for _, el := range response.Data {
		if el.PolicyName == name {
			return &el, nil
		}
	}

	return nil, errors.New("could not find specified provider")
}
