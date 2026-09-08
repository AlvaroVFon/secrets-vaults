// Package secrets
package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"secrets-vault/internal/consumers"
	"secrets-vault/internal/httpx"

	"github.com/google/uuid"
)

type mockSecretsService struct {
	createFunc  func(ctx context.Context, secret CreateSecretRequest) (*Secret, error)
	findAllFunc func(ctx context.Context, consumerID string) ([]Secret, error)
}

func (m *mockSecretsService) Create(ctx context.Context, secret CreateSecretRequest) (*Secret, error) {
	return m.createFunc(ctx, secret)
}

func (m *mockSecretsService) FindAllByConsumerID(ctx context.Context, consumerID string) ([]Secret, error) {
	return m.findAllFunc(ctx, consumerID)
}

type mockConsumersService struct {
	findByApikeyFunc func(ctx context.Context, apikey string) (*consumers.Consumer, error)
}

func (m *mockConsumersService) FindByApikey(ctx context.Context, apikey string) (*consumers.Consumer, error) {
	return m.findByApikeyFunc(ctx, apikey)
}

func testConsumer() *consumers.Consumer {
	return &consumers.Consumer{
		ID:     uuid.New().String(),
		Name:   "test-consumer",
		Apikey: "test-api-key",
		Active: true,
	}
}

func doRequest(t *testing.T, handler http.HandlerFunc, method, path, body, apikey string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reader)
	if apikey != "" {
		req.Header.Set("x-apikey", apikey)
	}

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder) httpx.Response {
	t.Helper()

	var res httpx.Response
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return res
}

func TestSecretsHandler_Create_NoApikey(t *testing.T) {
	handler := NewSecretsHandler(&mockSecretsService{}, &mockConsumersService{})

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Invalid apikey provided" {
		t.Errorf("expected message %q, got %q", "Invalid apikey provided", res.Message)
	}
}

func TestSecretsHandler_Create_ConsumerNotFound(t *testing.T) {
	handler := NewSecretsHandler(
		&mockSecretsService{},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
	)

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", `{"key":"db.password","value":"s3cret"}`, "valid-key")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumer not found" {
		t.Errorf("expected message %q, got %q", "Consumer not found", res.Message)
	}
}

func TestSecretsHandler_Create_ConsumerServiceError(t *testing.T) {
	handler := NewSecretsHandler(
		&mockSecretsService{},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, errors.New("db down")
			},
		},
	)

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", `{"key":"db.password","value":"s3cret"}`, "valid-key")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSecretsHandler_Create_InvalidBody(t *testing.T) {
	consumer := testConsumer()
	handler := NewSecretsHandler(
		&mockSecretsService{},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", `not-json`, consumer.Apikey)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Invalid request body" {
		t.Errorf("expected message %q, got %q", "Invalid request body", res.Message)
	}
}

func TestSecretsHandler_Create_ValidationError(t *testing.T) {
	consumer := testConsumer()
	handler := NewSecretsHandler(
		&mockSecretsService{},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", `{"key":"","value":"s3cret"}`, consumer.Apikey)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSecretsHandler_Create_ServiceError(t *testing.T) {
	consumer := testConsumer()
	handler := NewSecretsHandler(
		&mockSecretsService{
			createFunc: func(_ context.Context, _ CreateSecretRequest) (*Secret, error) {
				return nil, errors.New("repo failed")
			},
		},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", `{"key":"db.password","value":"s3cret"}`, consumer.Apikey)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSecretsHandler_Create_Success(t *testing.T) {
	consumer := testConsumer()
	expected := &Secret{
		ID:         uuid.New().String(),
		Key:        "db.password",
		Value:      "s3cret",
		ConsumerID: consumer.ID,
	}

	var received CreateSecretRequest
	handler := NewSecretsHandler(
		&mockSecretsService{
			createFunc: func(_ context.Context, secret CreateSecretRequest) (*Secret, error) {
				received = secret
				return expected, nil
			},
		},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	rec := doRequest(t, handler.Create, http.MethodPost, "/secrets", `{"key":"db.password","value":"s3cret"}`, consumer.Apikey)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if received.ConsumerID != consumer.ID {
		t.Errorf("expected ConsumerID %q, got %q", consumer.ID, received.ConsumerID)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secret created successfully" {
		t.Errorf("expected message %q, got %q", "Secret created successfully", res.Message)
	}
}

func TestSecretsHandler_FindAllByConsumerID_NoApikey(t *testing.T) {
	handler := NewSecretsHandler(&mockSecretsService{}, &mockConsumersService{})

	rec := doRequest(t, handler.FindAllByConsumerID, http.MethodGet, "/secrets", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSecretsHandler_FindAllByConsumerID_ConsumerNotFound(t *testing.T) {
	handler := NewSecretsHandler(
		&mockSecretsService{},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
	)

	rec := doRequest(t, handler.FindAllByConsumerID, http.MethodGet, "/secrets", "", "valid-key")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSecretsHandler_FindAllByConsumerID_ServiceError(t *testing.T) {
	consumer := testConsumer()
	handler := NewSecretsHandler(
		&mockSecretsService{
			findAllFunc: func(_ context.Context, _ string) ([]Secret, error) {
				return nil, errors.New("repo failed")
			},
		},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	rec := doRequest(t, handler.FindAllByConsumerID, http.MethodGet, "/secrets", "", consumer.Apikey)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSecretsHandler_FindAllByConsumerID_Success(t *testing.T) {
	consumer := testConsumer()
	expected := []Secret{
		{Key: "db.password", Value: "s3cret"},
		{Key: "db.host", Value: "localhost"},
	}

	handler := NewSecretsHandler(
		&mockSecretsService{
			findAllFunc: func(_ context.Context, gotConsumerID string) ([]Secret, error) {
				if gotConsumerID != consumer.ID {
					t.Errorf("expected consumerID %q, got %q", consumer.ID, gotConsumerID)
				}
				return expected, nil
			},
		},
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
	)

	rec := doRequest(t, handler.FindAllByConsumerID, http.MethodGet, "/secrets", "", consumer.Apikey)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secrets found successfully" {
		t.Errorf("expected message %q, got %q", "Secrets found successfully", res.Message)
	}

	data, err := json.Marshal(res.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if !strings.Contains(string(data), "db.password") {
		t.Errorf("expected data to contain db.password, got %s", data)
	}
}
