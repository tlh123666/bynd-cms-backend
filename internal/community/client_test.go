package community

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientForwardsInternalIdentityAndUnwrapsData(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "secret" {
			t.Fatal("missing internal token")
		}
		if r.Header.Get("X-Reviewer-ID") != "Admin" {
			t.Fatal("missing reviewer")
		}
		if r.URL.Query().Get("status") != "draft" {
			t.Fatal("missing query")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"msg":"OK","data":{"items":[]}}`))
	}))
	defer upstream.Close()

	client := NewClient(upstream.URL, "secret", time.Second)
	data, err := client.Do(context.Background(), http.MethodGet, "/api/v1/internal/community/challenges", url.Values{"status": {"draft"}}, nil, "Admin")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"items":[]}` {
		t.Fatalf("unexpected data: %s", data)
	}
}

func TestClientPreservesUpstreamError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":100009,"msg":"challenge is incomplete"}`))
	}))
	defer upstream.Close()

	_, err := NewClient(upstream.URL, "secret", time.Second).Do(context.Background(), http.MethodPost, "/publish", nil, nil, "Admin")
	upstreamErr, ok := err.(*UpstreamError)
	if !ok || upstreamErr.Status != http.StatusConflict || upstreamErr.Message != "challenge is incomplete" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestClientExplainsMissingCommunityCMSRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 page not found"))
	}))
	defer upstream.Close()

	_, err := NewClient(upstream.URL, "secret", time.Second).Do(context.Background(), http.MethodGet, "/api/v1/internal/community/challenges", nil, nil, "Admin")
	upstreamErr, ok := err.(*UpstreamError)
	if !ok || upstreamErr.Status != http.StatusBadGateway || upstreamErr.Message != "BYND API community CMS route is unavailable; deploy or restart BYND-backend with the latest internal community routes" {
		t.Fatalf("unexpected error: %#v", err)
	}
}
