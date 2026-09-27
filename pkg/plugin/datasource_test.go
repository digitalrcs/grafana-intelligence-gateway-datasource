package plugin

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync/atomic"
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

func TestGatewayEmptyAllowListPermitsOnlyDefaultRequests(t *testing.T) {
	received := make(chan providerRequest, 2)
	gateway := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		var request providerRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		received <- request
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	})
	gateway.settings.AllowedModels = nil
	for _, model := range []string{"", "default-model"} {
		response, err := gateway.requestChat(context.Background(), gatewayRequest{
			Prompt: &prompt{User: "hello"}, Model: model,
		})
		if err != nil {
			t.Fatal(err)
		}
		closeBody(response.Body)
		request := <-received
		if request.Model != "default-model" || request.Stream {
			t.Fatalf("unexpected provider request: %#v", request)
		}
	}
	_, err := gateway.requestChat(context.Background(), gatewayRequest{
		Prompt: &prompt{User: "hello"}, Model: "allowed-model",
	})
	if statusFor(err) != http.StatusBadRequest {
		t.Fatalf("expected non-default model to be rejected, got %v", err)
	}
	if len(received) != 0 {
		t.Fatal("a disallowed model request reached the provider")
	}
}

func TestGatewayRejectsStreamingDespiteLegacyConfiguration(t *testing.T) {
	var calls atomic.Int32
	gateway := testGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{}`))
	})
	settings, err := models.LoadPluginSettings(backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"provider":"custom","baseUrl":"https://example.com/v1","defaultModel":"default-model","allowStreaming":true}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	gateway.settings = settings
	ds := Datasource{gateway: gateway}
	response := httptest.NewRecorder()
	ds.handleChat(response, httptest.NewRequest(http.MethodPost, "/chat/completions", strings.NewReader(
		`{"prompt":{"user":"hello"},"stream":true}`,
	)))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "streaming is not supported") {
		t.Fatalf("expected unsupported streaming error, got %d %s", response.Code, response.Body.String())
	}
	if calls.Load() != 0 {
		t.Fatal("a streaming request reached the provider")
	}
}

func TestHandleChatRejectsUnexpectedProviderStreaming(t *testing.T) {
	gateway := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Accept") != "application/json" {
			t.Error("buffered requests must ask for JSON")
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		_, _ = w.Write([]byte("data: private-stream-metadata\n\n"))
	})
	ds := Datasource{gateway: gateway}
	response := httptest.NewRecorder()
	ds.handleChat(response, httptest.NewRequest(http.MethodPost, "/chat/completions", strings.NewReader(`{"prompt":{"user":"hello"}}`)))
	if response.Code != http.StatusBadGateway || strings.Contains(response.Body.String(), "private-stream-metadata") {
		t.Fatalf("unexpected streaming response leaked: %d %s", response.Code, response.Body.String())
	}
}

func TestFilterModelList(t *testing.T) {
	settings := &models.PluginSettings{
		DefaultModel: "default-model", AllowedModels: []string{"default-model", "allowed-model"},
	}
	tests := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "filter deduplicate and discard private metadata",
			body: `{"owner":"private-owner","data":[{"id":"blocked-model","path":"private-path"},{"id":"allowed-model","owner":"private-owner","object":"private-object"},{"id":"allowed-model"},{"id":" default-model "}]}`,
			want: []string{"allowed-model", "default-model"},
		},
		{
			name: "alternate models array and mixed entries",
			body: `{"models":["allowed-model",{"name":"default-model","path":"private-path"},{"id":"blocked-model","name":"default-model"},null,123,true,{},[],{"id":123},{"name":true},""]}`,
			want: []string{"allowed-model", "default-model"},
		},
		{name: "empty data is authoritative", body: `{"data":[],"models":["default-model"]}`, want: []string{}},
		{name: "null data falls back to models", body: `{"data":null,"models":["default-model"]}`, want: []string{"default-model"}},
		{name: "unapproved only", body: `{"data":["blocked-model"]}`, want: []string{}},
		{name: "empty alternate list", body: `{"models":[]}`, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := filterModelList([]byte(tt.body), settings)
			if err != nil {
				t.Fatal(err)
			}
			var payload struct {
				Object string              `json:"object"`
				Data   []map[string]string `json:"data"`
			}
			if err := json.Unmarshal(filtered, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Object != "list" || payload.Data == nil || strings.Contains(string(filtered), "private-") {
				t.Fatalf("unexpected model-list envelope or leaked metadata: %s", filtered)
			}
			got := make([]string, 0, len(payload.Data))
			for _, model := range payload.Data {
				if len(model) != 2 || model["object"] != "model" {
					t.Fatalf("unexpected model metadata: %#v", model)
				}
				got = append(got, model["id"])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got model IDs %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestFilterModelListRejectsMalformedPayloads(t *testing.T) {
	for _, body := range []string{
		``, `{`, `null`, `[]`, `"private-provider-error"`, `{}`, `{"error":"private-provider-error"}`,
		`{"data":null}`, `{"models":null}`, `{"data":{}}`, `{"data":123}`, `{"data":"private-provider-error"}`,
		`{"models":{}}`, `{"data":[],"models":[] trailing}`, `{"data":[]} {"data":[]}`,
	} {
		t.Run(body, func(t *testing.T) {
			_, err := filterModelList([]byte(body), &models.PluginSettings{DefaultModel: "default-model"})
			if err == nil || err.Error() != "provider returned an invalid model list" {
				t.Fatalf("expected sanitized model-list error, got %v", err)
			}
		})
	}
}

func TestHandleModelsFiltersWithDefaultOnlyPolicy(t *testing.T) {
	gateway := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v1/models" || req.Header.Get("Authorization") != "Bearer secret-token" {
			t.Error("unexpected models provider request")
		}
		w.Header().Set("X-Private-Provider", "private-header")
		_, _ = w.Write([]byte(`{"data":[{"id":"allowed-model"},{"id":"default-model","private":"private-metadata"}]}`))
	})
	gateway.settings.AllowedModels = nil
	ds := Datasource{gateway: gateway}
	response := httptest.NewRecorder()
	ds.handleModels(response, httptest.NewRequest(http.MethodGet, "/models", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected models response: %d %#v", response.Code, response.Header())
	}
	if response.Body.String() != `{"data":[{"id":"default-model","object":"model"}],"object":"list"}` {
		t.Fatalf("unfiltered models response: %s", response.Body.String())
	}
	if response.Header().Get("X-Private-Provider") != "" {
		t.Fatal("provider metadata header leaked")
	}
}

func TestHandleModelsRejectsMalformedProviderResponse(t *testing.T) {
	gateway := testGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":"private-provider-error"}`))
	})
	ds := Datasource{gateway: gateway}
	response := httptest.NewRecorder()
	ds.handleModels(response, httptest.NewRequest(http.MethodGet, "/models", nil))
	if response.Code != http.StatusBadGateway || strings.Contains(response.Body.String(), "private-provider-error") {
		t.Fatalf("unexpected malformed-model-list response: %d %s", response.Code, response.Body.String())
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
