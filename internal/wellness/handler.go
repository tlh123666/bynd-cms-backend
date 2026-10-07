package wellness

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"bynd-cms-backend/internal/auth"
	"bynd-cms-backend/internal/httpx"
)

type Handler struct{ client *Client }

func NewHandler(client *Client) *Handler { return &Handler{client: client} }

func (h *Handler) ListSleep(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/sleep", c.Request.URL.Query())
}
func (h *Handler) GetSleep(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/sleep/"+url.PathEscape(c.Param("id")), nil)
}
func (h *Handler) ListJournals(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/journals", c.Request.URL.Query())
}
func (h *Handler) GetJournal(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/journals/"+url.PathEscape(c.Param("id")), nil)
}
func (h *Handler) ListHeartRate(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/heart-rate", c.Request.URL.Query())
}
func (h *Handler) GetHeartRate(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/heart-rate/"+url.PathEscape(c.Param("id")), nil)
}

func (h *Handler) forward(c *gin.Context, path string, query url.Values) {
	data, err := h.client.Get(c.Request.Context(), path, query, auth.ClaimsFromContext(c).Subject)
	if err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) {
			httpx.Error(c, upstream.Status, upstream.Message)
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "wellness request failed")
		return
	}
	httpx.Success(c, data)
}
