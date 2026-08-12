package models

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func settingsSource(t *testing.T, settings PluginSettings) backend.DataSourceInstanceSettings {
	t.Helper()
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	return backend.DataSourceInstanceSettings{JSONData: data, DecryptedSecureJSONData: map[string]string{"apiKey": "key"}}
}

func TestLoadPluginSettings(t *testing.T) {
	settings, err := LoadPluginSettings(settingsSource(t, PluginSettings{
		Provider: "openai", BaseURL: "https://api.openai.com/v1/", DefaultModel: "model",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if settings.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected URL %q", settings.BaseURL)
	}
	if settings.TimeoutSeconds != DefaultTimeoutSeconds || settings.MaxOutputTokens != DefaultMaxOutputTokens {
		t.Fatal("defaults were not applied")
	}
	if settings.AuthorizationToken() != "key" {
		t.Fatal("API key was not loaded securely")
	}
}

func TestSettingsRejectInsecureRemoteURL(t *testing.T) {
	_, err := LoadPluginSettings(settingsSource(t, PluginSettings{
		Provider: "custom", BaseURL: "http://example.com/v1", DefaultModel: "model",
	}))
	if err == nil {
		t.Fatal("expected insecure remote URL to be rejected")
	}
}

func TestModelAllowList(t *testing.T) {
	settings := PluginSettings{AllowedModels: []string{"allowed"}}
	if !settings.ModelAllowed("allowed") || settings.ModelAllowed("blocked") {
		t.Fatal("allow-list was not enforced")
	}
}
