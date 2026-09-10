// Package management
package management

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"secrets-vault/internal/auth"
	"secrets-vault/internal/consumers"
	"secrets-vault/internal/httpx"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/secrets"

	"github.com/go-playground/validator/v10"
)

var ErrInternalServerError = errors.New("internal server error")

type consumersService interface {
	Create(ctx context.Context, req consumers.CreateConsumerRequest) (*consumers.Consumer, error)
	FindAll(ctx context.Context) ([]consumers.Consumer, error)
	FindByApikey(ctx context.Context, apikey string) (*consumers.Consumer, error)
	FindByID(ctx context.Context, id string) (*consumers.Consumer, error)
	Update(ctx context.Context, req consumers.UpdateConsumerRequest) (*consumers.Consumer, error)
	DeleteByID(ctx context.Context, id string) error
}

type rolesService interface {
	FindAll(ctx context.Context) ([]roles.Role, error)
	FindByID(ctx context.Context, id string) (*roles.Role, error)
}

type secretsService interface {
	Create(ctx context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error)
	FindAll(ctx context.Context) ([]secrets.Secret, error)
	FindAllFullByConsumerID(ctx context.Context, consumerID string) ([]secrets.Secret, error)
	CountByConsumerID(ctx context.Context, consumerID string) (int, error)
	Update(ctx context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error)
	DeleteByID(ctx context.Context, id string) error
}

type ManagementHandler struct {
	consumersService consumersService
	rolesService     rolesService
	secretsService   secretsService
	tokens           *auth.TokenService
	validate         *validator.Validate
}

func NewManagementHandler(c consumersService, r rolesService, s secretsService, tokens *auth.TokenService) *ManagementHandler {
	return &ManagementHandler{
		consumersService: c,
		rolesService:     r,
		secretsService:   s,
		tokens:           tokens,
		validate:         validator.New(),
	}
}

func (h *ManagementHandler) authorize(w http.ResponseWriter, r *http.Request) *auth.Claims {
	return auth.AuthenticateRequest(w, r, h.tokens)
}

func (h *ManagementHandler) internalServerError(w http.ResponseWriter) {
	httpx.WriteResponse(w, http.StatusInternalServerError, httpx.Response{Status: http.StatusInternalServerError, Message: ErrInternalServerError.Error()})
}

func (h *ManagementHandler) badRequest(w http.ResponseWriter, msg string) {
	httpx.WriteResponse(w, http.StatusBadRequest, httpx.Response{Status: http.StatusBadRequest, Message: msg})
}

type consumerSecrets struct {
	ConsumerID string           `json:"consumerId"`
	Secrets    []secrets.Secret `json:"secrets"`
}

// FindAllSecrets returns all secrets grouped by consumer, or the secrets of a
// single consumer when the consumerId query parameter is provided.
func (h *ManagementHandler) FindAllSecrets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	consumerID := r.URL.Query().Get("consumerId")
	if consumerID != "" {
		if _, err := h.consumersService.FindByID(ctx, consumerID); err != nil {
			if errors.Is(err, consumers.ErrorNotFound) {
				httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
				return
			}
			h.internalServerError(w)
			return
		}

		secretsByConsumer, err := h.secretsService.FindAllFullByConsumerID(ctx, consumerID)
		if err != nil {
			h.internalServerError(w)
			return
		}

		httpx.WriteResponse(w, http.StatusOK, httpx.Response{Status: http.StatusOK, Message: "Secrets found successfully", Data: secretsByConsumer})
		return
	}

	allSecrets, err := h.secretsService.FindAll(ctx)
	if err != nil {
		h.internalServerError(w)
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

func (h *ManagementHandler) FindAllConsumers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	allConsumers, err := h.consumersService.FindAll(ctx)
	if err != nil {
		h.internalServerError(w)
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{Status: http.StatusOK, Message: "Consumers found successfully", Data: allConsumers})
}

func (h *ManagementHandler) CreateConsumer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	var req consumers.CreateConsumerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.badRequest(w, err.Error())
		return
	}

	if _, err := h.rolesService.FindByID(ctx, req.RoleID); err != nil {
		if errors.Is(err, roles.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Role not found"})
			return
		}
		h.internalServerError(w)
		return
	}

	consumer, err := h.consumersService.Create(ctx, req)
	if err != nil {
		h.internalServerError(w)
		return
	}

	httpx.WriteResponse(w, http.StatusCreated, httpx.Response{
		Status:  http.StatusCreated,
		Message: "Consumer created successfully",
		Data:    consumer,
	})
}

func (h *ManagementHandler) UpdateConsumer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	id := r.PathValue("id")

	var req consumers.UpdateConsumerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.badRequest(w, "Invalid request body")
		return
	}

	req.ID = id

	if req.Name == nil && req.Apikey == nil && req.RoleID == nil && req.Active == nil {
		h.badRequest(w, "No fields to update")
		return
	}

	if req.RoleID != nil {
		if _, err := h.rolesService.FindByID(ctx, *req.RoleID); err != nil {
			if errors.Is(err, roles.ErrorNotFound) {
				httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Role not found"})
				return
			}
			h.internalServerError(w)
			return
		}
	}

	consumer, err := h.consumersService.Update(ctx, req)
	if err != nil {
		if errors.Is(err, consumers.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
			return
		}
		if errors.Is(err, consumers.ErrEmptyArgument) || errors.Is(err, consumers.ErrInvalidUUID) {
			h.badRequest(w, err.Error())
			return
		}
		h.internalServerError(w)
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{
		Status:  http.StatusOK,
		Message: "Consumer updated successfully",
		Data:    consumer,
	})
}

func (h *ManagementHandler) DeleteConsumer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	id := r.PathValue("id")

	if _, err := h.consumersService.FindByID(ctx, id); err != nil {
		if errors.Is(err, consumers.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
			return
		}
		h.internalServerError(w)
		return
	}

	count, err := h.secretsService.CountByConsumerID(ctx, id)
	if err != nil {
		h.internalServerError(w)
		return
	}
	if count > 0 {
		httpx.WriteResponse(w, http.StatusConflict, httpx.Response{Status: http.StatusConflict, Message: "Consumer has secrets, delete them first"})
		return
	}

	if err := h.consumersService.DeleteByID(ctx, id); err != nil {
		h.internalServerError(w)
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{
		Status:  http.StatusOK,
		Message: "Consumer deleted successfully",
	})
}

func (h *ManagementHandler) FindAllRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	allRoles, err := h.rolesService.FindAll(ctx)
	if err != nil {
		h.internalServerError(w)
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{Status: http.StatusOK, Message: "Roles found successfully", Data: allRoles})
}

func (h *ManagementHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.authorize(w, r) == nil {
		return
	}

	var createRequest secrets.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		h.badRequest(w, "Invalid request body")
		return
	}

	if err := h.validate.Struct(createRequest); err != nil {
		h.badRequest(w, err.Error())
		return
	}

	if _, err := h.consumersService.FindByID(ctx, createRequest.ConsumerID); err != nil {
		if errors.Is(err, consumers.ErrorNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Consumer not found"})
			return
		}
		h.internalServerError(w)
		return
	}

	secret, err := h.secretsService.Create(ctx, createRequest)
	if err != nil {
		h.internalServerError(w)
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

	if h.authorize(w, r) == nil {
		return
	}

	id := r.PathValue("id")

	var updateReq secrets.UpdateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		h.badRequest(w, "Invalid request body")
		return
	}

	updateReq.ID = id

	if updateReq.Key == nil && updateReq.Value == nil && updateReq.IsSecret == nil {
		h.badRequest(w, "No fields to update")
		return
	}

	if err := h.validate.Struct(updateReq); err != nil {
		h.badRequest(w, err.Error())
		return
	}

	secret, err := h.secretsService.Update(ctx, updateReq)
	if err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Secret not found"})
			return
		}
		if errors.Is(err, secrets.ErrEmptyArgument) || errors.Is(err, secrets.ErrInvalidUUID) {
			h.badRequest(w, err.Error())
			return
		}
		h.internalServerError(w)
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

	if h.authorize(w, r) == nil {
		return
	}

	id := r.PathValue("id")

	if err := h.secretsService.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			httpx.WriteResponse(w, http.StatusNotFound, httpx.Response{Status: http.StatusNotFound, Message: "Secret not found"})
			return
		}
		h.internalServerError(w)
		return
	}

	httpx.WriteResponse(w, http.StatusOK, httpx.Response{
		Status:  http.StatusOK,
		Message: "Secret deleted successfully",
	})
}
