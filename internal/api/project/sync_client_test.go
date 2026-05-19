package projectapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientPullLatestSendsProjectAndGithubUserID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != "/api/core/projects/1/pull/latest" {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.URL.Query().Get("githubUserId"); got != "octocat" {
			t.Fatalf("githubUserId = %q", got)
		}
		if got := request.URL.Query().Get("deviceId"); got != "20" {
			t.Fatalf("deviceId = %q", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("Authorization = %q", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "pulled",
				"historyId": 11,
				"projectId": 1,
				"envName": "envio-cli",
				"versionId": 3,
				"encryptedEnvironment": {
					"algorithm": "aes-256-gcm",
					"nonce": "nonce",
					"ciphertext": "ciphertext"
				},
				"wrappedMasterKey": "wrapped-key",
				"createdAt": "2026-05-12T01:01:17",
				"updatedAt": "2026-05-12T01:02:17"
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

	got, err := client.PullLatest(context.Background(), 1, "octocat", 20, "Bearer token-1")
	if err != nil {
		t.Fatalf("PullLatest() error = %v", err)
	}
	if got.ProjectID != 1 || got.VersionID != 3 || got.HistoryID != 11 {
		t.Fatalf("PullLatest() = %#v", got)
	}
	if got.EncryptedEnvironment["ciphertext"] != "ciphertext" {
		t.Fatalf("encryptedEnvironment = %#v", got.EncryptedEnvironment)
	}
	if got.WrappedMasterKey != "wrapped-key" {
		t.Fatalf("wrappedMasterKey = %q", got.WrappedMasterKey)
	}
}

func TestClientPushSendsEncryptedEnvironmentAndParentVersion(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != "/api/core/projects/1/push" {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}

		var body ProjectPushRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.GithubUserID != "octocat" ||
			body.ParentVersionID != 2 ||
			body.EncryptedEnvironment["ciphertext"] != "ciphertext" {
			t.Fatalf("body = %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "pushed",
				"historyId": 12,
				"projectId": 1,
				"envName": "envio-cli",
				"versionId": 3,
				"parentVersionId": 2
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

	got, err := client.Push(context.Background(), 1, ProjectPushRequest{
		GithubUserID:         "octocat",
		ParentVersionID:      2,
		EncryptedEnvironment: map[string]any{"ciphertext": "ciphertext"},
	}, "")
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if got.ProjectID != 1 || got.VersionID != 3 || got.ParentVersionID != 2 {
		t.Fatalf("Push() = %#v", got)
	}
}
