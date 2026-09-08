package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"secrets-vault/internal/consumers"
	"secrets-vault/internal/httpx"

	"github.com/go-playground/validator/v10"
)

var ErrInternalServerError = errors.New("internal server error")

type secretsService interface {
	Create(ctx context.Context, secret CreateSecretRequest) (*Secret, error)
	FindAllByConsumerID(ctx context.Context, consummerID string) ([]Secret, error)
}

type consumersService interface {
	FindByApikey(ctx context.Context, apikey string) (*consumers.Consumer, error)
}

type SecretsHandler struct {
	secretsService   secretsService
	consumersService consumersService
	validate         *validator.Validate
}

func NewSecretsHandler(s secretsService, c consumersService) *SecretsHandler {
	return &SecretsHandler{
		secretsService:   s,
		consumersService: c,
		validate:         validator.New(),
	}
}

func (h *SecretsHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apikey := r.Header.Get("x-apikey")
	if apikey == "" {
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Invalid apikey provided"})
		return
	}

	consumer, err := h.consumersService.FindByApikey(ctx, apikey)
	if err != nil {
		if errors.Is(err, consumers.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
			return
		}
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	var secretRequest CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&secretRequest); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	secretRequest.ConsumerID = consumer.ID

	if err := h.validate.Struct(secretRequest); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	secret, err := h.secretsService.Create(ctx, secretRequest)
	if err != nil {
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	httpx.WriteResponse(w, http.StatusCreated, httpx.Response{
		Status:  http.StatusCreated,
		Message: "Secret created successfully",
		Data:    secret,
	})
}

func (h *SecretsHandler) FindAllByConsumerID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apikey := r.Header.Get("x-apikey")
	if apikey == "" {
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Invalid apikey provided"})
		return
	}

	consumer, err := h.consumersService.FindByApikey(ctx, apikey)
	if err != nil {
		if errors.Is(err, consumers.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
			return
		}
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	secrets, err := h.secretsService.FindAllByConsumerID(ctx, consumer.ID)
	if err != nil {
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{Status: http.StatusOK, Message: "Secrets found successfully", Data: secrets})
}
