package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"bynd-cms-backend/internal/config"
)

func TestLoginAndCommunityProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "internal-secret" {
			t.Fatal("internal token was not forwarded")
		}
		_, _ = w.Write([]byte(`{"code":200000,"msg":"OK","data":{"items":[]}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		HTTPAddr: ":0", BYNDAPIBaseURL: upstream.URL, InternalToken: "internal-secret",
		JWTSecret: "01234567890123456789012345678901", AdminUsername: "Admin", AdminPassword: "123456",
		AllowedOrigins: []string{"http://localhost:3006"}, RequestTimeout: time.Second,
	}
	engine := New(cfg)

	login := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"userName":"Admin","password":"123456"}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(login, request)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d: %s", login.Code, login.Body.String())
	}
	var response struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	groups := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/community/groups?status=active", nil)
	request.Header.Set("Authorization", response.Data.Token)
	engine.ServeHTTP(groups, request)
	if groups.Code != http.StatusOK {
		t.Fatalf("groups status = %d: %s", groups.Code, groups.Body.String())
	}

	sleep := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/wellness/sleep?page=1&pageSize=20", nil)
	request.Header.Set("Authorization", response.Data.Token)
	engine.ServeHTTP(sleep, request)
	if sleep.Code != http.StatusOK {
		t.Fatalf("sleep status = %d: %s", sleep.Code, sleep.Body.String())
	}
}
