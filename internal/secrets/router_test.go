// Package secrets
package secrets

import (
	"context"
	"net/http"
	"testing"

	"secrets-vault/internal/consumers"
)

func setupRouter(t *testing.T) http.Handler {
	t.Helper()

	handler := NewSecretsHandler(&mockSecretsService{}, &mockConsumersService{})
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return mux
}

func TestRouter_Create_POST(t *testing.T) {
	router := setupRouter(t)

	rec := doRequest(t, router.ServeHTTP, http.MethodPost, "/secrets", `{"key":"db.password","value":"s3cret"}`, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRouter_FindAllByConsumerID_GET(t *testing.T) {
	consumer := testConsumer()

	handler := NewSecretsHandler(
		&mockSecretsService{
			findAllFunc: func(_ context.Context, _ string) ([]Secret, error) {
				return []Secret{}, nil
			},
		},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	rec := doRequest(t, mux.ServeHTTP, http.MethodGet, "/secrets/consumer", "", consumer.Apikey)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRouter_UnknownRoute(t *testing.T) {
	router := setupRouter(t)

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/unknown", "", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRouter_Create_GET_NotAllowed(t *testing.T) {
	router := setupRouter(t)

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/secrets", "", "")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestRouter_Create_WrongMethod_PUT(t *testing.T) {
	router := setupRouter(t)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/secrets", "{}", "")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestRouter_FindAll_POST_NotAllowed(t *testing.T) {
	router := setupRouter(t)

	rec := doRequest(t, router.ServeHTTP, http.MethodPost, "/secrets/consumer", "{}", "")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}
