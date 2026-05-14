package authapi

import (
	"context"
	"net/http"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

const (
	loginStartPath  = "/api/auth/cli/login/start"
	loginStatusPath = "/api/auth/cli/login/github/callback"
	registerKeyPath = "/api/auth/cli/login/save"
)

// Client 인증 API 클라이언트다.
// 실제 HTTP 요청 생성, baseURL 처리, JSON 인코딩/디코딩 전 단계 등은 internal/api 패키지의 Client에게 위임한다.
type Client struct {
	client api.Client
}

// NewClient authapi.Client를 생성한다.
// 테스트 시 mock api.Client를 넣어서 검증할 수 있다.
func NewClient(client api.Client) *Client {
	return &Client{client: client}
}

// NewHTTPClient HTTP 기반 인증 API 클라이언트를 생성한다.
// doer는 실제 HTTP 요청을 수행하는 객체이며, 보통 *http.Client를 사용한다.
// 테스트 시 fake doer를 넣어 서버 없이도 테스트할 수 있다.
func NewHTTPClient(baseURL string, doer api.HTTPDoer) (*Client, error) {
	client, err := api.NewHTTPClient(baseURL, doer)
	if err != nil {
		return nil, err
	}
	return NewClient(client), nil
}

// StartLogin 로그인 시작 API를 호출한다.
func (c *Client) StartLogin(ctx context.Context) (*LoginStartResponse, error) {

	// 공통 api.Client를 사용한다.
	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   loginStartPath,
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[LoginStartResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetLoginStatus 상태 서버에 풀링 API를 호출한다.
func (c *Client) GetLoginStatus(ctx context.Context, loginSessionID string) (*LoginStatusResponse, error) {
	// 로그인 세션 ID를 body에 담아 상태 API를 호출한다.
	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   loginStatusPath,
		Body: LoginStatusRequest{
			LoginSessionID: loginSessionID,
		},
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[LoginStatusResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// RegisterKey 공개키 저장 API를 호출한다.
func (c *Client) RegisterKey(ctx context.Context, request RegisterKeyRequest) (*RegisterKeyResponse, error) {
	// client Do 메서드를 사용해서 request 요청을 보낸다.
	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   registerKeyPath,
		Body:   request,
	})
	if err != nil {
		return nil, err
	}

	// 응답에서 data 필드만 추출해서 RegisterKeyResponse 타입으로 디코딩한다.
	data, err := api.DecodeData[RegisterKeyResponse](response)

	if err != nil {
		return nil, err
	}
	return &data, nil
}
