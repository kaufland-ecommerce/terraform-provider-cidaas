package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *client) SignIn() (*authResponse, error) {
	rb, err := json.Marshal(c.Credentials)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/token-srv/token", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	req.Header.Add("content-type", "application/json")

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	ar := authResponse{}
	err = json.Unmarshal(body, &ar)

	if err != nil {
		return nil, err
	}

	return &ar, nil
}
