package plugin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failingProviderBody struct{ err error }

func (r failingProviderBody) Read([]byte) (int, error) { return 0, r.err }

func TestResponseReadErrorsAreSanitized(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
	}{
		{errors.New("private-provider-URL-and-token"), http.StatusBadGateway},
		{context.DeadlineExceeded, http.StatusGatewayTimeout},
		{context.Canceled, http.StatusRequestTimeout},
	} {
		response := httptest.NewRecorder()
		copyResponse(response, &http.Response{Header: make(http.Header), Body: io.NopCloser(failingProviderBody{err: test.err})})
		assertPrivateJSONResponse(t, response)
		if response.Code != test.status || strings.Contains(response.Body.String(), "private-provider") {
			t.Fatalf("unsafe body read error: %d %s", response.Code, response.Body.String())
		}
	}
}

func assertPrivateJSONResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("missing private JSON response headers: %#v", response.Header())
	}
}

func TestCopyResponseAllowsOnlyBoundedJSONObjects(t *testing.T) {
	for _, test := range []struct {
		name, contentType, body string
		wantStatus              int
	}{
		{"JSON", "application/json", `{"choices":[]}`, http.StatusOK},
		{"JSON without type", "", ` {"choices":[]} `, http.StatusOK},
		{"mislabeled JSON is safely normalized", "text/html", `{"text":"<script>example</script>"}`, http.StatusOK},
		{"HTML", "text/html", `<script>private-provider-content</script>`, http.StatusBadGateway},
		{"non JSON mislabeled as JSON", "application/json", `private-provider-content`, http.StatusBadGateway},
		{"truncated JSON", "application/json", `{"private-provider-content":`, http.StatusBadGateway},
		{"multiple objects", "application/json", `{} {"private-provider-content":true}`, http.StatusBadGateway},
		{"array", "application/json", `["private-provider-content"]`, http.StatusBadGateway},
		{"null", "application/json", `null`, http.StatusBadGateway},
		{"JSON string", "application/json", `"private-provider-content"`, http.StatusBadGateway},
		{"empty", "application/json", ``, http.StatusBadGateway},
		{"stream", "text/event-stream", "data: private-provider-content", http.StatusBadGateway},
	} {
		t.Run(test.name, func(t *testing.T) {
			upstream := &http.Response{Header: http.Header{"Content-Type": []string{test.contentType}, "Cache-Control": []string{"public"}, "X-Provider-Secret": []string{"secret"}}, Body: io.NopCloser(strings.NewReader(test.body))}
			response := httptest.NewRecorder()
			copyResponse(response, upstream)
			assertPrivateJSONResponse(t, response)
			if response.Code != test.wantStatus || response.Header().Get("X-Provider-Secret") != "" {
				t.Fatalf("unsafe response: %d %#v", response.Code, response.Header())
			}
			if test.wantStatus == http.StatusOK && response.Body.String() != test.body {
				t.Fatal("valid compatible JSON response changed")
			}
			if test.wantStatus != http.StatusOK && strings.Contains(response.Body.String(), "private-provider-content") {
				t.Fatal("private upstream response leaked")
			}
		})
	}
}

func TestEveryResourceResponseIsNonCacheable(t *testing.T) {
	g := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/models") {
			_, _ = io.WriteString(w, `{"data":[{"id":"default-model"}]}`)
			return
		}
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"assessment"}}]}`)
	})
	ds := &Datasource{gateway: g}
	for _, test := range []struct {
		name, method, path, body string
		handler                  http.HandlerFunc
		status                   int
	}{
		{"capabilities", http.MethodGet, "/capabilities", "", ds.handleCapabilities, http.StatusOK},
		{"models", http.MethodGet, "/models", "", ds.handleModels, http.StatusOK},
		{"chat", http.MethodPost, "/chat/completions", `{"prompt":{"user":"hello"}}`, ds.handleChat, http.StatusOK},
		{"invalid body", http.MethodPost, "/chat/completions", `{"unknown":"private"}`, ds.handleChat, http.StatusBadRequest},
		{"invalid model", http.MethodPost, "/chat/completions", `{"prompt":{"user":"hello"},"model":"blocked"}`, ds.handleChat, http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			test.handler(response, httptest.NewRequest(test.method, test.path, strings.NewReader(test.body)))
			assertPrivateJSONResponse(t, response)
			if response.Code != test.status {
				t.Fatalf("unexpected status: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
