package client

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type Theme struct {
	Name string
	Css  string
}

// UpsertTheme creates or overwrites a Hosted Pages theme. Unlike every other resource in
// this package, hostedpages-srv/themes takes a multipart/form-data upload (field name
// "theme", filename "<name>.css", part content-type text/css) rather than a JSON body -
// confirmed by capturing a real "Create Theme" submission from the Admin UI.
func (c *client) UpsertTheme(theme Theme) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     "theme",
		"filename": theme.Name + ".css",
	}))
	header.Set("Content-Type", "text/css")

	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(part, theme.Css); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hostedpages-srv/themes", c.HostUrl), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, err = c.doRequest(req)
	return err
}

// GetTheme fetches a theme's CSS by name. hostedpages-srv/themes/{name} returns the raw
// CSS as text/css, not a JSON envelope like the rest of this API.
func (c *client) GetTheme(name string) (*Theme, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/hostedpages-srv/themes/%s", c.HostUrl, name), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return &Theme{Name: name, Css: string(body)}, nil
}

func (c *client) DeleteTheme(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/hostedpages-srv/themes/%s", c.HostUrl, name), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
