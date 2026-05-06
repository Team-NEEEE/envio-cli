package authapi

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

const (
	loginStartPath  = "/api/auth/cli/login/start"
	loginStatusPath = "/api/auth/cli/login/status"
	registerKeyPath = "/api/auth/users/me/keys"
)

type Client struct {
	client api.Client
}

func NewClient(client api.Client) *Client {
	return &Client{client: client}
}

func NewHTTPClient(baseURL string, doer api.HTTPDoer) (*Client, error) {
	client, err := api.NewHTTPClient(baseURL, doer)
	if err != nil {
		return nil, err
	}
	return NewClient(client), nil
}

func (c *Client) StartLogin(ctx context.Context) (*LoginStartResponse, error) {
	query := url.Values{}
	query.Set("redirectType", "CLI")

	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodGet,
		Path:   loginStartPath,
		Query:  query,
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

func (c *Client) GetLoginStatus(ctx context.Context, loginSessionID string) (*LoginStatusResponse, error) {
	query := url.Values{}
	query.Set("loginSessionId", loginSessionID)

	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodGet,
		Path:   loginStatusPath,
		Query:  query,
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

func (c *Client) RegisterKey(ctx context.Context, request RegisterKeyRequest) (*RegisterKeyResponse, error) {
	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   registerKeyPath,
		Body:   request,
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[RegisterKeyResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
