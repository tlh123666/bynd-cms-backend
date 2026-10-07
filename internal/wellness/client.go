package wellness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type upstreamResponse struct {
	Code uint32          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type UpstreamError struct {
	Status  int
	Message string
}

func (e *UpstreamError) Error() string { return e.Message }

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}}
}

func (c *Client) Get(ctx context.Context, path string, query url.Values, reviewer string) (json.RawMessage, error) {
	endpoint := c.baseURL + path
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, bytes.NewReader(nil))
	if err != nil {
		return nil, fmt.Errorf("create BYND request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Internal-Token", c.token)
	request.Header.Set("X-Reviewer-ID", reviewer)
	response, err := c.http.Do(request)
	if err != nil {
		return nil, &UpstreamError{Status: http.StatusBadGateway, Message: "BYND API is unavailable"}
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, &UpstreamError{Status: http.StatusBadGateway, Message: "invalid BYND API response"}
	}
	var envelope upstreamResponse
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, &UpstreamError{Status: http.StatusBadGateway, Message: fmt.Sprintf("BYND API returned a non-JSON response (HTTP %d)", response.StatusCode)}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := strings.TrimSpace(envelope.Msg)
		if message == "" {
			message = http.StatusText(response.StatusCode)
		}
		return nil, &UpstreamError{Status: response.StatusCode, Message: message}
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return json.RawMessage("null"), nil
	}
	return envelope.Data, nil
}
