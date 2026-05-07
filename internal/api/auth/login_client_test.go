package authapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientStartLogin(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != loginStartPath {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.URL.Query().Get("redirectType"); got != "CLI" {
			t.Fatalf("redirectType = %q", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "login started",
				"loginSessionId": "session-1",
				"authUrl": "https://github.com/login/oauth/authorize",
				"expiresIn": 300
			},
			"error": null,
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

	response, err := client.StartLogin(context.Background())
	if err != nil {
		t.Fatalf("StartLogin() error = %v", err)
	}
	if response.LoginSessionID != "session-1" {
		t.Fatalf("response.LoginSessionID = %q", response.LoginSessionID)
	}
}

func TestClientGetLoginStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != loginStatusPath {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}
		if got := request.URL.Query().Get("loginSessionId"); got != "" {
			t.Fatalf("query loginSessionId = %q", got)
		}

		var body LoginStatusRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.LoginSessionID != "session-1" {
			t.Fatalf("body = %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "completed",
				"status": "COMPLETED",
				"githubId": "octocat",
				"email": "mona@example.com"
			},
			"error": null,
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

	response, err := client.GetLoginStatus(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetLoginStatus() error = %v", err)
	}
	if response.Status != "COMPLETED" {
		t.Fatalf("response.Status = %q", response.Status)
	}
}

func TestClientRegisterKey(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != registerKeyPath {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}

		var body RegisterKeyRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.LoginSessionID != "session-1" || body.PublicKey != "public-key" || body.DeviceName != "desktop" {
			t.Fatalf("body = %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "registered",
				"userId": 10,
				"githubId": "octocat",
				"email": "mona@example.com",
				"deviceId": 20
			},
			"error": null,
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

	response, err := client.RegisterKey(context.Background(), RegisterKeyRequest{
		LoginSessionID: "session-1",
		PublicKey:      "public-key",
		DeviceName:     "desktop",
	})
	if err != nil {
		t.Fatalf("RegisterKey() error = %v", err)
	}
	if response.UserID != 10 || response.DeviceID != 20 {
		t.Fatalf("response = %#v", response)
	}
}
