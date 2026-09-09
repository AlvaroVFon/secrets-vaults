// Package auth
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"secrets-vault/internal/httpx"
	"secrets-vault/internal/users"

	"github.com/google/uuid"
)

type mockUsersService struct {
	authenticateFunc func(ctx context.Context, username, password string) (*users.User, error)
}

func (m *mockUsersService) Authenticate(ctx context.Context, username, password string) (*users.User, error) {
	return m.authenticateFunc(ctx, username, password)
}

func newTestAuthHandler(t *testing.T, u usersService) *AuthHandler {
	t.Helper()
	svc, err := NewTokenService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}
	return NewAuthHandler(u, svc)
}

func doLoginRequest(handler *AuthHandler, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAuthHandler_Login_Success(t *testing.T) {
	user := &users.User{ID: uuid.New().String(), Username: "admin", PasswordHash: "hash"}
	handler := newTestAuthHandler(t, &mockUsersService{
		authenticateFunc: func(_ context.Context, username, password string) (*users.User, error) {
			if username != "admin" || password != "secret123" {
				t.Errorf("expected admin/secret123, got %s/%s", username, password)
			}
			return user, nil
		},
	})

	rec := doLoginRequest(handler, `{"username":"admin","password":"secret123"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var res httpx.Response
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	raw, err := json.Marshal(res.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var login loginResponse
	if err := json.Unmarshal(raw, &login); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	if login.Token == "" {
		t.Error("expected non-empty token")
	}
	if login.Username != "admin" {
		t.Errorf("expected username %q, got %q", "admin", login.Username)
	}
	if login.ExpiresAt <= time.Now().Unix() {
		t.Error("expected ExpiresAt in the future")
	}

	claims, err := handler.tokens.Verify(login.Token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected UserID %q, got %q", user.ID, claims.UserID)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	handler := newTestAuthHandler(t, &mockUsersService{
		authenticateFunc: func(_ context.Context, _, _ string) (*users.User, error) {
			return nil, nil
		},
	})

	rec := doLoginRequest(handler, `{"username":"admin","password":"wrong"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var res httpx.Response
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.Message != ErrMessageInvalidCredentials {
		t.Errorf("expected message %q, got %q", ErrMessageInvalidCredentials, res.Message)
	}
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	handler := newTestAuthHandler(t, &mockUsersService{})

	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: `{}`},
		{name: "missing password", body: `{"username":"admin"}`},
		{name: "missing username", body: `{"password":"secret123"}`},
		{name: "invalid json", body: `{invalid`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doLoginRequest(handler, tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}
}

func TestAuthHandler_Login_ServiceError(t *testing.T) {
	handler := newTestAuthHandler(t, &mockUsersService{
		authenticateFunc: func(_ context.Context, _, _ string) (*users.User, error) {
			return nil, context.Canceled
		},
	})

	rec := doLoginRequest(handler, `{"username":"admin","password":"secret123"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestAuthHandler_Login_TrimsUsername(t *testing.T) {
	user := &users.User{ID: uuid.New().String(), Username: "admin", PasswordHash: "hash"}
	var gotUsername string
	handler := newTestAuthHandler(t, &mockUsersService{
		authenticateFunc: func(_ context.Context, username, _ string) (*users.User, error) {
			gotUsername = username
			return user, nil
		},
	})

	rec := doLoginRequest(handler, `{"username":"  admin  ","password":"secret123"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if gotUsername != "admin" {
		t.Errorf("expected trimmed username %q, got %q", "admin", gotUsername)
	}
}

func TestAuthHandler_AuthenticateRequest_ValidToken(t *testing.T) {
	handler := newTestAuthHandler(t, &mockUsersService{})
	userID := uuid.New().String()
	token, _, err := handler.tokens.Sign(userID, "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/management/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	claims := handler.AuthenticateBearer(rec, req)

	if claims == nil {
		t.Fatal("expected claims, got nil")
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %q, got %q", userID, claims.UserID)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected no response body, got %q", rec.Body.String())
	}
}

func TestAuthHandler_AuthenticateRequest_MissingHeader(t *testing.T) {
	handler := newTestAuthHandler(t, &mockUsersService{})

	req := httptest.NewRequest(http.MethodGet, "/management/secrets", nil)
	rec := httptest.NewRecorder()

	claims := handler.AuthenticateBearer(rec, req)

	if claims != nil {
		t.Fatalf("expected nil claims, got %+v", claims)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthHandler_AuthenticateRequest_InvalidToken(t *testing.T) {
	handler := newTestAuthHandler(t, &mockUsersService{})

	req := httptest.NewRequest(http.MethodGet, "/management/secrets", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()

	claims := handler.AuthenticateBearer(rec, req)

	if claims != nil {
		t.Fatalf("expected nil claims, got %+v", claims)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
