package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/DigitalRCS/grafana-intelligence-gateway-datasource/pkg/models"
)

const (
	maxRequestBytes   = 1 << 20
	maxResponseBytes  = 16 << 20
	requestsPerMinute = 30
)

type prompt struct {
	System string `json:"system"`
	User   string `json:"user"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type gatewayRequest struct {
	Prompt          *prompt       `json:"prompt,omitempty"`
	Messages        []chatMessage `json:"messages,omitempty"`
	Model           string        `json:"model,omitempty"`
	Temperature     *float64      `json:"temperature,omitempty"`
	MaxOutputTokens *int          `json:"maxOutputTokens,omitempty"`
	MaxTokens       *int          `json:"max_tokens,omitempty"`
	Stream          bool          `json:"stream,omitempty"`
}

type providerRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream"`
}

type gateway struct {
	settings   *models.PluginSettings
	client     *http.Client
	baseURL    *url.URL
	concurrent chan struct{}
	rate       *minuteLimiter
}

type minuteLimiter struct {
	mu      sync.Mutex
	started time.Time
	count   int
}

func (l *minuteLimiter) allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.started.IsZero() || now.Sub(l.started) >= time.Minute {
		l.started, l.count = now, 0
	}
	if l.count >= requestsPerMinute {
		return false
	}
	l.count++
	return true
}

func newGateway(ctx context.Context, settings *models.PluginSettings) (*gateway, error) {
	baseURL, err := url.Parse(settings.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if err := validateHost(ctx, baseURL, settings.Provider, settings.AllowInsecureHTTP, net.DefaultResolver); err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(settings.TimeoutSeconds) * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("provider redirects are disabled")
		},
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: time.Duration(settings.TimeoutSeconds) * time.Second,
			MaxIdleConns:          20,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
		},
	}
	return &gateway{
		settings:   settings,
		client:     client,
		baseURL:    baseURL,
		concurrent: make(chan struct{}, 4),
		rate:       &minuteLimiter{},
	}, nil
}

func validateHost(ctx context.Context, target *url.URL, provider string, allowInsecureHTTP bool, resolver *net.Resolver) error {
	host := strings.ToLower(target.Hostname())
	if host == "" {
		return errors.New("baseUrl host is required")
	}
	localName := host == "localhost" || host == "host.docker.internal"
	if target.Scheme == "http" && !allowInsecureHTTP {
		return errors.New("HTTP requires the Allow insecure HTTP data-source setting")
	}
	if provider == "openai" && host != "api.openai.com" {
		return errors.New("OpenAI provider baseUrl must use api.openai.com; use custom for other compatible hosts")
	}
	if localName && provider == "lmstudio" {
		return nil
	}
	addresses, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve provider host: %w", err)
	}
	if len(addresses) == 0 {
		return errors.New("provider host resolved to no addresses")
	}
	for _, address := range addresses {
		ip := address.IP
		if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return errors.New("provider host resolves to a prohibited network address")
		}
		if provider == "openai" && (ip.IsLoopback() || ip.IsPrivate()) {
			return errors.New("remote provider host must not resolve to a loopback or private address")
		}
	}
	return nil
}

func (g *gateway) close() {
	if transport, ok := g.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}

func (g *gateway) acquire(ctx context.Context) error {
	if !g.rate.allow(time.Now()) {
		return &gatewayError{status: http.StatusTooManyRequests, message: "data source request rate limit exceeded"}
	}
	select {
	case g.concurrent <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *gateway) release() { <-g.concurrent }

func (g *gateway) requestChat(ctx context.Context, request gatewayRequest) (*http.Response, error) {
	if err := g.acquire(ctx); err != nil {
		return nil, err
	}
	defer g.release()

	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = g.settings.DefaultModel
	}
	if !g.settings.ModelAllowed(model) {
		return nil, &gatewayError{status: http.StatusBadRequest, message: "requested model is not allowed by the administrator"}
	}
	if request.Stream && !g.settings.AllowStreaming {
		return nil, &gatewayError{status: http.StatusBadRequest, message: "streaming is disabled for this data source"}
	}
	if request.Temperature != nil && (*request.Temperature < 0 || *request.Temperature > 2) {
		return nil, &gatewayError{status: http.StatusBadRequest, message: "temperature must be between 0 and 2"}
	}
	messages := request.Messages
	if len(messages) == 0 && request.Prompt != nil {
		if strings.TrimSpace(request.Prompt.System) != "" {
			messages = append(messages, chatMessage{Role: "system", Content: request.Prompt.System})
		}
		messages = append(messages, chatMessage{Role: "user", Content: request.Prompt.User})
	}
	if len(messages) == 0 {
		return nil, &gatewayError{status: http.StatusBadRequest, message: "prompt or messages is required"}
	}
	for _, message := range messages {
		if message.Role != "system" && message.Role != "user" && message.Role != "assistant" {
			return nil, &gatewayError{status: http.StatusBadRequest, message: "message role is not allowed"}
		}
		if strings.TrimSpace(message.Content) == "" {
			return nil, &gatewayError{status: http.StatusBadRequest, message: "message content must not be empty"}
		}
	}
	cap := g.settings.MaxOutputTokens
	requested := request.MaxOutputTokens
	if requested == nil {
		requested = request.MaxTokens
	}
	if requested != nil {
		if *requested < 1 {
			return nil, &gatewayError{status: http.StatusBadRequest, message: "output token cap must be positive"}
		}
		if *requested < cap {
			cap = *requested
		}
	}
	body, err := json.Marshal(providerRequest{
		Model: model, Messages: messages, Temperature: request.Temperature, MaxTokens: cap, Stream: request.Stream,
	})
	if err != nil {
		return nil, err
	}
	if len(body) > maxRequestBytes {
		return nil, &gatewayError{status: http.StatusRequestEntityTooLarge, message: "request exceeds the 1 MiB limit"}
	}
	endpoint := *g.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/chat/completions"
	providerReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	providerReq.Header.Set("Content-Type", "application/json")
	providerReq.Header.Set("Accept", "application/json")
	if request.Stream {
		providerReq.Header.Set("Accept", "text/event-stream")
	}
	if token := g.settings.AuthorizationToken(); token != "" {
		providerReq.Header.Set("Authorization", "Bearer "+token)
	} else if g.settings.Provider == "openai" {
		return nil, &gatewayError{status: http.StatusBadRequest, message: "an API key or bearer token is required for OpenAI"}
	}
	response, err := g.client.Do(providerReq)
	if err != nil {
		return nil, &gatewayError{status: http.StatusBadGateway, message: "provider request failed"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		closeBody(response.Body)
		message := fmt.Sprintf("provider returned HTTP %d", response.StatusCode)
		switch response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			message = "provider authentication failed"
		case http.StatusTooManyRequests:
			message = "provider rate limit reached"
		}
		return nil, &gatewayError{status: response.StatusCode, message: message}
	}
	return response, nil
}

func (g *gateway) requestModels(ctx context.Context) (*http.Response, error) {
	if err := g.acquire(ctx); err != nil {
		return nil, err
	}
	defer g.release()
	endpoint := *g.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if token := g.settings.AuthorizationToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := g.client.Do(req)
	if err != nil {
		return nil, &gatewayError{status: http.StatusBadGateway, message: "provider connection failed"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		closeBody(response.Body)
		return nil, &gatewayError{status: response.StatusCode, message: fmt.Sprintf("provider returned HTTP %d", response.StatusCode)}
	}
	return response, nil
}

func readBounded(body io.Reader) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(body, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxResponseBytes {
		return nil, &gatewayError{status: http.StatusBadGateway, message: "provider response exceeded the 16 MiB limit"}
	}
	return content, nil
}

type gatewayError struct {
	status  int
	message string
}

func (e *gatewayError) Error() string { return e.message }

func statusFor(err error) int {
	var typed *gatewayError
	if errors.As(err, &typed) {
		return typed.status
	}
	return http.StatusInternalServerError
}
