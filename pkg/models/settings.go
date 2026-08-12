package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	DefaultTimeoutSeconds     = 300
	DefaultMaxOutputTokens    = 256000
	MaxTimeoutSeconds         = 600
	MaxConfiguredOutputTokens = 1048576
)

type PluginSettings struct {
	Provider        string                `json:"provider"`
	BaseURL         string                `json:"baseUrl"`
	DefaultModel    string                `json:"defaultModel"`
	TimeoutSeconds  int                   `json:"timeoutSeconds"`
	AllowedModels   []string              `json:"allowedModels"`
	MaxOutputTokens int                   `json:"maxOutputTokens"`
	AllowStreaming  bool                  `json:"allowStreaming"`
	Secrets         *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	APIKey       string `json:"apiKey"`
	BearerToken  string `json:"bearerToken"`
	ClientSecret string `json:"clientSecret"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal data source settings: %w", err)
	}
	settings.Secrets = &SecretPluginSettings{
		APIKey:       source.DecryptedSecureJSONData["apiKey"],
		BearerToken:  source.DecryptedSecureJSONData["bearerToken"],
		ClientSecret: source.DecryptedSecureJSONData["clientSecret"],
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
	seen := make(map[string]struct{}, len(s.AllowedModels))
	models := make([]string, 0, len(s.AllowedModels))
	for _, model := range s.AllowedModels {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; !ok {
			seen[model] = struct{}{}
			models = append(models, model)
		}
	}
	s.AllowedModels = models
}

func (s PluginSettings) Validate() error {
	switch s.Provider {
	case "openai", "lmstudio", "custom":
	default:
		return fmt.Errorf("provider must be openai, lmstudio, or custom")
	}
	parsed, err := url.Parse(s.BaseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("baseUrl must be an absolute URL without credentials, query, or fragment")
	}
	if parsed.Scheme != "https" && (s.Provider != "lmstudio" || parsed.Scheme != "http") {
		return fmt.Errorf("baseUrl must use HTTPS; HTTP is supported only for approved LM Studio local hosts")
	}
	if s.DefaultModel == "" {
		return fmt.Errorf("defaultModel is required")
	}
	if s.TimeoutSeconds < 1 || s.TimeoutSeconds > MaxTimeoutSeconds {
		return fmt.Errorf("timeoutSeconds must be between 1 and %d", MaxTimeoutSeconds)
	}
	if s.MaxOutputTokens < 1 || s.MaxOutputTokens > MaxConfiguredOutputTokens {
		return fmt.Errorf("maxOutputTokens must be between 1 and %d", MaxConfiguredOutputTokens)
	}
	return nil
}

func (s PluginSettings) ModelAllowed(model string) bool {
	if len(s.AllowedModels) == 0 {
		return true
	}
	for _, allowed := range s.AllowedModels {
		if model == allowed {
			return true
		}
	}
	return false
}

func (s PluginSettings) AuthorizationToken() string {
	if strings.TrimSpace(s.Secrets.BearerToken) != "" {
		return strings.TrimSpace(s.Secrets.BearerToken)
	}
	return strings.TrimSpace(s.Secrets.APIKey)
}
