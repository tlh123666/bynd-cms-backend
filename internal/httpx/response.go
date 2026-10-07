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
