package auth

import (
	"errors"
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	service := NewService("01234567890123456789012345678901")
	token, err := service.Issue("Admin", "R_SUPER", "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.Verify(token, "access")
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "Admin" || claims.Role != "R_SUPER" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if _, err := service.Verify(token, "refresh"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("wrong token type should fail, got %v", err)
	}
}

func TestRejectsTamperedToken(t *testing.T) {
	service := NewService("01234567890123456789012345678901")
	token, _ := service.Issue("Admin", "R_SUPER", "access", time.Hour)
	if _, err := service.Verify(token+"x", "access"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("tampered token should fail, got %v", err)
	}
}
