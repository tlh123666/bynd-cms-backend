package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"bynd-cms-backend/internal/httpx"
)

type Handler struct {
	service      *Service
	username     string
	password     string
	passwordHash string
}

func NewHandler(service *Service, username, password, passwordHash string) *Handler {
	return &Handler{service: service, username: username, password: password, passwordHash: passwordHash}
}

type loginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "username and password are required")
		return
	}
	usernameOK := subtle.ConstantTimeCompare([]byte(strings.ToLower(request.UserName)), []byte(strings.ToLower(h.username))) == 1
	passwordOK := h.validPassword(request.Password)
	if !usernameOK || !passwordOK {
		httpx.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}
	access, _ := h.service.Issue(h.username, "R_SUPER", "access", 8*time.Hour)
	refresh, _ := h.service.Issue(h.username, "R_SUPER", "refresh", 7*24*time.Hour)
	httpx.Success(c, gin.H{"token": "Bearer " + access, "refreshToken": refresh})
}

func (h *Handler) validPassword(password string) bool {
	if h.passwordHash != "" {
		return bcrypt.CompareHashAndPassword([]byte(h.passwordHash), []byte(password)) == nil
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(h.password)) == 1
}

func (h *Handler) UserInfo(c *gin.Context) {
	claims := ClaimsFromContext(c)
	httpx.Success(c, gin.H{
		"buttons": []string{"community:read", "community:write"},
		"roles":   []string{claims.Role}, "userId": 1, "userName": claims.Subject,
		"email": "", "avatar": "",
	})
}
