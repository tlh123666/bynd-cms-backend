package content

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"bynd-cms-backend/internal/auth"
	"bynd-cms-backend/internal/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct{ client *Client }

func NewHandler(client *Client) *Handler { return &Handler{client: client} }

func (h *Handler) ListCategories(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/readings/categories", nil, nil, "")
}
func (h *Handler) CreateCategory(c *gin.Context) {
	h.forwardBody(c, http.MethodPost, "/api/v1/internal/readings/categories")
}
func (h *Handler) UpdateCategory(c *gin.Context) {
	h.forwardBody(c, http.MethodPut, "/api/v1/internal/readings/categories/"+url.PathEscape(c.Param("categoryId")))
}
func (h *Handler) DeleteCategory(c *gin.Context) {
	h.forward(c, http.MethodDelete, "/api/v1/internal/readings/categories/"+url.PathEscape(c.Param("categoryId")), nil, nil, "")
}
func (h *Handler) ListPublications(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/readings/publications", query(c, "status", "page", "limit"), nil, "")
}
func (h *Handler) GetPublication(c *gin.Context) {
	h.forward(c, http.MethodGet, publicationPath(c), nil, nil, "")
}
func (h *Handler) CreatePublication(c *gin.Context) {
	h.forwardBody(c, http.MethodPost, "/api/v1/internal/readings/publications")
}
func (h *Handler) UpdatePublication(c *gin.Context) {
	h.forwardBody(c, http.MethodPut, publicationPath(c))
}
func (h *Handler) ArchivePublication(c *gin.Context) {
	h.forward(c, http.MethodDelete, publicationPath(c), nil, nil, "")
}
func (h *Handler) ListAssets(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/readings/assets", query(c, "kind", "page", "limit"), nil, "")
}
func (h *Handler) DeleteAsset(c *gin.Context) {
	h.forward(c, http.MethodDelete, "/api/v1/internal/readings/assets/"+url.PathEscape(c.Param("assetId")), nil, nil, "")
}
func (h *Handler) UploadAsset(c *gin.Context) {
	const maxUpload = 82 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUpload)
	data, err := h.client.do(c.Request.Context(), http.MethodPost, "/api/v1/internal/readings/assets", query(c, "kind"), c.Request.Body, c.GetHeader("Content-Type"), reviewer(c))
	h.respond(c, data, err)
}
func (h *Handler) GetGuide(c *gin.Context) {
	h.forward(c, http.MethodGet, "/api/v1/internal/guidance/body-foundations", nil, nil, "")
}
func (h *Handler) SaveGuide(c *gin.Context) {
	h.forwardBody(c, http.MethodPut, "/api/v1/internal/guidance/body-foundations")
}
func (h *Handler) SaveGuideDay(c *gin.Context) {
	h.forwardBody(c, http.MethodPut, "/api/v1/internal/guidance/body-foundations/days/"+url.PathEscape(c.Param("day")))
}

func (h *Handler) forwardBody(c *gin.Context, method, path string) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
	if err != nil || len(strings.TrimSpace(string(body))) == 0 {
		httpx.Error(c, http.StatusBadRequest, "request body is required")
		return
	}
	h.forward(c, method, path, nil, body, "application/json")
}
func (h *Handler) forward(c *gin.Context, method, path string, q url.Values, body []byte, contentType string) {
	data, err := h.client.Do(c.Request.Context(), method, path, q, body, contentType, reviewer(c))
	h.respond(c, data, err)
}
func (h *Handler) respond(c *gin.Context, data json.RawMessage, err error) {
	if err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) {
			httpx.Error(c, upstream.Status, upstream.Message)
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "content request failed")
		return
	}
	httpx.Success(c, data)
}
func query(c *gin.Context, keys ...string) url.Values {
	q := make(url.Values)
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			q.Set(key, value)
		}
	}
	return q
}
func publicationPath(c *gin.Context) string {
	return "/api/v1/internal/readings/publications/" + url.PathEscape(c.Param("publicationId"))
}
func reviewer(c *gin.Context) string { return auth.ClaimsFromContext(c).Subject }
