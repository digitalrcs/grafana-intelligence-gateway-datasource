package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	DefaultTimeoutSeconds     = 300
	DefaultMaxOutputTokens    = 256000
	MaxTimeoutSeconds         = 600
	MaxConfiguredOutputTokens = 1048576
	GrafanaLLMBasePath        = "/api/plugins/grafana-llm-app/resources/llm/v1"
)

type PluginSettings struct {
	Provider            string                `json:"provider"`
	BaseURL             string                `json:"baseUrl"`
	DefaultModel        string                `json:"defaultModel"`
	TimeoutSeconds      int                   `json:"timeoutSeconds"`
	AllowedModels       []string              `json:"allowedModels"`
	MaxOutputTokens     int                   `json:"maxOutputTokens"`
	AllowInsecureHTTP   bool                  `json:"allowInsecureHttp"`
	EnableVision        bool                  `json:"enableVision"`
	AllowedVisionModels []string              `json:"allowedVisionModels"`
	Secrets             *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	APIKey           string `json:"apiKey"`
	BearerToken      string `json:"bearerToken"`
	DirectLineSecret string `json:"directLineSecret"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal data source settings: %w", err)
	}
	settings.Secrets = &SecretPluginSettings{
		APIKey:           source.DecryptedSecureJSONData["apiKey"],
		BearerToken:      source.DecryptedSecureJSONData["bearerToken"],
		DirectLineSecret: source.DecryptedSecureJSONData["directLineSecret"],
	}
	settings.applyDefaults()
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *PluginSettings) applyDefaults() {
	s.Provider = strings.ToLower(strings.TrimSpace(s.Provider))
	if s.Provider == "" {
		s.Provider = "openai"
	}
	s.BaseURL = strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	s.DefaultModel = strings.TrimSpace(s.DefaultModel)
	if s.TimeoutSeconds == 0 {
		s.TimeoutSeconds = DefaultTimeoutSeconds
	}
	if s.MaxOutputTokens == 0 {
		s.MaxOutputTokens = DefaultMaxOutputTokens
	}
	s.AllowedModels = normalizeModels(s.AllowedModels)
	s.AllowedVisionModels = normalizeModels(s.AllowedVisionModels)
}

func normalizeModels(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	models := make([]string, 0, len(input))
	for _, model := range input {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; !ok {
			seen[model] = struct{}{}
			models = append(models, model)
		}
	}
	return models
}

func (s PluginSettings) Validate() error {
	switch s.Provider {
	case "openai", "lmstudio", "openwebui", "grafana-llm", "copilot-studio", "custom":
	default:
		return fmt.Errorf("provider must be openai, lmstudio, openwebui, grafana-llm, copilot-studio, or custom")
	}
	parsed, err := url.Parse(s.BaseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("baseUrl must be an absolute URL without credentials, query, or fragment")
	}
	if parsed.Scheme != "https" && (parsed.Scheme != "http" || !s.AllowInsecureHTTP) {
		return fmt.Errorf("baseUrl must use HTTPS unless allowInsecureHttp is enabled")
	}
	if s.DefaultModel == "" {
		return fmt.Errorf("defaultModel is required")
	}
	if s.Provider == "grafana-llm" {
		if !strings.HasSuffix(parsed.Path, GrafanaLLMBasePath) || parsed.RawPath != "" || path.Clean(parsed.Path) != parsed.Path {
			return fmt.Errorf("the Grafana LLM baseUrl must end with %s (a Grafana subpath may precede it)", GrafanaLLMBasePath)
		}
		if s.Secrets == nil || strings.TrimSpace(s.Secrets.BearerToken) == "" {
			return fmt.Errorf("the Grafana LLM requires a Grafana service-account token in secureJsonData.bearerToken")
		}
		for _, model := range append([]string{s.DefaultModel}, s.AllowedModels...) {
			if model != "base" && model != "large" {
				return fmt.Errorf("the Grafana LLM models must use the base or large aliases configured in grafana-llm-app")
			}
		}
	}
	if s.Provider == "copilot-studio" {
		allowedHost := parsed.Hostname() == "directline.botframework.com" || parsed.Hostname() == "europe.directline.botframework.com" || parsed.Hostname() == "india.directline.botframework.com"
		if parsed.Scheme != "https" || !allowedHost || (parsed.Port() != "" && parsed.Port() != "443") || parsed.Path != "/v3/directline" || parsed.RawPath != "" {
			return fmt.Errorf("Copilot Studio requires HTTPS on a supported Direct Line host at /v3/directline")
		}
		if s.Secrets == nil || strings.TrimSpace(s.Secrets.DirectLineSecret) == "" {
			return fmt.Errorf("Copilot Studio requires an approved channel secret in secureJsonData.directLineSecret")
		}
	}
	if s.TimeoutSeconds < 1 || s.TimeoutSeconds > MaxTimeoutSeconds {
		return fmt.Errorf("timeoutSeconds must be between 1 and %d", MaxTimeoutSeconds)
	}
	if s.MaxOutputTokens < 1 || s.MaxOutputTokens > MaxConfiguredOutputTokens {
		return fmt.Errorf("maxOutputTokens must be between 1 and %d", MaxConfiguredOutputTokens)
	}
	if len(s.AllowedModels) > 0 && !s.ModelAllowed(s.DefaultModel) {
		return fmt.Errorf("defaultModel must be included in allowedModels")
	}
	if s.EnableVision && len(s.AllowedVisionModels) == 0 {
		return fmt.Errorf("allowedVisionModels must explicitly list at least one image-capable model when enableVision is enabled")
	}
	for _, model := range s.AllowedVisionModels {
		if !s.ModelAllowed(model) {
			return fmt.Errorf("allowedVisionModels must be a subset of allowedModels (or the default model when allowedModels is empty)")
		}
	}
	return nil
}

func (s PluginSettings) ModelAllowed(model string) bool {
	if len(s.AllowedModels) == 0 {
		return model == s.DefaultModel
	}
	for _, allowed := range s.AllowedModels {
		if model == allowed {
			return true
		}
	}
	return false
}

func (s PluginSettings) AuthorizationToken() string {
	if s.Secrets == nil {
		return ""
	}
	if s.Provider == "grafana-llm" {
		return strings.TrimSpace(s.Secrets.BearerToken)
	}
	if s.Provider == "copilot-studio" {
		return strings.TrimSpace(s.Secrets.DirectLineSecret)
	}
	if strings.TrimSpace(s.Secrets.BearerToken) != "" {
		return strings.TrimSpace(s.Secrets.BearerToken)
	}
	return strings.TrimSpace(s.Secrets.APIKey)
}

func (s PluginSettings) VisionModelAllowed(model string) bool {
	if !s.EnableVision || !s.ModelAllowed(model) {
		return false
	}
	for _, allowed := range s.AllowedVisionModels {
		if allowed == model {
			return true
		}
	}
	return false
}
