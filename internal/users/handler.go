package users

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"bynd-cms-backend/internal/auth"
	"bynd-cms-backend/internal/httpx"
)

type Handler struct{ client *Client }

func NewHandler(client *Client) *Handler { return &Handler{client: client} }

func (h *Handler) List(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/cms/users", listQuery(c, "keyword", "status", "page", "pageSize"), nil)
}

func (h *Handler) Get(c *gin.Context) {
	h.forward(c, http.MethodGet, userPath(c), nil, nil)
}

func (h *Handler) Create(c *gin.Context) {
	h.forwardBody(c, http.MethodPost, "/api/v1/internal/cms/users")
}

func (h *Handler) Update(c *gin.Context) { h.forwardBody(c, http.MethodPut, userPath(c)) }

func (h *Handler) Delete(c *gin.Context) { h.forward(c, http.MethodDelete, userPath(c), nil, nil) }

func (h *Handler) forwardBody(c *gin.Context, method, path string) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil || len(strings.TrimSpace(string(body))) == 0 {
		httpx.Error(c, http.StatusBadRequest, "request body is required")
		return
	}
	h.forward(c, method, path, nil, body)
}

func (h *Handler) forward(c *gin.Context, method, path string, query url.Values, body []byte) {
	reviewer := auth.ClaimsFromContext(c).Subject
	data, err := h.client.Do(c.Request.Context(), method, path, query, body, reviewer, c.GetHeader("Accept-Language"))
	if err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) {
			httpx.Error(c, httpx.UpstreamStatus(upstream.Status), upstream.Message)
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "user request failed")
		return
	}
	httpx.Success(c, data)
}

func listQuery(c *gin.Context, keys ...string) url.Values {
	query := make(url.Values)
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			query.Set(key, value)
		}
	}
	return query
}

func userPath(c *gin.Context) string {
	return "/api/v1/internal/cms/users/" + url.PathEscape(c.Param("userId"))
}
