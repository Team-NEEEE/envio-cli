package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClientReturnsBackendError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != "/v1/login" {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", defaultContentType)
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := io.WriteString(writer, `{
			"success": false,
			"data": null,
			"error": {
				"status": "400 BAD_REQUEST",
				"message": "invalid login request",
				"method": "POST",
				"requestUri": "/v1/login",
				"errors": [
					{
						"field": "email",
						"message": "email is required"
					}
				]
			},
			"timestamp": "2026-05-06T00:04:31.127Z"
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	response, err := client.Do(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/v1/login",
		Body: map[string]string{
			"email": "",
		},
	})
	if err == nil {
		t.Fatal("client.Do() error = nil, want backend error")
	}
	if response == nil || response.Error == nil {
		t.Fatalf("response = %#v, want backend error response", response)
	}

	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("client.Do() error = %T, want *ErrorResponse", err)
	}
	if apiErr.Errors[0].Field != "email" {
		t.Fatalf("apiErr.Errors = %#v", apiErr.Errors)
	}
}

func TestHTTPClientReturnsHTTPResponseErrorForNonJSONBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusBadGateway)
		if _, err := io.WriteString(writer, "<html><body>bad gateway</body></html>"); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	response, err := client.Do(context.Background(), Request{Path: "/health"})
	if err == nil {
		t.Fatal("client.Do() error = nil, want HTTP response error")
	}
	if response != nil {
		t.Fatalf("response = %#v, want nil when common response cannot be decoded", response)
	}

	var responseErr *HTTPResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("client.Do() error = %T, want *HTTPResponseError", err)
	}
	if responseErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("responseErr.StatusCode = %d", responseErr.StatusCode)
	}
	if responseErr.BodySnippet != "<html><body>bad gateway</body></html>" {
		t.Fatalf("responseErr.BodySnippet = %q", responseErr.BodySnippet)
	}
}

func TestHTTPClientHandlesEmptySuccessfulResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	response, err := client.Do(context.Background(), Request{Path: "/empty"})
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	if response == nil || !response.Success {
		t.Fatalf("response = %#v, want successful empty response", response)
	}
}
