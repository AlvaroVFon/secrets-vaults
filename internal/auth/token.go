// Package auth provides token signing/verification and the login endpoint
// for the management panel. Tokens are HMAC-SHA256 signed and stateless.
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

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrEmptySecret  = errors.New("auth secret must not be empty")
)

type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) (*TokenService, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (s *TokenService) Sign(userID, username string) (string, time.Time, error) {
	exp := time.Now().Add(s.ttl)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Exp:      exp.Unix(),
	}

	raw, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	payload := base64.RawURLEncoding.EncodeToString(raw)
	sig := s.sign(payload)

	return payload + "." + sig, exp, nil
}

func (s *TokenService) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	payload, sig := parts[0], parts[1]

	if !hmac.Equal([]byte(sig), []byte(s.sign(payload))) {
		return nil, ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if claims.UserID == "" {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.Exp {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (s *TokenService) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
