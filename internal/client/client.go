package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var _ Client = (*client)(nil)

func NewClient(host *string, clientId *string, clientSecret *string) (Client, error) {
	c := client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostUrl:    *host,
	}

	c.Credentials = authStruct{
		GrantType:    "client_credentials",
		ClientId:     *clientId,
		ClientSecret: *clientSecret,
	}

	res, err := c.SignIn()

	if err != nil {
		return nil, err
	}

	c.Token = res.Token

	return &c, nil
}

type Client interface {
	GetHooks() ([]*Hook, error)
	GetHook(ID string) (*Hook, error)
	UpsertHook(hook Hook) (*Hook, error)
	DeleteHook(ID string) error

	GetSocialProvider(name string) (*SocialProvider, error)

	GetConsentInstance(name string) (*ConsentInstance, error)

	UpdatePasswordPolicy(policy PasswordPolicy) (*PasswordPolicy, error)
	GetPasswordPolicy(id string) (*PasswordPolicy, error)
	GetPasswordPolicyByName(name string) (*PasswordPolicy, error)
	DeletePasswordPolicy(id string) error

	GetTenantInfo() (*TenantInfo, error)

	CreateHostedPagesGroup(group HostedPageGroup) error
	UpdateHostedPagesGroup(group HostedPageGroup) error
	DeleteHostedPagesGroup(groupName string) error
	GetHostedPagesGroup(groupName string) (*HostedPageGroup, error)

	CreateApp(app *App) (*App, error)
	GetApp(ClientId string) (*App, error)
	UpdateApp(app App) (*App, error)
	DeleteApp(ID string) error

	GetRegistrationField(key string) (*RegistrationField, error)
	UpsertRegistrationField(field *RegistrationField) error
	DeleteRegistrationField(key string) error

	CreateTemplateGroup(group string) (*TemplateGroup, error)
	GetTemplateGroup(groupId string) (*TemplateGroup, error)
	UpdateTemplateGroup(group *TemplateGroup) error
	DeleteTemplateGroup(groupId string) error

	UpdateTemplate(template Template) (*Template, error)
	GetTemplate(template Template) (*Template, error)
}

type client struct {
	HTTPClient  *http.Client
	HostUrl     string
	Token       string
	Credentials authStruct
}

type authStruct struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type authResponse struct {
	Token string `json:"access_token"`
}

func (c *client) doRequest(req *http.Request) ([]byte, error) {
	token := c.Token

	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		return body, err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil, err
	}

	return nil, fmt.Errorf("status %d, body %s", res.StatusCode, body)
}
