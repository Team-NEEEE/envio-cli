package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Response[T any] struct {
	Success   bool           `json:"success"`
	Data      T              `json:"data"`
	Error     *ErrorResponse `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type RawResponse = Response[json.RawMessage]

type ErrorStatus string

type ErrorResponse struct {
	Status     ErrorStatus  `json:"status"`
	Message    string       `json:"message"`
	Method     string       `json:"method"`
	RequestURI string       `json:"requestUri"`
	Errors     []FieldError `json:"errors,omitempty"`
}

func (e *ErrorResponse) Error() string {
	if e == nil {
		return ""
	}

	parts := make([]string, 0, 3)
	if e.Status != "" {
		parts = append(parts, string(e.Status))
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.RequestURI != "" {
		parts = append(parts, e.RequestURI)
	}
	if len(parts) == 0 {
		return "api request failed"
	}
	return strings.Join(parts, ": ")
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func DecodeData[T any](response *RawResponse) (T, error) {
	var value T
	if response == nil || len(response.Data) == 0 || string(response.Data) == "null" {
		return value, nil
	}
	if err := json.Unmarshal(response.Data, &value); err != nil {
		return value, fmt.Errorf("decode api response data: %w", err)
	}
	return value, nil
}
