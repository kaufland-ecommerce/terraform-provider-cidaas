package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type registrationFieldResponse struct {
	Status int                    `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

var registrationFieldBaseTypes = map[string]string{
	"CONSENT": "bool",
}

func (c *client) GetRegistrationField(key string) (*RegistrationField, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/registration-setup-srv/fields/flat/field/%s", c.HostUrl, key),
		nil,
	)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	var response registrationFieldResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	id := response.Data["id"].(string)

	consentRefsRaw := response.Data["consent_refs"].([]interface{})

	consentRefs := make([]string, len(consentRefsRaw))

	for i, _ := range consentRefsRaw {
		consentRefs[i] = consentRefsRaw[i].(string)
	}

	field := RegistrationField{
		ReadOnly:      response.Data["readOnly"].(bool),
		Claimable:     response.Data["claimable"].(bool),
		Required:      response.Data["required"].(bool),
		Enabled:       response.Data["enabled"].(bool),
		ParentGroupID: response.Data["parent_group_id"].(string),
		ConsentRefs:   consentRefs,
		ID:            &id,
		FieldKey:      response.Data["fieldKey"].(string),
		DataType:      response.Data["dataType"].(string),
		Order:         int64(response.Data["order"].(float64)),
	}

	return &field, nil
}

func (c *client) UpsertRegistrationField(field *RegistrationField) error {
	_ = field.calculateFields()

	rb, err := json.Marshal(field)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/registration-setup-srv/fields", c.HostUrl),
		bytes.NewReader(rb),
	)
	req.Header.Add("content-type", "application/json")

	body, err := c.doRequest(req, nil)
	if err != nil {
		return err
	}

	var response registrationFieldResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	id := response.Data["id"].(string)

	field.ID = &id

	return nil
}

func (c *client) DeleteRegistrationField(key string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/registration-setup-srv/fields/%s", c.HostUrl, key),
		nil,
	)

	if err != nil {
		return err
	}

	_, err = c.doRequest(req, nil)

	return err
}

func (rf *RegistrationField) calculateFields() error {

	rf.BaseDataType = registrationFieldBaseTypes[rf.DataType]

	rf.IsGroup = false
	rf.IsList = false

	rf.FieldDefinition = FieldDefinition{
		Language: "de",
		Locale:   "de-DE",
	}

	rf.FieldType = "CUSTOM"
	rf.Scopes = []string{}

	rf.LocaleText = LocaleText{
		Locale:   "de-DE",
		Language: "de",
		ConsentLabel: ConsentLabel{
			Label:     fmt.Sprintf("<a href=\"%s\">Consent</a>", rf.ConsentRefs[0]),
			LabelText: fmt.Sprintf("<a href=\"%s\">Consent</a>", rf.ConsentRefs[0]),
		},
	}

	return nil
}
