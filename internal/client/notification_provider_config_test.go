package client

import (
	"encoding/json"
	"testing"
)

func TestInjectConfigDataId_AddsIdWhenAbsent(t *testing.T) {
	out, err := injectConfigDataId(`{"commProvider":"custom-ses-email","commMethod":"email","schemaData":{"region":"eu-central-1"}}`, "service-setup-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if payload["id"] != "service-setup-id" {
		t.Errorf("id = %v, want %q", payload["id"], "service-setup-id")
	}
	if payload["commProvider"] != "custom-ses-email" {
		t.Errorf("commProvider = %v, want %q", payload["commProvider"], "custom-ses-email")
	}
}

func TestInjectConfigDataId_LeavesExistingIdUntouched(t *testing.T) {
	out, err := injectConfigDataId(`{"id":"caller-supplied-id","commProvider":"custom-ses-email"}`, "service-setup-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if payload["id"] != "caller-supplied-id" {
		t.Errorf("id = %v, want the caller-supplied id to be left untouched", payload["id"])
	}
}

func TestInjectConfigDataId_RejectsInvalidJson(t *testing.T) {
	if _, err := injectConfigDataId("not json", "service-setup-id"); err == nil {
		t.Fatal("expected an error for invalid JSON input, got nil")
	}
}
