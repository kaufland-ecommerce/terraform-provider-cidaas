package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *client) SignIn() (*authResponse, error) {
	rb, err := json.Marshal(c.Credentials)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/token-srv/token", c.HostUrl),
		bytes.NewReader(rb),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode > http.StatusOK {
		return nil, fmt.Errorf("auth failed: %s", res.Body)
	}

	var response authResponse
	err = json.NewDecoder(res.Body).Decode(&response)

	return &response, err
}
