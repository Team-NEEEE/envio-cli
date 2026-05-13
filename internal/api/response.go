package api

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err == nil && object != nil {
		_, hasSuccess := object["success"]
		_, hasData := object["data"]
		_, hasError := object["error"]
		if !hasSuccess && !hasData && !hasError {
			if looksLikeDirectError(object) {
				var responseError ErrorResponse
				if err := json.Unmarshal(data, &responseError); err != nil {
					return err
				}
				r.Error = &responseError
				r.Success = false
				return nil
			}

			if err := setDirectData(&r.Data, data); err != nil {
				return err
			}
			r.Success = true
			return nil
		}
	}

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

func setDirectData[T any](target *T, data []byte) error {
	if raw, ok := any(target).(*json.RawMessage); ok {
		*raw = append((*raw)[:0], data...)
		return nil
	}
	return json.Unmarshal(data, target)
}

func looksLikeDirectError(object map[string]json.RawMessage) bool {
	if _, ok := object["status"]; !ok {
		return false
	}
	if _, ok := object["message"]; !ok {
		return false
	}
	if _, ok := object["code"]; ok {
		return true
	}
	if _, ok := object["path"]; ok {
		return true
	}
	return false
}

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

func (s *ErrorStatus) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		*s = ErrorStatus(value)
		return nil
	}

	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*s = ErrorStatus(strconv.Itoa(number))
		return nil
	}

	return fmt.Errorf("parse error status %s: expected string or number", string(data))
}

type ErrorResponse struct {
	Timestamp  time.Time    `json:"timestamp,omitempty"`
	Status     ErrorStatus  `json:"status"`
	Code       string       `json:"code,omitempty"`
	Message    string       `json:"message"`
	Method     string       `json:"method"`
	RequestURI string       `json:"requestUri"`
	Path       string       `json:"path,omitempty"`
	Errors     []FieldError `json:"errors,omitempty"`
}

func (e *ErrorResponse) UnmarshalJSON(data []byte) error {
	var response struct {
		Timestamp  responseTimestamp `json:"timestamp"`
		Status     ErrorStatus       `json:"status"`
		Code       string            `json:"code,omitempty"`
		Message    string            `json:"message"`
		Method     string            `json:"method"`
		RequestURI string            `json:"requestUri"`
		Path       string            `json:"path,omitempty"`
		Errors     []FieldError      `json:"errors,omitempty"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	e.Timestamp = response.Timestamp.Time
	e.Status = response.Status
	e.Code = response.Code
	e.Message = response.Message
	e.Method = response.Method
	e.RequestURI = response.RequestURI
	e.Path = response.Path
	e.Errors = response.Errors
	if e.RequestURI == "" {
		e.RequestURI = e.Path
	}
	return nil
}

func (e *ErrorResponse) Error() string {
	if e == nil {
		return ""
	}

	parts := make([]string, 0, 3)
	if e.Status != "" {
		parts = append(parts, string(e.Status))
	}
	if e.Code != "" {
		parts = append(parts, e.Code)
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
