package projectapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

func TestClientLinkProjectSendsPostJSONAndAuthorization(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != linkProjectPath {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Content-Type"); got != projectJSONContentType {
			t.Fatalf("Content-Type = %q", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("Authorization = %q", got)
		}

		var body LinkProjectRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.UserGithubID != "octocat" ||
			body.RepositoryURL != "https://github.com/Team-NEEEE/envio-cli.git" ||
			body.PublicKey != "ssh-rsa AAAA" ||
			body.Owner != "Team-NEEEE" ||
			body.RepoName != "envio-cli" ||
			body.DeviceID != 20 {
			t.Fatalf("body = %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "linked",
				"project": {
					"projectId": 1,
					"projectName": "envio-cli",
					"githubRepoName": "Team-NEEEE/envio-cli",
					"owner": "Team-NEEEE"
				},
				"wrappedMasterKey": "base64-wrapped-key",
				"joinStatus": "APPROVED"
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

	got, err := client.LinkProject(context.Background(), LinkProjectRequest{
		PublicKey:     "ssh-rsa AAAA",
		DeviceID:      20,
		UserGithubID:  "octocat",
		RepositoryURL: "https://github.com/Team-NEEEE/envio-cli.git",
		Owner:         "Team-NEEEE",
		RepoName:      "envio-cli",
	}, "Bearer token-1")
	if err != nil {
		t.Fatalf("LinkProject() error = %v", err)
	}
	if got.Project.ProjectID != 1 || got.WrappedMasterKey != "base64-wrapped-key" || got.JoinStatus != "APPROVED" {
		t.Fatalf("LinkProject() = %#v", got)
	}
}

func TestClientLinkProjectDecodesDirectResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"message": "linked",
			"project": {
				"projectId": 1,
				"projectName": "envio-cli",
				"githubRepoName": "Team-NEEEE/envio-cli",
				"owner": "Team-NEEEE"
			},
			"wrappedMasterKey": "base64-wrapped-key",
			"joinStatus": "APPROVED"
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	got, err := client.LinkProject(context.Background(), LinkProjectRequest{}, "")
	if err != nil {
		t.Fatalf("LinkProject() error = %v", err)
	}
	if got.Project.Owner != "Team-NEEEE" || got.JoinStatus != "APPROVED" {
		t.Fatalf("LinkProject() = %#v", got)
	}
}

func TestClientLinkProjectReturnsBackendError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusConflict)
		if _, err := io.WriteString(writer, `{
			"timestamp": "2026-04-30T13:30:00",
			"status": 409,
			"code": "JOIN_STATUS_PENDING",
			"message": "프로젝트 가입 승인이 아직 완료되지 않았습니다.",
			"path": "/api/core/projects/link"
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	_, err = client.LinkProject(context.Background(), LinkProjectRequest{}, "")
	var apiErr *api.ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("LinkProject() error = %T, want *api.ErrorResponse", err)
	}
	if apiErr.Code != "JOIN_STATUS_PENDING" {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}
