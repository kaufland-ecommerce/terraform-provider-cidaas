package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// injectConfigDataId parses configData (the wizard-shaped commProvider/commMethod/schemaData
// JSON a caller supplies) and ensures it carries an "id" matching serviceSetupId, since
// notifications-srv upserts a provider config keyed by that field. A caller-supplied "id"
// is left untouched.
func injectConfigDataId(configData string, serviceSetupId string) (string, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(configData), &payload); err != nil {
		return "", fmt.Errorf("config_data is not valid JSON: %w", err)
	}

	if _, ok := payload["id"]; !ok {
		payload["id"] = serviceSetupId
	}

	rb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(rb), nil
}

// UpsertNotificationProviderConfig creates or updates the provider credentials for a
// service setup via notifications-srv/providerconfigs. Both create and update are POST -
// the API upserts keyed by "id" = serviceSetupId.
func (c *client) UpsertNotificationProviderConfig(serviceSetupId string, configData string) error {
	body, err := injectConfigDataId(configData, serviceSetupId)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/notifications-srv/providerconfigs", c.HostUrl),
		strings.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")

	_, err = c.doRequest(req)
	return err
}

// GetNotificationProviderConfig only confirms the provider config still exists - it
// deliberately doesn't surface the response body, since notifications-srv may redact
// credential fields on read and we don't want that fighting the value in state/plan.
func (c *client) GetNotificationProviderConfig(serviceSetupId string) error {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/notifications-srv/providerconfigs/%s", c.HostUrl, serviceSetupId),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")

	_, err = c.doRequest(req)
	return err
}
