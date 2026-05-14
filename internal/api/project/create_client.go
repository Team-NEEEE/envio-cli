package projectapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

const (
	createProjectPath      = "/api/v1/cli/create"
	projectJSONContentType = "application/json; charset=UTF-8"
	wrappedKeysPath        = "/api/projects/%d/wrapped-keys"
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

func (c *Client) CreateProject(
	ctx context.Context,
	request CreateProjectRequest,
) (*CreateProjectResponse, error) {
	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   createProjectPath,
		Header: projectJSONHeader(""),
		Body:   request,
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[CreateProjectResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *Client) SaveWrappedKeys(
	ctx context.Context,
	projectID int64,
	request SaveWrappedKeysRequest,
	authorization string,
) (*SaveWrappedKeysResponse, error) {
	if projectID <= 0 {
		return nil, fmt.Errorf("projectId must be positive")
	}

	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPut,
		Path:   fmt.Sprintf(wrappedKeysPath, projectID),
		Header: projectJSONHeader(authorization),
		Body:   request,
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[SaveWrappedKeysResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func projectJSONHeader(authorization string) http.Header {
	header := http.Header{}
	header.Set("Content-Type", projectJSONContentType)
	if authorization != "" {
		header.Set("Authorization", authorization)
	}
	return header
}
