// Package management
package management

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"secrets-vault/internal/consumers"
	"secrets-vault/internal/httpx"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/secrets"

	"github.com/go-playground/validator/v10"
)

const superadminRole = "superadmin"

var ErrInternalServerError = errors.New("internal server error")

type consumersService interface {
	FindByApikey(ctx context.Context, apikey string) (*consumers.Consumer, error)
}

type rolesService interface {
	FindByID(ctx context.Context, id string) (*roles.Role, error)
}

type secretsService interface {
	Create(ctx context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error)
	FindAll(ctx context.Context) ([]secrets.Secret, error)
	Update(ctx context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error)
	DeleteByID(ctx context.Context, id string) error
}

type ManagementHandler struct {
	consumersService consumersService
	rolesService     rolesService
	secretsService   secretsService
	validate         *validator.Validate
}

func NewManagementHandler(c consumersService, r rolesService, s secretsService) *ManagementHandler {
	return &ManagementHandler{
		consumersService: c,
		rolesService:     r,
		secretsService:   s,
		validate:         validator.New(),
	}
}

type consumerSecrets struct {
	ConsumerID string           `json:"consumerId"`
	Secrets    []secrets.Secret `json:"secrets"`
}

func (h *ManagementHandler) authorizeSuperadmin(w http.ResponseWriter, r *http.Request) *consumers.Consumer {
	ctx := r.Context()

	apikey := r.Header.Get("x-apikey")
	if apikey == "" {
		httpx.WriteResponse(w, http.StatusUnauthorized, httpx.Response{Status: http.StatusUnauthorized, Message: "Invalid apikey provided"})
		return nil
	}

	consumer, err := h.consumersService.FindByApikey(ctx, apikey)
	if err != nil {
		if errors.Is(err, consumers.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
			return nil
		}
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return nil
	}

	role, err := h.rolesService.FindByID(ctx, consumer.RoleID)
	if err != nil {
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return nil
	}

	if role.Name != superadminRole {
		httpx.WriteResponse(w, http.StatusForbidden, httpx.Response{Status: http.StatusForbidden, Message: "Access denied: superadmin role required"})
		return nil
	}

	return consumer
}

func (h *ManagementHandler) FindAllSecretsGroupedByConsumer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorizeSuperadmin(w, r) == nil {
		return
	}

	allSecrets, err := h.secretsService.FindAll(ctx)
	if err != nil {
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	grouped := make([]consumerSecrets, 0)
	index := make(map[string]int)
	for _, secret := range allSecrets {
		i, ok := index[secret.ConsumerID]
		if !ok {
			i = len(grouped)
			index[secret.ConsumerID] = i
			grouped = append(grouped, consumerSecrets{
				ConsumerID: secret.ConsumerID,
				Secrets:    []secrets.Secret{},
			})
		}
		grouped[i].Secrets = append(grouped[i].Secrets, secret)
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{Status: http.StatusOK, Message: "Secrets found successfully", Data: grouped})
}

func (h *ManagementHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	consumer := h.authorizeSuperadmin(w, r)
	if consumer == nil {
		return
	}

	var createRequest secrets.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	createRequest.ConsumerID = consumer.ID

	if err := h.validate.Struct(createRequest); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	secret, err := h.secretsService.Create(ctx, createRequest)
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

func (h *ManagementHandler) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorizeSuperadmin(w, r) == nil {
		return
	}

	id := r.PathValue("id")

	var updateReq secrets.UpdateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: "Invalid request body"})
		return
	}

	updateReq.ID = id

	if updateReq.Key == nil && updateReq.Value == nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: "No fields to update"})
		return
	}

	if err := h.validate.Struct(updateReq); err != nil {
		httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	secret, err := h.secretsService.Update(ctx, updateReq)
	if err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Secret not found"})
			return
		}
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{
		Status:  http.StatusOK,
		Message: "Secret updated successfully",
		Data:    secret,
	})
}

func (h *ManagementHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorizeSuperadmin(w, r) == nil {
		return
	}

	id := r.PathValue("id")

	if err := h.secretsService.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Secret not found"})
			return
		}
		httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{
		Status:  http.StatusOK,
		Message: "Secret deleted successfully",
	})
}
