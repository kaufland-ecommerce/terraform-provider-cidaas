package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultLocale = "en-us"

type hostedPagesRequest struct {
	Name string `json:"hosted_page_group"`
}

type hostedByLangRequest struct {
	HostedPageGroup string `json:"hosted_page_group"`
	HostedPageId    string `json:"hosted_page_id"`
	AcceptLanguage  string `json:"acceptLanguage"`
}

type hostedRequest struct {
	HostedPagesGroup string `json:"hosted_page_group"`
	HostedPageId     string `json:"hosted_page_id"`
	Locale           string `json:"locale"`
	Url              string `json:"url"`
}
type availablePagesResponse struct {
	Data []string `json:"data"`
}

type hostedByLangResponse struct {
	Data struct {
		HostedPageGroup string `json:"hosted_page_group"`
		HostedPageId    string `json:"hosted_page_id"`
		Url             string `json:"url"`
		Locale          string `json:"locale"`
	}
}

func (c *client) CreateHostedPagesGroup(group HostedPageGroup) error {
	rb, err := json.Marshal(hostedPagesRequest{Name: group.Name})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/hosted-srv/hostedgroup", c.HostUrl),
		strings.NewReader(string(rb)),
	)

	req.Header.Add("content-type", "application/json")

	if err != nil {
		return err
	}

	_, err = c.doRequest(req, nil)
	if err != nil {
		return err
	}

	for page, link := range group.Pages {
		rb, err := json.Marshal(hostedRequest{
			HostedPagesGroup: group.Name,
			HostedPageId:     page,
			Locale:           defaultLocale,
			Url:              link,
		})

		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("%s/hosted-srv/hosted", c.HostUrl),
			strings.NewReader(string(rb)),
		)

		req.Header.Add("content-type", "application/json")
		if err != nil {
			return err
		}

		_, err = c.doRequest(req, nil)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) GetHostedPagesGroup(groupName string) (*HostedPageGroup, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/hosted-srv/hosted/availablepages", c.HostUrl), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)

	if err != nil {
		return nil, err
	}

	var response availablePagesResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	var group = HostedPageGroup{
		Name:  groupName,
		Pages: map[string]string{},
	}

	for _, page := range response.Data {
		// Hack to make sure there isn't something else set for german languages or other english ones
		rb, err := json.Marshal(hostedByLangRequest{
			HostedPageGroup: groupName,
			HostedPageId:    page,
			AcceptLanguage:  "de",
		})

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("%s/hosted-srv/hosted/bylang", c.HostUrl),
			strings.NewReader(string(rb)),
		)

		req.Header.Add("content-type", "application/json")

		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req, nil)
		if err != nil {
			return nil, err
		}

		if len(body) == 0 {
			continue
		}

		var response hostedByLangResponse

		err = json.Unmarshal(body, &response)

		if err != nil {
			return nil, err
		}

		if response.Data.Url == "" {
			continue
		}

		group.Pages[response.Data.HostedPageId] = response.Data.Url
	}

	return &group, nil
}

func (c *client) UpdateHostedPagesGroup(group HostedPageGroup) error {
	for page, link := range group.Pages {
		rb, err := json.Marshal(hostedRequest{
			HostedPagesGroup: group.Name,
			HostedPageId:     page,
			Locale:           defaultLocale,
			Url:              link,
		})

		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("%s/hosted-srv/hosted", c.HostUrl),
			strings.NewReader(string(rb)),
		)

		req.Header.Add("content-type", "application/json")
		if err != nil {
			return err
		}

		_, err = c.doRequest(req, nil)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) DeleteHostedPagesGroup(groupName string) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/hosted-srv/hostedgroup?groupname=%s", c.HostUrl, groupName),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req, nil)

	return err
}
