package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientForwardsUserRequestMetadata(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "secret" {
			t.Fatal("missing internal token")
		}
		if r.Header.Get("X-Reviewer-ID") != "Admin" {
			t.Fatal("missing reviewer")
		}
		if r.Header.Get("Accept-Language") != "pt-BR" {
			t.Fatal("missing language")
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatal("missing query")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"msg":"OK","data":{"items":[],"total":0}}`))
	}))
	defer upstream.Close()

	client := NewClient(upstream.URL, "secret", time.Second)
	data, err := client.Do(context.Background(), http.MethodGet, "/api/v1/internal/cms/users", url.Values{"page": {"2"}}, nil, "Admin", "pt-BR")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"items":[],"total":0}` {
		t.Fatalf("unexpected data: %s", data)
	}
}

func TestClientPreservesUserConflict(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":100009,"msg":"email is already in use"}`))
	}))
	defer upstream.Close()

	_, err := NewClient(upstream.URL, "secret", time.Second).Do(context.Background(), http.MethodPost, "/api/v1/internal/cms/users", nil, []byte(`{}`), "Admin", "en")
	upstreamErr, ok := err.(*UpstreamError)
	if !ok || upstreamErr.Status != http.StatusConflict || upstreamErr.Message != "email is already in use" {
		t.Fatalf("unexpected error: %#v", err)
	}
}
