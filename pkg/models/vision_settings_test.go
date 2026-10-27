package models

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestVisionPolicyIsExplicitAndSubsetOfTextPolicy(t *testing.T) {
	base := PluginSettings{Provider: "custom", BaseURL: "https://models.example/v1", DefaultModel: "default", AllowedModels: []string{"default", "vision"}}
	for _, test := range []struct {
		name      string
		enabled   bool
		models    []string
		wantError string
	}{
		{name: "disabled by default"},
		{name: "enabled needs explicit model", enabled: true, wantError: "allowedVisionModels"},
		{name: "enabled subset", enabled: true, models: []string{" vision ", "vision", " "}},
		{name: "nonallowed image model", enabled: true, models: []string{"unknown"}, wantError: "subset"},
		{name: "invalid dormant policy", models: []string{"unknown"}, wantError: "subset"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := base
			config.EnableVision, config.AllowedVisionModels = test.enabled, test.models
			loaded, err := LoadPluginSettings(settingsSource(t, config))
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("expected %q, got %v", test.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if loaded.VisionModelAllowed("vision") != test.enabled || loaded.VisionModelAllowed("default") || loaded.VisionModelAllowed("unknown") {
				t.Fatal("vision policy did not preserve explicit approval")
			}
			if test.enabled && len(loaded.AllowedVisionModels) != 1 {
				t.Fatal("vision models were not normalized")
			}
		})
	}
	base.AllowedModels = nil
	base.EnableVision, base.AllowedVisionModels = true, []string{"default"}
	loaded, err := LoadPluginSettings(settingsSource(t, base))
	if err != nil || !loaded.VisionModelAllowed("default") {
		t.Fatalf("default-only vision policy failed: %v", err)
	}
}

func TestGrafanaLLMSettingsRequireScopedResourceAndServiceAccountToken(t *testing.T) {
	for _, test := range []struct {
		name, baseURL, model, token string
		allowed                     []string
		wantError                   string
	}{
		{name: "root", baseURL: "https://grafana.example" + GrafanaLLMBasePath, model: "base", token: "service-account"},
		{name: "subpath", baseURL: "https://grafana.example/monitoring" + GrafanaLLMBasePath + "/", model: "large", token: "service-account", allowed: []string{"base", "large"}},
		{name: "no token", baseURL: "https://grafana.example" + GrafanaLLMBasePath, model: "base", wantError: "service-account token"},
		{name: "wrong resource", baseURL: "https://grafana.example/api/admin", model: "base", token: "service-account", wantError: "baseUrl must end"},
		{name: "endpoint instead of base", baseURL: "https://grafana.example" + GrafanaLLMBasePath + "/models", model: "base", token: "service-account", wantError: "baseUrl must end"},
		{name: "traversal", baseURL: "https://grafana.example/prefix/../" + strings.TrimPrefix(GrafanaLLMBasePath, "/"), model: "base", token: "service-account", wantError: "baseUrl must end"},
		{name: "encoded path", baseURL: "https://grafana.example/%2e%2e" + GrafanaLLMBasePath, model: "base", token: "service-account", wantError: "baseUrl must end"},
		{name: "model id instead of alias", baseURL: "https://grafana.example" + GrafanaLLMBasePath, model: "gpt-4.1", token: "service-account", wantError: "base or large"},
		{name: "allowed model id", baseURL: "https://grafana.example" + GrafanaLLMBasePath, model: "base", token: "service-account", allowed: []string{"base", "gpt-4.1"}, wantError: "base or large"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := PluginSettings{Provider: "grafana-llm", BaseURL: test.baseURL, DefaultModel: test.model, AllowedModels: test.allowed}
			source := settingsSource(t, config)
			source.DecryptedSecureJSONData["bearerToken"] = test.token
			loaded, err := LoadPluginSettings(source)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("expected %q, got %v", test.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if loaded.AuthorizationToken() != test.token {
				t.Fatal("Grafana LLM must use the service-account bearer token")
			}
		})
	}
}

func TestVisionProvisioningCannotExposeSecretsOrInferCapabilities(t *testing.T) {
	source := backend.DataSourceInstanceSettings{
		JSONData:                []byte(`{"provider":"openwebui","baseUrl":"https://webui.example/api","defaultModel":"vision-name","enableVision":false,"allowedVisionModels":["vision-name"],"apiKey":"ignored-browser-key","bearerToken":"ignored-browser-token"}`),
		DecryptedSecureJSONData: map[string]string{"apiKey": "secure-api-key"},
	}
	loaded, err := LoadPluginSettings(source)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.VisionModelAllowed("vision-name") {
		t.Fatal("vision must be explicitly enabled")
	}
	if loaded.AuthorizationToken() != "secure-api-key" {
		t.Fatal("secret did not come exclusively from secure storage")
	}
	encoded, err := json.Marshal(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secure-api-key") || strings.Contains(string(encoded), "ignored-browser") {
		t.Fatal("secrets leaked through JSON settings")
	}
}

func TestCopilotSettingsRequireApprovedDirectLineBoundary(t *testing.T) {
	for _, test := range []struct {
		name, baseURL, secret string
		insecure              bool
		wantError             string
	}{
		{name: "global", baseURL: "https://directline.botframework.com/v3/directline", secret: "channel-secret"},
		{name: "Europe", baseURL: "https://europe.directline.botframework.com/v3/directline/", secret: "channel-secret"},
		{name: "India TLS port", baseURL: "https://india.directline.botframework.com:443/v3/directline", secret: "channel-secret"},
		{name: "missing channel secret", baseURL: "https://directline.botframework.com/v3/directline", wantError: "directLineSecret"},
		{name: "wrong host", baseURL: "https://attacker.example/v3/directline", secret: "channel-secret", wantError: "supported Direct Line"},
		{name: "suffix host", baseURL: "https://directline.botframework.com.attacker.example/v3/directline", secret: "channel-secret", wantError: "supported Direct Line"},
		{name: "wrong route", baseURL: "https://directline.botframework.com/v3/directline/conversations", secret: "channel-secret", wantError: "supported Direct Line"},
		{name: "wrong port", baseURL: "https://directline.botframework.com:8443/v3/directline", secret: "channel-secret", wantError: "supported Direct Line"},
		{name: "HTTP override not valid", baseURL: "http://directline.botframework.com/v3/directline", secret: "channel-secret", insecure: true, wantError: "supported Direct Line"},
		{name: "encoded path", baseURL: "https://directline.botframework.com/v3/%64irectline", secret: "channel-secret", wantError: "supported Direct Line"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := settingsSource(t, PluginSettings{Provider: "copilot-studio", BaseURL: test.baseURL, DefaultModel: "operations-agent", AllowedModels: []string{"operations-agent"}, EnableVision: true, AllowedVisionModels: []string{"operations-agent"}, AllowInsecureHTTP: test.insecure})
			source.DecryptedSecureJSONData["directLineSecret"] = test.secret
			source.DecryptedSecureJSONData["bearerToken"] = "not-a-channel-secret"
			loaded, err := LoadPluginSettings(source)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("wanted %q, got %v", test.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if loaded.AuthorizationToken() != test.secret || !loaded.VisionModelAllowed("operations-agent") {
				t.Fatal("channel credential or explicit alias policy lost")
			}
			encoded, _ := json.Marshal(loaded)
			if strings.Contains(string(encoded), test.secret) {
				t.Fatal("Direct Line secret leaked into settings JSON")
			}
		})
	}
}
