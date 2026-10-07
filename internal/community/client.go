package community

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

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body []byte, reviewer string) (json.RawMessage, error) {
	endpoint := c.baseURL + path
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Internal-Token", c.token)
	if reviewer != "" {
		request.Header.Set("X-Reviewer-ID", reviewer)
	}
	if len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, &UpstreamError{Status: http.StatusBadGateway, Message: "BYND API is unavailable"}
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, &UpstreamError{Status: http.StatusBadGateway, Message: "invalid BYND API response"}
	}
	var envelope upstreamResponse
	if err := json.Unmarshal(payload, &envelope); err != nil {
		if response.StatusCode == http.StatusNotFound {
			return nil, &UpstreamError{
				Status:  http.StatusBadGateway,
				Message: "BYND API community CMS route is unavailable; deploy or restart BYND-backend with the latest internal community routes",
			}
		}
		return nil, &UpstreamError{
			Status:  http.StatusBadGateway,
			Message: fmt.Sprintf("BYND API returned a non-JSON response (HTTP %d)", response.StatusCode),
		}
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
