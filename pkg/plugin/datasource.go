package plugin

import (
	"context"
	"encoding/json"
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
	mux.HandleFunc("POST /chat/completions", ds.handleChat)
	mux.HandleFunc("POST /analyze", ds.handleChat)
	ds.CallResourceHandler = httpadapter.New(mux)
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
	defer providerResponse.Body.Close()
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
	response, err := d.gateway.requestModels(ctx)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	response.Body.Close()
	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "Provider connection and credentials are valid"}, nil
}

func (d *Datasource) handleModels(w http.ResponseWriter, req *http.Request) {
	response, err := d.gateway.requestModels(req.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	defer response.Body.Close()
	copyResponse(w, response)
}

func (d *Datasource) handleChat(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	req.Body = http.MaxBytesReader(w, req.Body, maxRequestBytes)
	decoder := json.NewDecoder(req.Body)
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
	response, err := d.gateway.requestChat(req.Context(), request)
	if err != nil {
		writeError(w, err)
		return
	}
	defer response.Body.Close()
	copyResponse(w, response)
}

func copyResponse(w http.ResponseWriter, response *http.Response) {
	contentType := response.Header.Get("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "text/event-stream") {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, io.LimitReader(response.Body, maxResponseBytes))
		return
	}
	body, err := readBounded(response.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	if contentType == "" {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func writeError(w http.ResponseWriter, err error) { writeJSONError(w, statusFor(err), err.Error()) }

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"message": message}})
}
