package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bynd-cms-backend/internal/httpx"
)

const claimsKey = "cmsClaims"

func Required(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := Bearer(c.GetHeader("Authorization"))
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		claims, err := service.Verify(token, "access")
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, ErrInvalidToken.Error())
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) Claims {
	claims, _ := c.Get(claimsKey)
	value, _ := claims.(Claims)
	return value
}
