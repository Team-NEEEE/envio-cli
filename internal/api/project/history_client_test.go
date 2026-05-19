package projectapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHistoryGetsProjectHistories(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != "/api/cli/projects/1/history" {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("Authorization = %q", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "histories",
				"histories": [
					{
						"history_id": 11,
						"project_id": 1,
						"version_id": 8,
						"base_version_id": 7,
						"github_id": "octocat",
						"created_at": "2026-04-20 11:00:00",
						"encrypted_environment": {
							"algorithm": "aes-256-gcm",
							"nonce": "nonce",
							"ciphertext": "ciphertext"
						}
					}
				]
			},
			"error": null,
			"timestamp": "2026-05-12T01:02:17"
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	got, err := client.History(context.Background(), 1, "Bearer token-1")
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(got.Histories) != 1 {
		t.Fatalf("histories length = %d", len(got.Histories))
	}
	history := got.Histories[0]
	if history.ProjectID != 1 || history.VersionID != 8 || history.BaseVersionID != 7 || history.HistoryID != 11 {
		t.Fatalf("history = %#v", history)
	}
	if history.GithubID != "octocat" || history.EncryptedEnvironment["ciphertext"] != "ciphertext" {
		t.Fatalf("history = %#v", history)
	}
}

func TestClientHistoryAcceptsDirectArrayData(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != "/api/cli/projects/1/history" {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": [
				{
					"histories_id": 11,
					"project_id": 1,
					"version_id": 8,
					"base_version_id": 7,
					"github_id": "octocat",
					"created_at": "2026-04-20T11:00:00",
					"encrypted_environment": {
						"algorithm": "aes-256-gcm",
						"nonce": "nonce",
						"ciphertext": "ciphertext"
					}
				}
			],
			"error": null,
			"timestamp": "2026-05-12T01:02:17"
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	got, err := client.History(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(got.Histories) != 1 {
		t.Fatalf("histories length = %d", len(got.Histories))
	}
	history := got.Histories[0]
	if history.HistoryID != 11 || history.ProjectID != 1 || history.VersionID != 8 || history.BaseVersionID != 7 {
		t.Fatalf("history = %#v", history)
	}
	if history.CreatedAt != "2026-04-20T11:00:00" || history.EncryptedEnvironment["ciphertext"] != "ciphertext" {
		t.Fatalf("history = %#v", history)
	}
}
