package plugin

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DigitalRCS/grafana-intelligence-gateway-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func testGateway(t *testing.T, handler http.HandlerFunc) *gateway {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	parsed, err := url.Parse(server.URL + "/v1")
	if err != nil {
		t.Fatal(err)
	}
	return &gateway{
		settings: &models.PluginSettings{
			Provider: "custom", BaseURL: parsed.String(), DefaultModel: "default-model",
			AllowedModels: []string{"default-model", "allowed-model"}, MaxOutputTokens: 100,
			Secrets: &models.SecretPluginSettings{BearerToken: "secret-token"},
		},
		client: server.Client(), baseURL: parsed, concurrent: make(chan struct{}, 4), rate: &minuteLimiter{},
	}
}

func TestQueryDataReturnsAnswerFrame(t *testing.T) {
	gateway := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer secret-token" {
			t.Error("missing backend authorization")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]string{"content": "assessment"}}},
		})
	})
	ds := Datasource{gateway: gateway}
	queryJSON := []byte(`{"userPrompt":"assess this","maxOutputTokens":50}`)
	response, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: queryJSON}},
	})
	if err != nil {
		t.Fatal(err)
	}
	result := response.Responses["A"]
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if len(result.Frames) != 1 || result.Frames[0].Fields[0].Len() != 1 {
		t.Fatal("expected one answer")
	}
}

func TestGatewayEnforcesTokenCeilingAndModel(t *testing.T) {
	var received providerRequest
	gateway := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	})
	requested := 500
	response, err := gateway.requestChat(context.Background(), gatewayRequest{
		Prompt: &prompt{User: "hello"}, Model: "allowed-model", MaxOutputTokens: &requested,
	})
	if err != nil {
		t.Fatal(err)
	}
	closeBody(response.Body)
	if received.MaxTokens != 100 {
		t.Fatalf("expected administrator ceiling 100, got %d", received.MaxTokens)
	}

	_, err = gateway.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "hello"}, Model: "blocked"})
	if statusFor(err) != http.StatusBadRequest {
		t.Fatalf("expected blocked model, got %v", err)
	}
}

func TestProviderErrorsAreRedacted(t *testing.T) {
	gateway := testGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"secret provider detail"}}`))
	})
	_, err := gateway.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "hello"}})
	if err == nil || strings.Contains(err.Error(), "secret provider detail") {
		t.Fatalf("provider error body leaked: %v", err)
	}
}

func TestLMStudioAllowsPrivateHTTPAddress(t *testing.T) {
	target, err := url.Parse("http://192.168.1.25:1234/v1")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateHost(context.Background(), target, "lmstudio", true, net.DefaultResolver); err != nil {
		t.Fatalf("expected private LM Studio address to be allowed: %v", err)
	}
}

func TestHTTPOverrideAllowsAddressWithoutLocalClassification(t *testing.T) {
	target, err := url.Parse("http://8.8.8.8:1234/v1")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateHost(context.Background(), target, "custom", true, net.DefaultResolver); err != nil {
		t.Fatalf("expected explicit HTTP override to bypass address classification: %v", err)
	}
}

func TestHTTPRequiresExplicitOverride(t *testing.T) {
	target, err := url.Parse("http://192.168.1.25:1234/v1")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateHost(context.Background(), target, "lmstudio", false, net.DefaultResolver); err == nil {
		t.Fatal("expected HTTP without the override to be rejected")
	}
}

func TestMinuteLimiter(t *testing.T) {
	limiter := &minuteLimiter{}
	now := time.Now()
	for i := 0; i < requestsPerMinute; i++ {
		if !limiter.allow(now) {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if limiter.allow(now) {
		t.Fatal("request above rate limit should be rejected")
	}
	if !limiter.allow(now.Add(time.Minute)) {
		t.Fatal("new minute should reset limit")
	}
}

func TestExtractAnswer(t *testing.T) {
	answer, err := extractAnswer([]byte(`{"choices":[{"message":{"content":"  result  "}}]}`))
	if err != nil || answer != "result" {
		t.Fatalf("unexpected result %q: %v", answer, err)
	}
	if _, err := extractAnswer([]byte(`{"choices":[]}`)); err == nil {
		t.Fatal("expected empty response error")
	}
}
