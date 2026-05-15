package projectapi

import (
	"context"
	"net/http"

	"github.com/Team-NEEEE/envio-cli/internal/api"
)

const linkProjectPath = "/api/core/projects/link"

func (c *Client) LinkProject(
	ctx context.Context,
	request LinkProjectRequest,
	authorization string,
) (*LinkProjectResponse, error) {
	response, err := c.client.Do(ctx, api.Request{
		Method: http.MethodPost,
		Path:   linkProjectPath,
		Header: projectJSONHeader(authorization),
		Body:   request,
	})
	if err != nil {
		return nil, err
	}

	data, err := api.DecodeData[LinkProjectResponse](response)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
