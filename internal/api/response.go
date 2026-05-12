package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var backendTimestampLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
}

var backendTimestampLocation = time.FixedZone("KST", 9*60*60)

type Response[T any] struct {
	Timestamp time.Time      `json:"timestamp"`
	Data      T              `json:"data"`
	Error     *ErrorResponse `json:"error,omitempty"`
	Success   bool           `json:"success"`
}

func (r *Response[T]) UnmarshalJSON(data []byte) error {
	var response struct {
		Timestamp responseTimestamp `json:"timestamp"`
		Data      T                 `json:"data"`
		Error     *ErrorResponse    `json:"error,omitempty"`
		Success   bool              `json:"success"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	r.Timestamp = response.Timestamp.Time
	r.Data = response.Data
	r.Error = response.Error
	r.Success = response.Success
	return nil
}

type RawResponse = Response[json.RawMessage]

type responseTimestamp struct {
	time.Time
}

func (t *responseTimestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		t.Time = parsed
		return nil
	}

	for _, layout := range backendTimestampLayouts {
		parsed, err = time.ParseInLocation(layout, value, backendTimestampLocation)
		if err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("parse response timestamp %q: unsupported timestamp format", value)
}

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
