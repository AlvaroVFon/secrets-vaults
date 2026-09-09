// Package auth
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"secrets-vault/internal/httpx"
	"secrets-vault/internal/users"

	"github.com/go-playground/validator/v10"
)

const ErrMessageInvalidCredentials = "invalid username or password"

type usersService interface {
	Authenticate(ctx context.Context, username, password string) (*users.User, error)
}

type AuthHandler struct {
	usersService usersService
	tokens       *TokenService
	validate     *validator.Validate
}

func NewAuthHandler(u usersService, tokens *TokenService) *AuthHandler {
	return &AuthHandler{
		usersService: u,
		tokens:       tokens,
		validate:     validator.New(),
	}
}

type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	Token     string `json:"token"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expiresAt"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	if err := h.validate.Struct(req); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	user, err := h.usersService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: "internal server error"})
		return
	}
	if user == nil {
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: ErrMessageInvalidCredentials})
		return
	}

	token, exp, err := h.tokens.Sign(user.ID, user.Username)
	if err != nil {
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: "internal server error"})
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{
		Status:  http.StatusOK,
		Message: "Login successful",
		Data: loginResponse{
			Token:     token,
			Username:  user.Username,
			ExpiresAt: exp.Unix(),
		},
	})
}

// AuthenticateBearer validates the Authorization header and returns the claims
// of the authenticated user. It writes the error response and returns nil on failure.
func (h *AuthHandler) AuthenticateBearer(w http.ResponseWriter, r *http.Request) *Claims {
	return AuthenticateRequest(w, r, h.tokens)
}

// AuthenticateRequest validates the Authorization header against the token
// service. On failure it writes the error response and returns nil.
func AuthenticateRequest(w http.ResponseWriter, r *http.Request, tokens *TokenService) *Claims {
	header := r.Header.Get("Authorization")
	if header == "" {
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Missing authorization header"})
		return nil
	}

	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Invalid authorization header"})
		return nil
	}

	claims, err := tokens.Verify(token)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Token expired"})
			return nil
		}
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Invalid token"})
		return nil
	}

	return claims
}
