package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Msg: "OK", Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Response{Code: status, Msg: message})
}

// UpstreamStatus keeps an authentication failure from the trusted BYND API
// separate from the CMS operator's own session. Returning 401 here would make
// the browser discard an otherwise valid CMS login token.
func UpstreamStatus(status int) int {
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return http.StatusBadGateway
	}
	return status
}
