package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	authapi "github.com/Team-NEEEE/envio-cli/internal/api/auth"
	"github.com/Team-NEEEE/envio-cli/internal/config"
)

type fakeLoginAPI struct {
	// startResp/startErr는 StartLogin 호출 결과를 테스트 케이스별로 제어한다.
	startResp *authapi.LoginStartResponse
	startErr  error
	started   bool

	// statusResp/statusErr는 로그인 상태 조회 결과를 제어하고,
	// statusID는 서비스가 올바른 loginSessionID를 전달했는지 확인하는 데 사용한다.
	statusResp *authapi.LoginStatusResponse
	statusErr  error
	statusID   string

	// registerResp/registerErr는 공개키 등록 결과를 제어하고,
	// registerReq는 LoginService가 API 경계로 넘긴 요청 본문을 검증하는 데 사용한다.
	registerResp *authapi.RegisterKeyResponse
	registerErr  error
	registerReq  authapi.RegisterKeyRequest
}

func (f *fakeLoginAPI) StartLogin(context.Context) (*authapi.LoginStartResponse, error) {
	// startLogin은 브라우저 실행 전의 첫 API 호출만 검증하면 되므로 호출 여부를 기록한다.
	f.started = true
	return f.startResp, f.startErr
}

func (f *fakeLoginAPI) GetLoginStatus(_ context.Context, loginSessionID string) (*authapi.LoginStatusResponse, error) {
	// waitLoginComplete/getLoginStatus가 같은 세션 ID로 상태를 조회하는지 확인한다.
	f.statusID = loginSessionID
	return f.statusResp, f.statusErr
}

func (f *fakeLoginAPI) RegisterKey(
	_ context.Context,
	req authapi.RegisterKeyRequest,
) (*authapi.RegisterKeyResponse, error) {
	// registerKey가 세션 ID, 공개키, 기기명을 손실 없이 전달하는지 확인하기 위해 요청을 보관한다.
	f.registerReq = req
	return f.registerResp, f.registerErr
}

func TestLoginReturnsClientError(t *testing.T) {
	// NewLoginService에서 HTTP 클라이언트 생성에 실패한 경우에는 브라우저 실행,
	// 키 생성, 로컬 세션 저장 같은 후속 부작용 없이 즉시 에러를 반환해야 한다.
	setUserConfigDir(t)
	wantErr := errors.New("invalid api url")
	service := &LoginService{clientErr: wantErr}

	got, err := service.Login(context.Background(), "desktop")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Login() error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Fatalf("Login() response = %#v, want nil", got)
	}
}

func TestLoginReturnsAlreadyLoggedInBeforeServerRequest(t *testing.T) {
	userConfigDir := setUserConfigDir(t)
	if err := config.SaveGlobalSession(config.GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceID:   20,
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}); err != nil {
		t.Fatalf("SaveGlobalSession() error = %v", err)
	}

	client := &fakeLoginAPI{
		startResp: &authapi.LoginStartResponse{
			LoginSessionID: "session-1",
			AuthURL:        "https://example.com/auth",
			ExpiresIn:      300,
		},
	}
	service := &LoginService{client: client}

	got, err := service.Login(context.Background(), "desktop")
	if !errors.Is(err, ErrAlreadyLoggedIn) {
		t.Fatalf("Login() error = %v, want %v", err, ErrAlreadyLoggedIn)
	}
	if got != nil {
		t.Fatalf("Login() response = %#v, want nil", got)
	}
	if client.started {
		t.Fatal("StartLogin should not be called when global session already exists")
	}
	if _, err := os.Stat(filepath.Join(userConfigDir, "envio", "session.json")); err != nil {
		t.Fatalf("session file should exist: %v", err)
	}
}

func TestStartLogin(t *testing.T) {
	t.Parallel()

	// startLogin은 서버의 로그인 시작 응답을 검증하는 얇은 경계다.
	// API 오류는 그대로 반환하고, 이후 흐름에 필요한 loginSessionId/authUrl 누락은
	// 서비스 레벨에서 명확한 입력 오류로 막아야 한다.
	tests := []struct {
		name      string
		client    *fakeLoginAPI
		wantErr   string
		wantStart bool
	}{
		{
			name: "success",
			client: &fakeLoginAPI{
				startResp: &authapi.LoginStartResponse{
					LoginSessionID: "session-1",
					AuthURL:        "https://example.com/auth",
					ExpiresIn:      300,
				},
			},
			wantStart: true,
		},
		{
			name: "client error",
			client: &fakeLoginAPI{
				startErr: errors.New("start failed"),
			},
			wantErr:   "start failed",
			wantStart: true,
		},
		{
			name: "empty login session id",
			client: &fakeLoginAPI{
				startResp: &authapi.LoginStartResponse{
					AuthURL: "https://example.com/auth",
				},
			},
			wantErr:   "loginSessionId",
			wantStart: true,
		},
		{
			name: "empty auth url",
			client: &fakeLoginAPI{
				startResp: &authapi.LoginStartResponse{
					LoginSessionID: "session-1",
				},
			},
			wantErr:   "authUrl",
			wantStart: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &LoginService{client: tt.client}
			got, err := service.startLogin(context.Background())
			if tt.wantErr != "" {
				// 오류 메시지 전체는 로캘/문구 변경에 취약하므로 핵심 식별자만 확인한다.
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("startLogin() error = %v, want containing %q", err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("startLogin() response = %#v, want nil", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("startLogin() error = %v", err)
			}
			if got.LoginSessionID != "session-1" || got.AuthURL == "" {
				t.Fatalf("startLogin() response = %#v", got)
			}
			// fake 클라이언트가 실제로 호출됐는지 확인해 테스트가 단순 기본값 비교로 통과하지 않게 한다.
			if tt.client.started != tt.wantStart {
				t.Fatalf("StartLogin called = %v, want %v", tt.client.started, tt.wantStart)
			}
		})
	}
}

func TestGetLoginStatus(t *testing.T) {
	t.Parallel()

	// getLoginStatus는 API 응답을 받아 LoginService가 처리할 수 있는 상태값인지 확인한다.
	// 빈 status는 이후 상태 머신에서 의미가 없으므로 여기서 에러로 변환되어야 한다.
	tests := []struct {
		name        string
		client      *fakeLoginAPI
		wantErr     bool
		wantErrText string
	}{
		{
			name: "success",
			client: &fakeLoginAPI{
				statusResp: &authapi.LoginStatusResponse{
					Status:   loginStatusCompleted,
					GithubID: "octocat",
					Email:    "mona@example.com",
				},
			},
		},
		{
			name: "client error",
			client: &fakeLoginAPI{
				statusErr: errors.New("status failed"),
			},
			wantErr:     true,
			wantErrText: "status failed",
		},
		{
			name: "empty status",
			client: &fakeLoginAPI{
				statusResp: &authapi.LoginStatusResponse{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &LoginService{client: tt.client}
			got, err := service.getLoginStatus(context.Background(), "session-1")
			if tt.wantErr {
				// 빈 status 응답은 현재 한국어 메시지를 반환하므로, 그 케이스는 에러 발생 여부만 검증한다.
				if err == nil {
					t.Fatal("getLoginStatus() error = nil, want error")
				}
				if tt.wantErrText != "" && !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("getLoginStatus() error = %v, want containing %q", err, tt.wantErrText)
				}
				if got != nil {
					t.Fatalf("getLoginStatus() response = %#v, want nil", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("getLoginStatus() error = %v", err)
			}
			if got.Status != loginStatusCompleted {
				t.Fatalf("getLoginStatus() response = %#v", got)
			}
			// 로그인 상태 조회는 반드시 startLogin에서 받은 loginSessionID를 그대로 사용해야 한다.
			if tt.client.statusID != "session-1" {
				t.Fatalf("GetLoginStatus loginSessionID = %q", tt.client.statusID)
			}
		})
	}
}

func TestWaitLoginCompleteHandlesTerminalStatuses(t *testing.T) {
	t.Parallel()

	// waitLoginComplete는 서버 상태 문자열을 LoginService의 상태 머신으로 해석한다.
	// COMPLETED는 대소문자와 무관하게 성공해야 하고, EXPIRED/FAILED/알 수 없는 상태는
	// 즉시 에러로 종료되어야 한다.
	tests := []struct {
		name    string
		status  string
		wantErr string
	}{
		{
			name:   "completed",
			status: "completed",
		},
		{
			name:    "expired",
			status:  loginStatusExpired,
			wantErr: "GitHub",
		},
		{
			name:    "failed",
			status:  loginStatusFailed,
			wantErr: "GitHub",
		},
		{
			name:    "unknown",
			status:  "CANCELLED",
			wantErr: "CANCELLED",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeLoginAPI{
				statusResp: &authapi.LoginStatusResponse{
					Status: tt.status,
				},
			}
			service := &LoginService{client: client}

			// PENDING은 ticker를 기다리므로 이 테스트에서는 즉시 종료되는 terminal status만 다룬다.
			got, err := service.waitLoginComplete(context.Background(), "session-1", 1)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("waitLoginComplete() error = %v, want containing %q", err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("waitLoginComplete() response = %#v, want nil", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("waitLoginComplete() error = %v", err)
			}
			if got.Status != tt.status {
				t.Fatalf("waitLoginComplete() response = %#v", got)
			}
			if client.statusID != "session-1" {
				t.Fatalf("GetLoginStatus loginSessionID = %q", client.statusID)
			}
		})
	}
}

func TestRegisterKey(t *testing.T) {
	t.Parallel()

	// registerKey는 공개키 등록 API의 얇은 위임 계층이다.
	// 응답 ID가 그대로 반환되는지와 요청 본문이 변경 없이 전달되는지를 함께 확인한다.
	client := &fakeLoginAPI{
		registerResp: &authapi.RegisterKeyResponse{
			UserID:   10,
			GithubID: "octocat",
			Email:    "mona@example.com",
			DeviceID: 20,
		},
	}
	service := &LoginService{client: client}
	req := authapi.RegisterKeyRequest{
		LoginSessionID: "session-1",
		PublicKey:      "public-key",
		DeviceName:     "desktop",
	}

	got, err := service.registerKey(context.Background(), req)
	if err != nil {
		t.Fatalf("registerKey() error = %v", err)
	}
	if got.UserID != 10 || got.DeviceID != 20 {
		t.Fatalf("registerKey() response = %#v", got)
	}
	if client.registerReq != req {
		t.Fatalf("RegisterKey request = %#v, want %#v", client.registerReq, req)
	}
}

func TestRegisterKeyReturnsClientError(t *testing.T) {
	t.Parallel()

	// 공개키 등록 실패는 호출자가 로그인 실패로 처리할 수 있도록 원래 에러를 보존해야 한다.
	wantErr := errors.New("register failed")
	service := &LoginService{
		client: &fakeLoginAPI{
			registerErr: wantErr,
		},
	}

	got, err := service.registerKey(context.Background(), authapi.RegisterKeyRequest{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("registerKey() error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Fatalf("registerKey() response = %#v, want nil", got)
	}
}

func setUserConfigDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", dir)
		return dir
	case "darwin":
		t.Setenv("HOME", dir)
		return filepath.Join(dir, "Library", "Application Support")
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
		return dir
	}
}
