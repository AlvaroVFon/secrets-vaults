// Package management
package management

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *ManagementHandler) {
	mux.HandleFunc("GET /management/secrets", handler.FindAllSecretsGroupedByConsumer)
	mux.HandleFunc("POST /management/secrets", handler.CreateSecret)
	mux.HandleFunc("PUT /management/secrets/{id}", handler.UpdateSecret)
	mux.HandleFunc("DELETE /management/secrets/{id}", handler.DeleteSecret)
}
