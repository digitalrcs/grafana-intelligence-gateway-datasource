package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DigitalRCS/grafana-intelligence-gateway-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type Datasource struct {
	backend.CallResourceHandler
	gateway *gateway
}

func NewDatasource(ctx context.Context, source backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := models.LoadPluginSettings(source)
	if err != nil {
		return nil, err
	}
	gateway, err := newGateway(ctx, settings)
	if err != nil {
		return nil, err
	}
	ds := &Datasource{gateway: gateway}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /models", ds.handleModels)
	mux.HandleFunc("GET /capabilities", ds.handleCapabilities)
	mux.HandleFunc("POST /chat/completions", ds.handleChat)
	mux.HandleFunc("POST /analyze", ds.handleChat)
	ds.CallResourceHandler = httpadapter.New(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		setResourceHeaders(w)
		mux.ServeHTTP(w, req)
	}))
	return ds, nil
}

func (d *Datasource) Dispose() { d.gateway.close() }

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, query := range req.Queries {
		response.Responses[query.RefID] = d.query(ctx, query)
	}
	return response, nil
}

type queryModel struct {
	SystemPrompt    string   `json:"systemPrompt"`
	UserPrompt      string   `json:"userPrompt"`
	Model           string   `json:"model"`
	Temperature     *float64 `json:"temperature"`
	MaxOutputTokens *int     `json:"maxOutputTokens"`
}

func (d *Datasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var model queryModel
	if err := json.Unmarshal(query.JSON, &model); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "invalid query")
	}
	providerResponse, err := d.gateway.requestChat(ctx, gatewayRequest{
		Prompt: &prompt{System: model.SystemPrompt, User: model.UserPrompt},
		Model:  model.Model, Temperature: model.Temperature, MaxOutputTokens: model.MaxOutputTokens,
	})
	if err != nil {
		return backend.ErrDataResponse(backend.StatusUnknown, err.Error())
	}
	defer closeBody(providerResponse.Body)
	body, err := readBounded(providerResponse.Body)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusUnknown, err.Error())
	}
	answer, err := extractAnswer(body)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusUnknown, err.Error())
	}
	frame := data.NewFrame("AI assessment", data.NewField("answer", nil, []string{answer}))
	frame.RefID = query.RefID
	return backend.DataResponse{Frames: data.Frames{frame}}
}

func extractAnswer(body []byte) (string, error) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
		OutputText string `json:"output_text"`
		Text       string `json:"text"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", errorsForAnswer("provider returned invalid JSON")
	}
	answer := payload.OutputText
	if answer == "" {
		answer = payload.Text
	}
	if answer == "" && len(payload.Choices) > 0 {
		answer = payload.Choices[0].Message.Content
		if answer == "" {
			answer = payload.Choices[0].Text
		}
	}
	if strings.TrimSpace(answer) == "" {
		return "", errorsForAnswer("provider returned no text content")
	}
	return strings.TrimSpace(answer), nil
}

func errorsForAnswer(message string) error { return fmt.Errorf("%s", message) }

func (d *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.gateway.settings.Provider == "copilot-studio" {
		if err := d.gateway.checkCopilotHealth(ctx); err != nil {
			return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
		}
		return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "Direct Line credentials and connectivity are valid; agent inference and end-user SSO were not tested"}, nil
	}
	response, err := d.gateway.requestModels(ctx)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	closeBody(response.Body)
	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "Provider connection and credentials are valid"}, nil
}

func (d *Datasource) handleModels(w http.ResponseWriter, req *http.Request) {
	setResourceHeaders(w)
	response, err := d.gateway.requestModels(req.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	defer closeBody(response.Body)
	body, err := readBounded(response.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	filtered, err := filterModelList(body, d.gateway.settings)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(filtered)
}

func filterModelList(body []byte, settings *models.PluginSettings) ([]byte, error) {
	var payload struct {
		Data   json.RawMessage `json:"data"`
		Models json.RawMessage `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, errorsForAnswer("provider returned an invalid model list")
	}
	entriesJSON := payload.Data
	if len(entriesJSON) == 0 || string(entriesJSON) == "null" {
		entriesJSON = payload.Models
	}
	var entries []json.RawMessage
	if len(entriesJSON) == 0 || string(entriesJSON) == "null" || json.Unmarshal(entriesJSON, &entries) != nil {
		return nil, errorsForAnswer("provider returned an invalid model list")
	}
	seen := make(map[string]struct{}, len(entries))
	allowed := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		var modelID string
		if err := json.Unmarshal(entry, &modelID); err != nil {
			var item struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(entry, &item); err != nil {
				continue
			}
			modelID = item.ID
			if modelID == "" {
				modelID = item.Name
			}
		}
		modelID = strings.TrimSpace(modelID)
		if modelID == "" || !settings.ModelAllowed(modelID) {
			continue
		}
		if _, exists := seen[modelID]; exists {
			continue
		}
		seen[modelID] = struct{}{}
		allowed = append(allowed, map[string]string{"id": modelID, "object": "model"})
	}
	return json.Marshal(map[string]any{"object": "list", "data": allowed})
}

func (d *Datasource) handleChat(w http.ResponseWriter, req *http.Request) {
	setResourceHeaders(w)
	defer closeBody(req.Body)
	if err := d.gateway.acquire(req.Context()); err != nil {
		writeError(w, err)
		return
	}
	defer d.gateway.release()
	req.Body = http.MaxBytesReader(w, req.Body, maxVisionRequestBytes)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "encoded request exceeds the 18 MiB limit")
		} else {
			writeJSONError(w, http.StatusBadRequest, "could not read request body")
		}
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var request gatewayRequest
	if err := decoder.Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		writeJSONError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return
	}
	if len(request.Images) == 0 && len(body) > maxRequestBytes {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "text request exceeds the 1 MiB limit")
		return
	}
	response, err := d.gateway.requestChatWithSlot(req.Context(), request)
	if err != nil {
		writeError(w, err)
		return
	}
	defer closeBody(response.Body)
	copyResponse(w, response)
}

func copyResponse(w http.ResponseWriter, response *http.Response) {
	setResourceHeaders(w)
	contentType := response.Header.Get("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "text/event-stream") {
		writeJSONError(w, http.StatusBadGateway, "provider returned an unsupported streaming response")
		return
	}
	body, err := readBounded(response.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	// Never serve provider HTML or another active document from a Grafana origin.
	// Some compatible endpoints omit or mislabel the media type; accept their
	// valid JSON object, but always return the safe, explicit JSON media type.
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(body) {
		writeJSONError(w, http.StatusBadGateway, "provider returned an invalid JSON response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func writeError(w http.ResponseWriter, err error) { writeJSONError(w, statusFor(err), err.Error()) }

func closeBody(body io.Closer) { _ = body.Close() }

func setResourceHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	setResourceHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"message": message}})
}
