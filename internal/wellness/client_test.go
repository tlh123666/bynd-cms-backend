package wellness

import (
	"context"
	"errors"
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

func TestClientAcceptsCanonicalCodeAndLanguage(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Language") != "pt-BR" || r.URL.Query().Get("considerTraining") != "false" {
			t.Error("language or explicit false filter was lost")
		}
		_, _ = w.Write([]byte(`{"code":200,"data":{"items":[],"stats":{"averageEnergy":null}}}`))
	}))
	defer upstream.Close()
	_, err := NewClient(upstream.URL, "secret", time.Second).Get(context.Background(), "/api/v1/internal/cms/vitality", url.Values{"considerTraining": {"false"}}, "Admin", "pt-BR")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientRejectsUnsuccessfulOrMalformedEnvelope(t *testing.T) {
	for _, payload := range []string{`{"code":100005,"msg":"database failure"}`, `{"data":[]}`, `not json`} {
		t.Run(payload, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(payload)) }))
			defer upstream.Close()
			_, err := NewClient(upstream.URL, "secret", time.Second).Get(context.Background(), "/api/v1/internal/cms/vitality", nil, "Admin")
			var failure *UpstreamError
			if !errors.As(err, &failure) || failure.Status != http.StatusBadGateway {
				t.Fatalf("expected a gateway error, got %v", err)
			}
		})
	}
}

func TestLocalizedErrors(t *testing.T) {
	for _, language := range []string{"en-US", "zh-CN", "pt-BR"} {
		if localizedError(language, 404) == "" {
			t.Fatal("empty operator message")
		}
	}
	if localizedError("pt-BR, en;q=0.9", 404) == localizedError("en", 404) {
		t.Fatal("Portuguese error was not localized")
	}
}
