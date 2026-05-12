package authapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClientStartLogin 로그인 시작 API 요청과 응답 디코딩을 검증한다.
func TestClientStartLogin(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// 로그인 시작 요청은 POST 메서드를 사용해야 한다.
		if request.Method != http.MethodPost {
			t.Fatalf("request.Method = %s", request.Method)
		}
		if request.URL.Path != loginStartPath {
			t.Fatalf("request.URL.Path = %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(writer, `{
			"success": true,
			"data": {
				"message": "login started",
				"loginSessionId": "session-1",
				"loginUrl": "https://github.com/login/oauth/authorize",
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
	// 공통 응답의 data 필드가 LoginStartResponse로 디코딩되는지 확인한다.
	if response.LoginSessionID != "session-1" {
		t.Fatalf("response.LoginSessionID = %q", response.LoginSessionID)
	}
	if response.AuthURL != "https://github.com/login/oauth/authorize" {
		t.Fatalf("response.AuthURL = %q", response.AuthURL)
	}
}

// TestClientGetLoginStatus 로그인 상태 조회 API 요청과 응답 디코딩을 검증한다.
func TestClientGetLoginStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// loginSessionId는 query가 아니라 JSON body로 전송해야 한다.
		if request.Method != http.MethodPost {
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
	// 서버 상태 문자열이 응답 구조체로 전달되는지 확인한다.
	if response.Status != "COMPLETED" {
		t.Fatalf("response.Status = %q", response.Status)
	}
}

// TestClientRegisterKey 공개키 저장 API 요청과 응답 디코딩을 검증한다.
func TestClientRegisterKey(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// 공개키 저장 요청은 POST JSON body로 session, public key, device name을 전송해야 한다.
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
	// 등록 결과의 사용자 ID와 기기 ID가 응답 구조체로 전달되는지 확인한다.
	if response.UserID != 10 || response.DeviceID != 20 {
		t.Fatalf("response = %#v", response)
	}
}
