// Package auth
package auth

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *AuthHandler) {
	mux.HandleFunc("POST /auth/login", handler.Login)
}
