package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	authapi "github.com/Team-NEEEE/envio-cli/internal/api/auth"
	"github.com/Team-NEEEE/envio-cli/internal/browser"
	"github.com/Team-NEEEE/envio-cli/internal/config"
	envcrypto "github.com/Team-NEEEE/envio-cli/internal/crypto"
)

// 로그인 풀링 상태와 timeout 설정
const (
	loginStatusPending   = "PENDING"
	loginStatusCompleted = "COMPLETED"
	loginStatusFailed    = "FAILED"
	loginStatusExpired   = "EXPIRED"

	defaultLoginTimeout = 5 * time.Minute
	loginPollInterval   = 2 * time.Second
)

type LoginService struct {
	client    loginAPI
	clientErr error
}

// NewLoginService FIXME: API 요청 방식 수정
func NewLoginService(apiURL string) *LoginService {
	client, err := authapi.NewHTTPClient(apiURL, nil)
	return &LoginService{
		client:    client,
		clientErr: err,
	}
}

type loginAPI interface {
	StartLogin(context.Context) (*authapi.LoginStartResponse, error)
	GetLoginStatus(context.Context, string) (*authapi.LoginStatusResponse, error)
	RegisterKey(context.Context, authapi.RegisterKeyRequest) (*authapi.RegisterKeyResponse, error)
}

// LoginStartResponse login cli 실행시 서버에서 받는 객체
type LoginStartResponse = authapi.LoginStartResponse

// LoginStatusResponse 서버로 풀링 후 받는 결과
type LoginStatusResponse = authapi.LoginStatusResponse

// RegisterKeyRequest 공개키 등록 API 요청이다.
type RegisterKeyRequest = authapi.RegisterKeyRequest

// RegisterKeyResponse 공개키 등록 API 응답이다.
type RegisterKeyResponse = authapi.RegisterKeyResponse

// Login GitHub OAuth 로그인 후 로컬 세션과 키를 저장한다.
// TODO: 전역 폴더 생성
func (s *LoginService) Login(ctx context.Context, deviceName string) (*RegisterKeyResponse, error) {
	if s.clientErr != nil {
		return nil, s.clientErr
	}

	// 내부 기기 이름 확인
	if deviceName == "" {
		host, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("deviceName 확인 실패: %w", err)
		}
		deviceName = host
	}

	// 서버로 로그인 요청
	startResp, err := s.startLogin(ctx)
	if err != nil {
		return nil, err
	}

	if err := browser.Open(startResp.AuthURL); err != nil {
		return nil, fmt.Errorf("브라우저 열기 실패: %w", err)
	}

	if _, err := s.waitLoginComplete(ctx, startResp.LoginSessionID, startResp.ExpiresIn); err != nil {
		return nil, err
	}

	// 공개키, 비밀키 생성
	privatePEM, publicPEM, err := envcrypto.GenerateRSAKeyPairPEM()
	if err != nil {
		return nil, err
	}

	// 비밀키 저장
	if err := envcrypto.SavePrivateKey(privatePEM); err != nil {
		return nil, fmt.Errorf("private key 저장 실패: %w", err)
	}

	// 서버에 공개키와 device 저장
	resp, err := s.registerKey(ctx, RegisterKeyRequest{
		LoginSessionID: startResp.LoginSessionID,
		PublicKey:      publicPEM,
		DeviceName:     deviceName,
	})
	if err != nil {
		return nil, err
	}

	// TODO: 공개키 저장 위치
	if err := config.SaveLocalSession(config.Session{
		UserID:     resp.UserID,
		GithubID:   resp.GithubID,
		DeviceID:   resp.DeviceID,
		DeviceName: deviceName,
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

// startLogin은 서버에 CLI 로그인 시작을 요청하고,
// 브라우저에서 열 GitHub OAuth URL과 로그인 세션 ID를 받아온다.
func (s *LoginService) startLogin(ctx context.Context) (*LoginStartResponse, error) {
	// FIXME: API 요청 방식 수정
	out, err := s.client.StartLogin(ctx)
	if err != nil {
		return nil, err
	}

	if out.LoginSessionID == "" {
		return nil, errors.New("loginSessionId 응답이 비어 있습니다")
	}

	if out.AuthURL == "" {
		return nil, errors.New("authUrl 응답이 비어 있습니다")
	}

	return out, nil
}

func (s *LoginService) waitLoginComplete(
	ctx context.Context,
	loginSessionID string,
	expiresIn int,
) (*LoginStatusResponse, error) {
	timeout := time.Duration(expiresIn) * time.Second
	if timeout <= 0 {
		timeout = defaultLoginTimeout
	}

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(loginPollInterval)
	defer ticker.Stop()

	for {
		statusResp, err := s.getLoginStatus(pollCtx, loginSessionID)
		if err != nil {
			return nil, err
		}

		status := strings.ToUpper(statusResp.Status)

		if status == loginStatusCompleted {
			return statusResp, nil
		}

		if status == loginStatusExpired {
			return nil, errors.New("GitHub 로그인 세션이 만료되었습니다")
		}

		if status == loginStatusFailed {
			return nil, errors.New("GitHub 로그인에 실패했습니다")
		}

		if status != loginStatusPending {
			return nil, fmt.Errorf("알 수 없는 로그인 상태입니다: %s", statusResp.Status)
		}

		select {
		case <-pollCtx.Done():
			return nil, fmt.Errorf("GitHub 로그인 대기 시간 초과: %w", pollCtx.Err())
		case <-ticker.C:
		}
	}
}

// login 상태 풀링
func (s *LoginService) getLoginStatus(ctx context.Context, loginSessionID string) (*LoginStatusResponse, error) {
	// FIXME: API 요청 방식 수정
	out, err := s.client.GetLoginStatus(ctx, loginSessionID)
	if err != nil {
		return nil, err
	}

	if out.Status == "" {
		return nil, errors.New("로그인 상태 응답이 비어 있습니다")
	}

	return out, nil
}

// registerKey는 완료된 loginSessionId를 기반으로,
// CLI 로컬에서 생성한 공개키와 기기 이름을 서버에 등록한다.
func (s *LoginService) registerKey(ctx context.Context, req RegisterKeyRequest) (*RegisterKeyResponse, error) {
	// FIXME: API 요청 방식 수정
	out, err := s.client.RegisterKey(ctx, req)
	if err != nil {
		return nil, err
	}

	return out, nil
}
