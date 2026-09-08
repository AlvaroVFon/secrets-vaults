package secrets

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *SecretsHandler) {
	mux.HandleFunc("POST /secrets", handler.Create)
	mux.HandleFunc("GET /secrets/consumer", handler.FindAllByConsumerID)
}
