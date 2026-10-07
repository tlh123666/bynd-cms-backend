package community

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

func (h *Handler) ListGroups(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/community/groups", listQuery(c, "keyword", "visibility", "status"), nil)
}

func (h *Handler) GetGroup(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/community/groups/"+url.PathEscape(c.Param("groupId")), nil, nil)
}

func (h *Handler) ListChallenges(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/community/challenges", listQuery(c, "keyword", "domain", "status"), nil)
}

func (h *Handler) GetChallenge(c *gin.Context) {
	h.forward(c, http.MethodGet, challengePath(c), nil, nil)
}

func (h *Handler) CreateChallenge(c *gin.Context) {
	h.forwardBody(c, http.MethodPost, "/api/v1/internal/community/challenges")
}

func (h *Handler) UpdateChallenge(c *gin.Context) {
	h.forwardBody(c, http.MethodPatch, challengePath(c))
}

func (h *Handler) SaveChallengeDay(c *gin.Context) {
	path := challengePath(c) + "/days/" + url.PathEscape(c.Param("dayNumber"))
	h.forwardBody(c, http.MethodPut, path)
}

func (h *Handler) PublishChallenge(c *gin.Context) {
	h.forward(c, http.MethodPost, challengePath(c)+"/publish", nil, nil)
}

func (h *Handler) ArchiveChallenge(c *gin.Context) {
	h.forward(c, http.MethodPost, challengePath(c)+"/archive", nil, nil)
}

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
	data, err := h.client.Do(c.Request.Context(), method, path, query, body, reviewer)
	if err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) {
			httpx.Error(c, upstream.Status, upstream.Message)
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "community request failed")
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

func challengePath(c *gin.Context) string {
	return "/api/v1/internal/community/challenges/" + url.PathEscape(c.Param("challengeId"))
}
