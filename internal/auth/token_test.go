// Package auth
package auth

import (
	"errors"
	"testing"
	"time"
)

func newTestTokenService(t *testing.T) *TokenService {
	t.Helper()
	svc, err := NewTokenService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}
	return svc
}

func TestNewTokenService_EmptySecret(t *testing.T) {
	_, err := NewTokenService("", time.Hour)
	if !errors.Is(err, ErrEmptySecret) {
		t.Fatalf("expected ErrEmptySecret, got %v", err)
	}
}

func TestTokenService_SignAndVerify(t *testing.T) {
	svc := newTestTokenService(t)

	token, exp, err := svc.Sign("user-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if exp.IsZero() {
		t.Fatal("expected non-zero expiration")
	}

	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected UserID %q, got %q", "user-1", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("expected Username %q, got %q", "admin", claims.Username)
	}
}

func TestTokenService_Verify_WrongSecret(t *testing.T) {
	svc := newTestTokenService(t)
	other, err := NewTokenService("other-secret", time.Hour)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}

	token, _, err := svc.Sign("user-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := other.Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenService_Verify_MalformedToken(t *testing.T) {
	svc := newTestTokenService(t)

	cases := []string{"", "only-one-part", "a.b.c", "%%%.%%%"}
	for _, tc := range cases {
		if _, err := svc.Verify(tc); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("token %q: expected ErrInvalidToken, got %v", tc, err)
		}
	}
}

func TestTokenService_Verify_TamperedPayload(t *testing.T) {
	svc := newTestTokenService(t)

	token, _, err := svc.Sign("user-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	tampered := "tampered-payload." + token[len(token)-43:]
	if _, err := svc.Verify(tampered); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenService_Verify_Expired(t *testing.T) {
	svc, err := NewTokenService("test-secret", -time.Minute)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}

	token, _, err := svc.Sign("user-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := svc.Verify(token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}

func TestTokenService_Verify_DifferentInstances(t *testing.T) {
	first, err := NewTokenService("same-secret", time.Hour)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}
	second, err := NewTokenService("same-secret", time.Hour)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}

	token, _, err := first.Sign("user-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := second.Verify(token); err != nil {
		t.Fatalf("verify with different instance: %v", err)
	}
}
