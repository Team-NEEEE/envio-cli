package projectapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

const (
	projectPullLatestPath = "/api/core/projects/%d/pull/latest"
	projectPushPath       = "/api/core/projects/%d/push"
)

func (c *Client) PullLatest(
	ctx context.Context,
	projectID int64,
	githubUserID string,
	deviceID int64,
	authorization string,
) (*ProjectPullResponse, error) {
	if projectID <= 0 {
		return nil, fmt.Errorf("projectId must be positive")
	}
	githubUserID = strings.TrimSpace(githubUserID)
	if githubUserID == "" {
		return nil, fmt.Errorf("githubUserId is required")
	}

	query := url.Values{}
	query.Set("githubUserId", githubUserID)
	if deviceID > 0 {
		query.Set("deviceId", fmt.Sprintf("%d", deviceID))
	}

	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   fmt.Sprintf(projectPullLatestPath, projectID),
		Query:  query,
		Header: projectJSONHeader(authorization),
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[ProjectPullResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *Client) Push(
	ctx context.Context,
	projectID int64,
	request ProjectPushRequest,
	authorization string,
) (*ProjectPushResponse, error) {
	if projectID <= 0 {
		return nil, fmt.Errorf("projectId must be positive")
	}

	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   fmt.Sprintf(projectPushPath, projectID),
		Header: projectJSONHeader(authorization),
		Body:   request,
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[ProjectPushResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
