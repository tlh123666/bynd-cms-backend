package wellness

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientForwardsQueryAndReviewer(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/internal/cms/sleep" || r.URL.Query().Get("page") != "2" {
			t.Fatalf("unexpected upstream request: %s", r.URL.String())
		}
		if r.Header.Get("X-Internal-Token") != "secret" || r.Header.Get("X-Reviewer-ID") != "Admin" {
			t.Fatal("missing internal identity")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200000,"msg":"OK","data":{"items":[],"total":0}}`))
	}))
	defer upstream.Close()

	data, err := NewClient(upstream.URL, "secret", time.Second).Get(context.Background(), "/api/v1/internal/cms/sleep", url.Values{"page": {"2"}}, "Admin")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"items":[],"total":0}` {
		t.Fatalf("unexpected data: %s", data)
	}
}
