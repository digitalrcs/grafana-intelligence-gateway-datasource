package plugin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DigitalRCS/grafana-intelligence-gateway-datasource/pkg/models"
)

func testImage(t *testing.T, format string, width, height int) imageAttachment {
	t.Helper()
	var body bytes.Buffer
	pixels := image.NewNRGBA(image.Rect(0, 0, width, height))
	var err error
	if format == "jpeg" {
		err = jpeg.Encode(&body, pixels, &jpeg.Options{Quality: 80})
	} else {
		err = png.Encode(&body, pixels)
	}
	if err != nil {
		t.Fatal(err)
	}
	return imageAttachment{Name: "Chart", DataURL: "data:image/" + format + ";base64," + base64.StdEncoding.EncodeToString(body.Bytes())}
}

func enableTestVision(g *gateway) {
	g.settings.EnableVision = true
	g.settings.AllowedVisionModels = []string{"default-model", "allowed-model"}
}

func TestVisionUpstreamHasValidatedMultipartAndServerCredentials(t *testing.T) {
	for _, format := range []string{"png", "jpeg"} {
		t.Run(format, func(t *testing.T) {
			attachment := testImage(t, format, 2, 2)
			var received providerRequest
			g := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/v1/chat/completions" || req.Header.Get("Authorization") != "Bearer secret-token" {
					t.Error("unexpected upstream route or credentials")
				}
				if err := json.NewDecoder(req.Body).Decode(&received); err != nil {
					t.Error(err)
				}
				_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"chart assessment"}}]}`))
			})
			enableTestVision(g)
			response, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{System: "Be precise", User: "Analyze the chart"}, Images: []imageAttachment{attachment}})
			if err != nil {
				t.Fatal(err)
			}
			closeBody(response.Body)
			if received.Model != "default-model" || received.MaxTokens != 100 || received.Stream || len(received.Messages) != 2 {
				t.Fatalf("wrong upstream envelope: %#v", received)
			}
			if string(received.Messages[0].Content) != `"Be precise"` || received.Messages[1].Role != "user" {
				t.Fatal("text message roles changed")
			}
			var parts []providerContentPart
			if err := json.Unmarshal(received.Messages[1].Content, &parts); err != nil {
				t.Fatal(err)
			}
			want := []providerContentPart{{Type: "text", Text: "Analyze the chart"}, {Type: "text", Text: "Image 1: Chart"}, {Type: "image_url", ImageURL: &providerImageURL{URL: attachment.DataURL}}}
			if !reflect.DeepEqual(parts, want) {
				t.Fatal("upstream image parts differ from validated input")
			}
			encoded, _ := json.Marshal(received)
			if strings.Contains(string(encoded), "secret-token") || strings.Contains(string(encoded), "dataUrl") {
				t.Fatal("client-only attachment fields or credentials leaked into provider body")
			}
		})
	}
}

func TestVisionGateAndInvalidRequestsNeverReachProvider(t *testing.T) {
	valid := testImage(t, "png", 2, 2)
	for _, test := range []struct {
		name    string
		enabled bool
		allowed []string
		images  []imageAttachment
		want    string
	}{
		{name: "disabled", images: []imageAttachment{valid}, want: "disabled"},
		{name: "unapproved model", enabled: true, allowed: []string{"allowed-model"}, images: []imageAttachment{valid}, want: "not approved"},
		{name: "remote URL", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "remote", DataURL: "https://private.example/secret"}}, want: "remote URLs"},
		{name: "MIME mismatch", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "wrong MIME", DataURL: strings.Replace(valid.DataURL, "image/png", "image/jpeg", 1)}}, want: "MIME"},
		{name: "invalid base64", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "invalid", DataURL: "data:image/png;base64,AA!!"}}, want: "base64"},
		{name: "base64 whitespace", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "whitespace", DataURL: valid.DataURL + "\n"}}, want: "base64"},
		{name: "base64 bad trailing bits", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "bad bits", DataURL: "data:image/png;base64,AB=="}}, want: "base64"},
		{name: "truncated image", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "truncated", DataURL: "data:image/png;base64," + strings.Split(valid.DataURL, ",")[1][:44]}}, want: "truncated"},
		{name: "unsupported format", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{Name: "svg", DataURL: "data:image/svg+xml;base64,PHN2Zz4="}}, want: "PNG or JPEG"},
		{name: "empty name", enabled: true, allowed: []string{"default-model"}, images: []imageAttachment{{DataURL: valid.DataURL}}, want: "name"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			g := testGateway(t, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1) })
			g.settings.EnableVision, g.settings.AllowedVisionModels = test.enabled, test.allowed
			_, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "look"}, Images: test.images})
			if err == nil || !strings.Contains(err.Error(), test.want) || calls.Load() != 0 {
				t.Fatalf("expected %q without provider call, got %v (calls %d)", test.want, err, calls.Load())
			}
		})
	}
}

func TestVisionSizeCountAndPixelLimits(t *testing.T) {
	valid := testImage(t, "png", 2, 2)
	tooMany := make([]imageAttachment, maxImageCount+1)
	for index := range tooMany {
		tooMany[index] = valid
	}
	if _, err := validateImages(tooMany); statusFor(err) != http.StatusRequestEntityTooLarge {
		t.Fatalf("count limit not enforced: %v", err)
	}
	oversized := imageAttachment{Name: "oversized", DataURL: "data:image/png;base64," + strings.Repeat("A", base64.StdEncoding.EncodedLen(maxImageBytes)+4)}
	if _, err := validateImages([]imageAttachment{oversized}); statusFor(err) != http.StatusRequestEntityTooLarge {
		t.Fatalf("byte limit not enforced: %v", err)
	}
	for _, dimensions := range [][2]int{{4097, 2}, {4096, 4096}, {0, 2}} {
		decoded, _ := base64.StdEncoding.DecodeString(strings.Split(valid.DataURL, ",")[1])
		binary.BigEndian.PutUint32(decoded[16:20], uint32(dimensions[0]))
		binary.BigEndian.PutUint32(decoded[20:24], uint32(dimensions[1]))
		binary.BigEndian.PutUint32(decoded[29:33], crc32.ChecksumIEEE(decoded[12:29]))
		invalid := imageAttachment{Name: "large dimensions", DataURL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(decoded)}
		if _, err := validateImages([]imageAttachment{invalid}); err == nil {
			t.Fatalf("invalid dimensions accepted: %v", dimensions)
		}
	}
	largePixels := testImage(t, "png", 4000, 4000)
	if _, err := validateImages([]imageAttachment{largePixels, largePixels}); err != nil {
		t.Fatalf("exact pixel boundary rejected: %v", err)
	}
	if _, err := validateImages([]imageAttachment{largePixels, largePixels, largePixels}); statusFor(err) != http.StatusRequestEntityTooLarge || !strings.Contains(err.Error(), "pixel") {
		t.Fatalf("aggregate pixel bound not enforced: %v", err)
	}
	// A legal ancillary PNG text chunk supplies compressed-file bytes without
	// allocating a large raster. The total file bytes still count against limits.
	decoded, _ := base64.StdEncoding.DecodeString(strings.Split(valid.DataURL, ",")[1])
	chunk := make([]byte, (3<<20)+128)
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(chunk)-12))
	copy(chunk[4:8], "tEXt")
	copy(chunk[8:16], "comment\x00")
	binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
	padded := append(append(append([]byte{}, decoded[:len(decoded)-12]...), chunk...), decoded[len(decoded)-12:]...)
	largeBytes := imageAttachment{Name: "large bytes", DataURL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(padded)}
	if _, err := validateImages([]imageAttachment{largeBytes, largeBytes, largeBytes, largeBytes}); statusFor(err) != http.StatusRequestEntityTooLarge || !strings.Contains(err.Error(), "12 MiB") {
		t.Fatalf("aggregate byte bound not enforced: %v", err)
	}
}

func TestTextCompatibilityAndImagesAttachToLastUserMessage(t *testing.T) {
	input := []chatMessage{{Role: "system", Content: "system"}, {Role: "user", Content: "earlier"}, {Role: "assistant", Content: "prior answer"}, {Role: "user", Content: "current"}}
	messages, err := buildProviderMessages(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	for index, message := range messages {
		var text string
		if err := json.Unmarshal(message.Content, &text); err != nil || text != input[index].Content || message.Role != input[index].Role {
			t.Fatal("existing string messages changed")
		}
	}
	messages, err = buildProviderMessages(input, []imageAttachment{testImage(t, "png", 1, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if messages[0].Content[0] != '"' || messages[1].Content[0] != '"' || messages[2].Content[0] != '"' || messages[3].Content[0] != '[' {
		t.Fatal("images must attach only to the final user message")
	}
	if _, err := buildProviderMessages([]chatMessage{{Role: "assistant", Content: "response"}}, []imageAttachment{testImage(t, "png", 1, 1)}); err == nil {
		t.Fatal("images without a user message were accepted")
	}
}

func TestImageResourceRejectsUnknownFieldsAndOversizedBodies(t *testing.T) {
	var calls atomic.Int32
	g := testGateway(t, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); _, _ = w.Write([]byte(`{}`)) })
	enableTestVision(g)
	ds := &Datasource{gateway: g}
	for _, test := range []struct {
		body   string
		status int
	}{
		{`{"prompt":{"user":"hello"},"url":"https://attacker.example"}`, http.StatusBadRequest},
		{`{"prompt":{"user":"hello"},"images":[{"name":"x","dataUrl":"data:image/png;base64,AA==","url":"https://attacker.example"}]}`, http.StatusBadRequest},
		{`{"messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":"https://attacker.example"}}]}]}`, http.StatusBadRequest},
		{`{"prompt":{"user":"hello"}} {}`, http.StatusBadRequest},
		{`{"prompt":{"user":"` + strings.Repeat("a", maxRequestBytes) + `"}}`, http.StatusRequestEntityTooLarge},
		{`{"prompt":{"user":"hello"}}` + strings.Repeat(" ", maxRequestBytes), http.StatusRequestEntityTooLarge},
		{strings.Repeat(" ", maxVisionRequestBytes+1), http.StatusRequestEntityTooLarge},
	} {
		response := httptest.NewRecorder()
		ds.handleChat(response, httptest.NewRequest(http.MethodPost, "/chat/completions", strings.NewReader(test.body)))
		if response.Code != test.status {
			t.Fatalf("expected status %d, got %d: %s", test.status, response.Code, response.Body.String())
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid payload reached provider")
	}
	_, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: strings.Repeat("a", maxRequestBytes)}, Images: []imageAttachment{testImage(t, "png", 1, 1)}})
	if statusFor(err) != http.StatusRequestEntityTooLarge {
		t.Fatalf("text cap was bypassed by adding an image: %v", err)
	}
}

func TestCapabilitiesAreLocalFilteredAndContainNoSecrets(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		g := testGateway(t, func(w http.ResponseWriter, _ *http.Request) { t.Error("capabilities must not contact provider") })
		g.settings.EnableVision = enabled
		g.settings.AllowedVisionModels = []string{"default-model", "not-allowed"}
		ds := &Datasource{gateway: g}
		recorder := httptest.NewRecorder()
		ds.handleCapabilities(recorder, httptest.NewRequest(http.MethodGet, "/capabilities", nil))
		var result struct {
			Vision visionCapabilities `json:"vision"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		wantModels := []string{}
		if enabled {
			wantModels = []string{"default-model"}
		}
		if result.Vision.Enabled != enabled || !reflect.DeepEqual(result.Vision.Models, wantModels) || result.Vision.Limits.MaxRequestBytes != 18<<20 || result.Vision.Limits.MaxImages != 8 || result.Vision.Limits.MaxImagePixels != 16_000_000 || !reflect.DeepEqual(result.Vision.MIMETypes, []string{"image/png", "image/jpeg"}) {
			t.Fatalf("incorrect capabilities: %s", recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), "secret-token") || strings.Contains(recorder.Body.String(), "baseUrl") || recorder.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("capabilities exposed credential/configuration or allowed stale caching")
		}
	}
}

func TestProviderRoutesAndGrafanaLLMCredentials(t *testing.T) {
	for _, test := range []struct{ provider, base, expected, model string }{
		{"openwebui", "", "/api", "default-model"},
		{"openwebui", "/api", "/api", "default-model"},
		{"openwebui", "/ai", "/ai/api", "default-model"},
		{"openwebui", "/ai/api/", "/ai/api", "default-model"},
		{"grafana-llm", "/grafana" + models.GrafanaLLMBasePath, "/grafana" + models.GrafanaLLMBasePath, "base"},
	} {
		t.Run(test.provider+test.base, func(t *testing.T) {
			var requests []string
			g := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
				requests = append(requests, req.URL.Path)
				if req.Header.Get("Authorization") != "Bearer secret-token" {
					t.Error("missing server-side credentials")
				}
				if strings.HasSuffix(req.URL.Path, "/models") {
					_, _ = w.Write([]byte(`{"data":[]}`))
					return
				}
				var body providerRequest
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body.Model != test.model || body.Messages[0].Content[0] != '[' {
					t.Error("provider alias or multimodal content changed")
				}
				_, _ = w.Write([]byte(`{}`))
			})
			g.settings.Provider, g.settings.DefaultModel = test.provider, test.model
			g.settings.AllowedModels = []string{test.model}
			g.settings.EnableVision, g.settings.AllowedVisionModels = true, []string{test.model}
			g.baseURL.Path = test.base
			response, err := g.requestModels(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			closeBody(response.Body)
			response, err = g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "look"}, Images: []imageAttachment{testImage(t, "png", 1, 1)}})
			if err != nil {
				t.Fatal(err)
			}
			closeBody(response.Body)
			if !reflect.DeepEqual(requests, []string{test.expected + "/models", test.expected + "/chat/completions"}) {
				t.Fatalf("unexpected paths: %v", requests)
			}
			if test.provider == "grafana-llm" {
				g.settings.Secrets.BearerToken, g.settings.Secrets.APIKey = "", "must-not-fallback"
				if _, err := g.requestModels(context.Background()); err == nil || !strings.Contains(err.Error(), "service-account") {
					t.Fatalf("API key fallback allowed: %v", err)
				}
				g.settings.AllowedModels = []string{"gpt-4.1"}
				if _, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "hello"}, Model: "gpt-4.1"}); err == nil || !strings.Contains(err.Error(), "alias") {
					t.Fatalf("non-alias model permitted: %v", err)
				}
			}
		})
	}
}

func TestProviderTimeoutAndResponseBoundsRemainSanitized(t *testing.T) {
	if _, err := readBounded(io.LimitReader(strings.NewReader(strings.Repeat("s", maxResponseBytes+1)), maxResponseBytes+1)); statusFor(err) != http.StatusBadGateway {
		t.Fatalf("response bound missing: %v", err)
	}
	for _, err := range []error{context.DeadlineExceeded, &url.Error{Op: "POST", URL: "https://secret.example/token", Err: context.DeadlineExceeded}} {
		actual := providerConnectionError(err)
		if statusFor(actual) != http.StatusGatewayTimeout || strings.Contains(actual.Error(), "secret.example") {
			t.Fatalf("timeout not sanitized: %v", actual)
		}
	}
	g := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		select {
		case <-req.Context().Done():
		case <-time.After(100 * time.Millisecond):
		}
	})
	g.client.Timeout = 10 * time.Millisecond
	if _, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "hello"}}); statusFor(err) != http.StatusGatewayTimeout {
		t.Fatalf("real client timeout not surfaced: %v", err)
	}
}

func TestCopilotCapabilitiesAndModelsExposeAgentAliasesWithoutModelControls(t *testing.T) {
	g := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
		t.Error("local aliases/capabilities must not contact Direct Line")
	})
	g.settings.Provider, g.settings.DefaultModel = "copilot-studio", "operations-agent"
	g.settings.AllowedModels = []string{"operations-agent"}
	g.settings.EnableVision, g.settings.AllowedVisionModels = true, []string{"operations-agent"}
	ds := &Datasource{gateway: g}
	recorder := httptest.NewRecorder()
	ds.handleCapabilities(recorder, httptest.NewRequest(http.MethodGet, "/capabilities", nil))
	var capabilities struct {
		Provider           string             `json:"provider"`
		GenerationControls map[string]bool    `json:"generationControls"`
		Vision             visionCapabilities `json:"vision"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &capabilities); err != nil {
		t.Fatal(err)
	}
	if capabilities.Provider != "copilot-studio" || capabilities.GenerationControls["temperature"] || capabilities.GenerationControls["maxOutputTokens"] || len(capabilities.GenerationControls) != 2 || !reflect.DeepEqual(capabilities.Vision.Models, []string{"operations-agent"}) {
		t.Fatalf("incorrect agent capabilities: %s", recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ds.handleModels(recorder, httptest.NewRequest(http.MethodGet, "/models", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"data":[{"id":"operations-agent","object":"model"}],"object":"list"}` {
		t.Fatalf("wrong local alias response: %d %s", recorder.Code, recorder.Body.String())
	}
}
