package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestResponseUnmarshalsBackendError(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"success": false,
		"data": null,
		"error": {
			"status": "100 CONTINUE",
			"message": "string",
			"method": "string",
			"requestUri": "string",
			"errors": [
				{
					"field": "string",
					"message": "string"
				}
			]
		},
		"timestamp": "2026-05-06T00:04:31.127Z"
	}`)

	var response RawResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Success {
		t.Fatal("response.Success = true, want false")
	}
	if response.Error == nil {
		t.Fatal("response.Error = nil, want backend error")
	}
	if response.Error.Status != ErrorStatus("100 CONTINUE") {
		t.Fatalf("response.Error.Status = %q", response.Error.Status)
	}
	if len(response.Error.Errors) != 1 || response.Error.Errors[0].Field != "string" {
		t.Fatalf("response.Error.Errors = %#v", response.Error.Errors)
	}
	if response.Timestamp.Format(time.RFC3339Nano) != "2026-05-06T00:04:31.127Z" {
		t.Fatalf("response.Timestamp = %s", response.Timestamp.Format(time.RFC3339Nano))
	}
}

func TestDecodeDataUnmarshalsCommonResponseData(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"success": true,
		"data": {
			"projectId": "envio"
		},
		"error": null,
		"timestamp": "2026-05-06T00:04:31.127Z"
	}`)

	var response RawResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	data, err := DecodeData[struct {
		ProjectID string `json:"projectId"`
	}](&response)
	if err != nil {
		t.Fatalf("DecodeData() error = %v", err)
	}
	if data.ProjectID != "envio" {
		t.Fatalf("data.ProjectID = %q", data.ProjectID)
	}
}
