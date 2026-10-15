package device

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

func (h *Handler) List(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/devices", c.Request.URL.Query())
}

func (h *Handler) Detail(c *gin.Context) {
	h.forward(c, "/api/v1/internal/cms/devices/"+url.PathEscape(c.Param("deviceId")), nil)
}

func (h *Handler) forward(c *gin.Context, path string, query url.Values) {
	data, err := h.client.Get(c.Request.Context(), path, query, auth.ClaimsFromContext(c).Subject)
	if err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) {
			httpx.Error(c, httpx.UpstreamStatus(upstream.Status), upstream.Message)
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "device request failed")
		return
	}
	httpx.Success(c, data)
}
