package content

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
	baseURL, token string
	http           *http.Client
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
	return &Client{strings.TrimRight(baseURL, "/"), token, &http.Client{Timeout: timeout}}
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body []byte, contentType, reviewer string) (json.RawMessage, error) {
	return c.do(ctx, method, path, query, bytes.NewReader(body), contentType, reviewer)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType, reviewer string) (json.RawMessage, error) {
	endpoint := c.baseURL + path
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Internal-Token", c.token)
	if reviewer != "" {
		req.Header.Set("X-Reviewer-ID", reviewer)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &UpstreamError{http.StatusBadGateway, "BYND API is unavailable"}
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, &UpstreamError{http.StatusBadGateway, "invalid BYND API response"}
	}
	var envelope upstreamResponse
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, &UpstreamError{http.StatusBadGateway, fmt.Sprintf("BYND API returned a non-JSON response (HTTP %d)", resp.StatusCode)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(envelope.Msg)
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return nil, &UpstreamError{resp.StatusCode, message}
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return json.RawMessage("null"), nil
	}
	return envelope.Data, nil
}
