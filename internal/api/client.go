package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultContentType = "application/json"
const maxErrorBodySnippet = 512

var ErrUnsuccessfulResponse = errors.New("api response was not successful")

type Client interface {
	Do(ctx context.Context, request Request) (*RawResponse, error)
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Request struct {
	Body   any
	Query  url.Values
	Header http.Header
	Method string
	Path   string
}

type HTTPClient struct {
	baseURL *url.URL
	doer    HTTPDoer
}

func NewHTTPClient(baseURL string, doer HTTPDoer) (*HTTPClient, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse api base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("api base url must include scheme and host")
	}
	if doer == nil {
		doer = http.DefaultClient
	}
	return &HTTPClient{baseURL: parsed, doer: doer}, nil
}

func (c *HTTPClient) Do(ctx context.Context, request Request) (*RawResponse, error) {
	if c == nil {
		return nil, errors.New("api client is nil")
	}

	httpRequest, err := c.newHTTPRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.doer.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("send api request: %w", err)
	}
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("read api response: %w", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		if isHTTPSuccess(httpResponse.StatusCode) {
			return &RawResponse{Success: true}, nil
		}
		return nil, newHTTPResponseError(httpResponse, body, nil)
	}

	var response RawResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, newHTTPResponseError(httpResponse, body, err)
	}
	if response.Error != nil {
		return &response, response.Error
	}
	if !isHTTPSuccess(httpResponse.StatusCode) {
		return &response, newHTTPResponseError(httpResponse, body, nil)
	}
	if !response.Success {
		return &response, ErrUnsuccessfulResponse
	}
	return &response, nil
}

func (c *HTTPClient) newHTTPRequest(ctx context.Context, request Request) (*http.Request, error) {
	method := strings.TrimSpace(request.Method)
	if method == "" {
		method = http.MethodGet
	}

	endpoint, err := c.endpoint(request.Path, request.Query)
	if err != nil {
		return nil, err
	}

	body, encodedAsJSON, err := requestBody(request.Body)
	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create api request: %w", err)
	}
	httpRequest.Header = cloneHeader(request.Header)
	if httpRequest.Header.Get("Accept") == "" {
		httpRequest.Header.Set("Accept", defaultContentType)
	}
	if encodedAsJSON && httpRequest.Header.Get("Content-Type") == "" {
		httpRequest.Header.Set("Content-Type", defaultContentType)
	}
	return httpRequest, nil
}

func (c *HTTPClient) endpoint(requestPath string, query url.Values) (string, error) {
	if strings.TrimSpace(requestPath) == "" {
		return "", errors.New("api request path is required")
	}

	joined, err := url.JoinPath(c.baseURL.String(), requestPath)
	if err != nil {
		return "", fmt.Errorf("join api request path: %w", err)
	}
	parsed, err := url.Parse(joined)
	if err != nil {
		return "", fmt.Errorf("parse api request url: %w", err)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func requestBody(body any) (io.Reader, bool, error) {
	switch typed := body.(type) {
	case nil:
		return nil, false, nil
	case io.Reader:
		return typed, false, nil
	case []byte:
		return bytes.NewReader(typed), false, nil
	case string:
		return strings.NewReader(typed), false, nil
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return nil, false, fmt.Errorf("encode api request body: %w", err)
		}
		return bytes.NewReader(encoded), true, nil
	}
}

func cloneHeader(header http.Header) http.Header {
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func isHTTPSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func newHTTPResponseError(response *http.Response, body []byte, cause error) *HTTPResponseError {
	statusCode := 0
	status := ""
	if response != nil {
		statusCode = response.StatusCode
		status = response.Status
	}
	return &HTTPResponseError{
		StatusCode:  statusCode,
		Status:      status,
		BodySnippet: bodySnippet(body),
		Cause:       cause,
	}
}

type HTTPResponseError struct {
	Cause       error
	Status      string
	BodySnippet string
	StatusCode  int
}

func (e *HTTPResponseError) Error() string {
	if e == nil {
		return ""
	}

	status := e.Status
	if status == "" && e.StatusCode != 0 {
		status = fmt.Sprintf("%d", e.StatusCode)
	}
	if status == "" {
		status = "unknown status"
	}

	parts := []string{fmt.Sprintf("api response %s", status)}
	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}
	if e.BodySnippet != "" {
		parts = append(parts, "body: "+e.BodySnippet)
	}
	return strings.Join(parts, ": ")
}

func (e *HTTPResponseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func bodySnippet(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}
	if len(body) > maxErrorBodySnippet {
		body = body[:maxErrorBodySnippet]
	}
	return strings.Join(strings.Fields(string(body)), " ")
}
