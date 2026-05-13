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

func TestClientCreateProjectSendsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != createProjectPath {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Content-Type"); got != projectJSONContentType {
			t.Fatalf("Content-Type = %q", got)
		}

		var body CreateProjectRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.RepositoryURL != "https://github.com/Team-NEEEE/envio-cli.git" {
			t.Fatalf("body = %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(writer, `{
			"message": "프로젝트 생성에 성공했습니다.",
			"projectId": 1,
			"projectName": "envio-cli",
			"githubRepoName": "Team-NEEEE/envio-cli",
			"installationId": 9,
			"members": [
				{
					"userId": 10,
					"userDeviceId": 20,
					"githubId": "octocat",
					"publicKey": "-----BEGIN PUBLIC KEY-----test-----END PUBLIC KEY-----",
					"projectRole": "ADMIN"
				}
			]
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	got, err := client.CreateProject(context.Background(), CreateProjectRequest{
		RepositoryURL: "https://github.com/Team-NEEEE/envio-cli.git",
	})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if got.ProjectID != 1 || got.Members[0].UserDeviceID != 20 {
		t.Fatalf("CreateProject() = %#v", got)
	}
}

func TestClientSaveWrappedKeysSendsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != "/api/projects/1/wrapped-keys" {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Content-Type"); got != projectJSONContentType {
			t.Fatalf("Content-Type = %q", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("Authorization = %q", got)
		}

		var body SaveWrappedKeysRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.WrappedKeys) != 1 ||
			body.WrappedKeys[0].UserID != 10 ||
			body.WrappedKeys[0].UserDeviceID != 20 ||
			body.WrappedKeys[0].EncryptedKey != "base64-wrapped-key" {
			t.Fatalf("body = %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"message": "프로젝트 마스터 키 분배 등록에 성공했습니다.",
			"projectId": 1,
			"updatedCount": 1
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	got, err := client.SaveWrappedKeys(context.Background(), 1, SaveWrappedKeysRequest{
		WrappedKeys: []WrappedKey{
			{UserID: 10, UserDeviceID: 20, EncryptedKey: "base64-wrapped-key"},
		},
	}, "Bearer token-1")
	if err != nil {
		t.Fatalf("SaveWrappedKeys() error = %v", err)
	}
	if got.ProjectID != 1 || got.UpdatedCount != 1 {
		t.Fatalf("SaveWrappedKeys() = %#v", got)
	}
}

func TestClientCreateProjectReturnsBackendError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusConflict)
		if _, err := io.WriteString(writer, `{
			"timestamp": "2026-04-30T13:30:00",
			"status": 409,
			"code": "PROJECT_ALREADY_EXISTS",
			"message": "이미 등록된 프로젝트입니다.",
			"path": "/api/v1/cli/create"
		}`); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	_, err = client.CreateProject(context.Background(), CreateProjectRequest{})
	var apiErr *api.ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("CreateProject() error = %T, want *api.ErrorResponse", err)
	}
	if apiErr.Code != "PROJECT_ALREADY_EXISTS" || apiErr.Message != "이미 등록된 프로젝트입니다." {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}

func TestClientReturnsHTTPErrorForNonJSONFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusBadGateway)
		if _, err := io.WriteString(writer, "<html>bad gateway</html>"); err != nil {
			t.Fatalf("write response body: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	_, err = client.SaveWrappedKeys(context.Background(), 1, SaveWrappedKeysRequest{}, "")
	var httpErr *api.HTTPResponseError
	if !errors.As(err, &httpErr) {
		t.Fatalf("SaveWrappedKeys() error = %T, want *api.HTTPResponseError", err)
	}
	if httpErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("httpErr = %#v", httpErr)
	}
}
