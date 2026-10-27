package plugin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"unicode"
)

const (
	maxImageCount         = 8
	maxImageBytes         = 4 << 20
	maxTotalImageBytes    = 12 << 20
	maxImageDimension     = 4096
	maxImagePixels        = 16_000_000
	maxTotalImagePixels   = 32_000_000
	maxVisionRequestBytes = 18 << 20
)

type imageAttachment struct {
	Name    string `json:"name"`
	DataURL string `json:"dataUrl"`
}

// The incoming message schema remains text-only. Only validated attachments can
// produce image_url parts, so callers cannot supply arbitrary remote image URLs.
type providerContentPart struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL *providerImageURL `json:"image_url,omitempty"`
}

type providerImageURL struct {
	URL string `json:"url"`
}

type providerMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type imageLimits struct {
	MaxImages           int `json:"maxImages"`
	MaxImageBytes       int `json:"maxImageBytes"`
	MaxTotalImageBytes  int `json:"maxTotalImageBytes"`
	MaxImageWidth       int `json:"maxImageWidth"`
	MaxImageHeight      int `json:"maxImageHeight"`
	MaxImagePixels      int `json:"maxImagePixels"`
	MaxTotalImagePixels int `json:"maxTotalImagePixels"`
	MaxRequestBytes     int `json:"maxRequestBytes"`
}

type visionCapabilities struct {
	Enabled   bool        `json:"enabled"`
	Models    []string    `json:"models"`
	Limits    imageLimits `json:"limits"`
	MIMETypes []string    `json:"mimeTypes"`
}

type generationControls struct {
	Temperature     bool `json:"temperature"`
	MaxOutputTokens bool `json:"maxOutputTokens"`
}

type gatewayCapabilities struct {
	Provider           string             `json:"provider"`
	Vision             visionCapabilities `json:"vision"`
	GenerationControls generationControls `json:"generationControls"`
}

func (d *Datasource) handleCapabilities(w http.ResponseWriter, _ *http.Request) {
	setResourceHeaders(w)
	allowed := make([]string, 0, len(d.gateway.settings.AllowedVisionModels))
	for _, model := range d.gateway.settings.AllowedVisionModels {
		if d.gateway.settings.VisionModelAllowed(model) {
			allowed = append(allowed, model)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gatewayCapabilities{Provider: d.gateway.settings.Provider,
		GenerationControls: generationControls{
			Temperature: d.gateway.settings.Provider != "copilot-studio", MaxOutputTokens: d.gateway.settings.Provider != "copilot-studio",
		}, Vision: visionCapabilities{
			Enabled: d.gateway.settings.EnableVision, Models: allowed,
			MIMETypes: []string{"image/png", "image/jpeg"},
			Limits: imageLimits{
				MaxImages: maxImageCount, MaxImageBytes: maxImageBytes, MaxTotalImageBytes: maxTotalImageBytes,
				MaxImageWidth: maxImageDimension, MaxImageHeight: maxImageDimension,
				MaxImagePixels: maxImagePixels, MaxTotalImagePixels: maxTotalImagePixels,
				MaxRequestBytes: maxVisionRequestBytes,
			},
		}})
}

func imageError(index int, message string) error {
	return &gatewayError{status: http.StatusBadRequest, message: fmt.Sprintf("image %d: %s", index+1, message)}
}

func validateImages(images []imageAttachment) ([]providerContentPart, error) {
	if len(images) > maxImageCount {
		return nil, &gatewayError{status: http.StatusRequestEntityTooLarge, message: "at most 8 images may be attached"}
	}
	parts := make([]providerContentPart, 0, 2*len(images))
	totalBytes, totalPixels := 0, int64(0)
	for index, attachment := range images {
		name := strings.TrimSpace(attachment.Name)
		if name == "" || len(name) > 256 || strings.ContainsFunc(name, unicode.IsControl) {
			return nil, imageError(index, "name must contain 1-256 bytes without control characters")
		}
		var encoded, expectedFormat string
		switch {
		case strings.HasPrefix(attachment.DataURL, "data:image/png;base64,"):
			encoded, expectedFormat = strings.TrimPrefix(attachment.DataURL, "data:image/png;base64,"), "png"
		case strings.HasPrefix(attachment.DataURL, "data:image/jpeg;base64,"):
			encoded, expectedFormat = strings.TrimPrefix(attachment.DataURL, "data:image/jpeg;base64,"), "jpeg"
		default:
			return nil, imageError(index, "only inline base64 PNG or JPEG images are supported; remote URLs are not permitted")
		}
		if len(encoded) > base64.StdEncoding.EncodedLen(maxImageBytes) {
			return nil, &gatewayError{status: http.StatusRequestEntityTooLarge, message: fmt.Sprintf("image %d exceeds the 4 MiB decoded image limit", index+1)}
		}
		if encoded == "" || strings.ContainsAny(encoded, "\r\n\t ") {
			return nil, imageError(index, "image data must use strict standard base64 without whitespace")
		}
		decoded, err := base64.StdEncoding.Strict().DecodeString(encoded)
		if err != nil {
			return nil, imageError(index, "image data is not valid base64")
		}
		if len(decoded) > maxImageBytes {
			return nil, &gatewayError{status: http.StatusRequestEntityTooLarge, message: fmt.Sprintf("image %d exceeds the 4 MiB decoded image limit", index+1)}
		}
		totalBytes += len(decoded)
		if totalBytes > maxTotalImageBytes {
			return nil, &gatewayError{status: http.StatusRequestEntityTooLarge, message: "images exceed the 12 MiB total decoded image limit"}
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(decoded))
		if err != nil {
			return nil, imageError(index, "image is invalid or truncated")
		}
		if format != expectedFormat {
			return nil, imageError(index, "declared MIME type does not match the image format")
		}
		pixels := int64(config.Width) * int64(config.Height)
		if config.Width <= 0 || config.Height <= 0 || config.Width > maxImageDimension || config.Height > maxImageDimension || pixels > maxImagePixels {
			return nil, imageError(index, "image dimensions must be at most 4096 per side and 16000000 pixels")
		}
		totalPixels += pixels
		if totalPixels > maxTotalImagePixels {
			return nil, &gatewayError{status: http.StatusRequestEntityTooLarge, message: "images exceed the 32000000 total pixel limit"}
		}
		// Decode after bounded header checks: a valid header alone does not prove
		// the compressed image is complete. Never include decoder errors or data
		// in the response because they can contain provider/user content.
		if _, actualFormat, err := image.Decode(bytes.NewReader(decoded)); err != nil || actualFormat != expectedFormat {
			return nil, imageError(index, "image is invalid or truncated")
		}
		parts = append(parts,
			providerContentPart{Type: "text", Text: fmt.Sprintf("Image %d: %s", index+1, name)},
			providerContentPart{Type: "image_url", ImageURL: &providerImageURL{URL: attachment.DataURL}},
		)
	}
	return parts, nil
}

func buildProviderMessages(messages []chatMessage, images []imageAttachment) ([]providerMessage, error) {
	result := make([]providerMessage, len(messages))
	lastUser := -1
	for index, message := range messages {
		content, err := json.Marshal(message.Content)
		if err != nil {
			return nil, err
		}
		result[index] = providerMessage{Role: message.Role, Content: content}
		if message.Role == "user" {
			lastUser = index
		}
	}
	if len(images) == 0 {
		return result, nil
	}
	if lastUser < 0 {
		return nil, &gatewayError{status: http.StatusBadRequest, message: "images require a user message"}
	}
	parts, err := validateImages(images)
	if err != nil {
		return nil, err
	}
	content := append([]providerContentPart{{Type: "text", Text: messages[lastUser].Content}}, parts...)
	result[lastUser].Content, err = json.Marshal(content)
	return result, err
}
