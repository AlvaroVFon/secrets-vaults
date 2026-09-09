// Package management
package management

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *ManagementHandler) {
	mux.HandleFunc("GET /management/consumers", handler.FindAllConsumers)
	mux.HandleFunc("POST /management/consumers", handler.CreateConsumer)
	mux.HandleFunc("PUT /management/consumers/{id}", handler.UpdateConsumer)
	mux.HandleFunc("DELETE /management/consumers/{id}", handler.DeleteConsumer)
	mux.HandleFunc("GET /management/secrets", handler.FindAllSecrets)
	mux.HandleFunc("POST /management/secrets", handler.CreateSecret)
	mux.HandleFunc("PUT /management/secrets/{id}", handler.UpdateSecret)
	mux.HandleFunc("DELETE /management/secrets/{id}", handler.DeleteSecret)
	mux.HandleFunc("GET /management/roles", handler.FindAllRoles)
}
