package projectapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

const projectHistoryPath = "/api/cli/projects/%d/history"

func (c *Client) History(
	ctx context.Context,
	projectID int64,
	authorization string,
) (*ProjectHistoryResponse, error) {
	if projectID <= 0 {
		return nil, fmt.Errorf("projectId must be positive")
	}

	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf(projectHistoryPath, projectID),
		Header: projectJSONHeader(authorization),
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[ProjectHistoryResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
