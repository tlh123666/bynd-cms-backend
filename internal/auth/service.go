package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	Subject string `json:"sub"`
	Role    string `json:"role"`
	Type    string `json:"typ"`
	Expires int64  `json:"exp"`
	Issued  int64  `json:"iat"`
}

type Service struct{ secret []byte }

func NewService(secret string) *Service { return &Service{secret: []byte(secret)} }

func (s *Service) Issue(subject, role, tokenType string, lifetime time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{Subject: subject, Role: role, Type: tokenType, Issued: now.Unix(), Expires: now.Add(lifetime).Unix()}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := encode(header) + "." + encode(payload)
	return unsigned + "." + s.signature(unsigned), nil
}

func (s *Service) Verify(token, tokenType string) (Claims, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.signature(unsigned))) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if claims.Subject == "" || claims.Type != tokenType || time.Now().UTC().Unix() >= claims.Expires {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) signature(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(value))
	return encode(mac.Sum(nil))
}

func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func Bearer(value string) (string, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization bearer token is required")
	}
	return parts[1], nil
}
