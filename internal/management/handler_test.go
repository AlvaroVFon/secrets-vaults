// Package management
package management

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"secrets-vault/internal/auth"
	"secrets-vault/internal/consumers"
	"secrets-vault/internal/httpx"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/secrets"

	"github.com/google/uuid"
)

type mockConsumersService struct {
	createFunc       func(ctx context.Context, req consumers.CreateConsumerRequest) (*consumers.Consumer, error)
	findAllFunc      func(ctx context.Context) ([]consumers.Consumer, error)
	findByApikeyFunc func(ctx context.Context, apikey string) (*consumers.Consumer, error)
	findByIDFunc     func(ctx context.Context, id string) (*consumers.Consumer, error)
	updateFunc       func(ctx context.Context, req consumers.UpdateConsumerRequest) (*consumers.Consumer, error)
	deleteFunc       func(ctx context.Context, id string) error
}

func (m *mockConsumersService) Create(ctx context.Context, req consumers.CreateConsumerRequest) (*consumers.Consumer, error) {
	return m.createFunc(ctx, req)
}

func (m *mockConsumersService) FindAll(ctx context.Context) ([]consumers.Consumer, error) {
	return m.findAllFunc(ctx)
}

func (m *mockConsumersService) FindByApikey(ctx context.Context, apikey string) (*consumers.Consumer, error) {
	return m.findByApikeyFunc(ctx, apikey)
}

func (m *mockConsumersService) FindByID(ctx context.Context, id string) (*consumers.Consumer, error) {
	return m.findByIDFunc(ctx, id)
}

func (m *mockConsumersService) Update(ctx context.Context, req consumers.UpdateConsumerRequest) (*consumers.Consumer, error) {
	return m.updateFunc(ctx, req)
}

func (m *mockConsumersService) DeleteByID(ctx context.Context, id string) error {
	return m.deleteFunc(ctx, id)
}

type mockRolesService struct {
	findAllFunc  func(ctx context.Context) ([]roles.Role, error)
	findByIDFunc func(ctx context.Context, id string) (*roles.Role, error)
}

func (m *mockRolesService) FindAll(ctx context.Context) ([]roles.Role, error) {
	return m.findAllFunc(ctx)
}

func (m *mockRolesService) FindByID(ctx context.Context, id string) (*roles.Role, error) {
	return m.findByIDFunc(ctx, id)
}

type mockSecretsService struct {
	createFunc                func(ctx context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error)
	findAllFunc               func(ctx context.Context) ([]secrets.Secret, error)
	findAllFullByConsumerFunc func(ctx context.Context, consumerID string) ([]secrets.Secret, error)
	countByConsumerFunc       func(ctx context.Context, consumerID string) (int, error)
	updateFunc                func(ctx context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error)
	deleteByIDFunc            func(ctx context.Context, id string) error
}

func (m *mockSecretsService) Create(ctx context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error) {
	return m.createFunc(ctx, req)
}

func (m *mockSecretsService) FindAll(ctx context.Context) ([]secrets.Secret, error) {
	return m.findAllFunc(ctx)
}

func (m *mockSecretsService) FindAllFullByConsumerID(ctx context.Context, consumerID string) ([]secrets.Secret, error) {
	return m.findAllFullByConsumerFunc(ctx, consumerID)
}

func (m *mockSecretsService) CountByConsumerID(ctx context.Context, consumerID string) (int, error) {
	return m.countByConsumerFunc(ctx, consumerID)
}

func (m *mockSecretsService) Update(ctx context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error) {
	return m.updateFunc(ctx, req)
}

func (m *mockSecretsService) DeleteByID(ctx context.Context, id string) error {
	return m.deleteByIDFunc(ctx, id)
}

func testTokenService(t *testing.T) *auth.TokenService {
	t.Helper()
	svc, err := auth.NewTokenService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}
	return svc
}

func mustTokenService() *auth.TokenService {
	svc, err := auth.NewTokenService("test-secret", time.Hour)
	if err != nil {
		panic(err)
	}
	return svc
}

func testToken(t *testing.T) string {
	t.Helper()
	token, _, err := testTokenService(t).Sign(uuid.New().String(), "admin")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func testConsumer() *consumers.Consumer {
	return &consumers.Consumer{
		ID:     uuid.New().String(),
		Name:   "app-1",
		Apikey: "app-1-key",
		RoleID: uuid.New().String(),
		Active: true,
	}
}

func testRole() *roles.Role {
	return &roles.Role{ID: uuid.New().String(), Name: "admin"}
}

func setupManagementRouter(c consumersService, r rolesService, s secretsService) http.Handler {
	handler := NewManagementHandler(c, r, s, mustTokenService())
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)
	return mux
}

func doRequest(t *testing.T, handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

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

func TestManagement_NoToken_Unauthorized(t *testing.T) {
	router := setupManagementRouter(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{})

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/management/consumers", ""},
		{http.MethodGet, "/management/secrets", ""},
		{http.MethodPost, "/management/secrets", `{"key":"k","value":"v"}`},
		{http.MethodPut, "/management/secrets/" + uuid.New().String(), `{"key":"k"}`},
		{http.MethodDelete, "/management/secrets/" + uuid.New().String(), ""},
		{http.MethodPost, "/management/consumers", `{}`},
		{http.MethodPut, "/management/consumers/" + uuid.New().String(), `{}`},
		{http.MethodDelete, "/management/consumers/" + uuid.New().String(), ""},
		{http.MethodGet, "/management/roles", ""},
	}

	for _, route := range routes {
		rec := doRequest(t, router, route.method, route.path, route.body, "")

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("(%s %s) expected status %d, got %d", route.method, route.path, http.StatusUnauthorized, rec.Code)
		}

		res := decodeResponse(t, rec)
		if res.Message != "Missing authorization header" {
			t.Errorf("(%s %s) expected message %q, got %q", route.method, route.path, "Missing authorization header", res.Message)
		}
	}
}

func TestManagement_InvalidToken_Unauthorized(t *testing.T) {
	router := setupManagementRouter(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{})

	rec := doRequest(t, router, http.MethodGet, "/management/secrets", "", "invalid-token")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Invalid token" {
		t.Errorf("expected message %q, got %q", "Invalid token", res.Message)
	}
}

func TestManagement_ExpiredToken_Unauthorized(t *testing.T) {
	svc, err := auth.NewTokenService("test-secret", -time.Minute)
	if err != nil {
		t.Fatalf("new token service: %v", err)
	}
	token, _, err := svc.Sign(uuid.New().String(), "admin")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := NewManagementHandler(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{}, svc)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	rec := doRequest(t, mux, http.MethodGet, "/management/secrets", "", token)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Token expired" {
		t.Errorf("expected message %q, got %q", "Token expired", res.Message)
	}
}

func TestManagement_FindAllSecretsGroupedByConsumer_Success(t *testing.T) {
	token := testToken(t)
	c1 := uuid.New().String()
	c2 := uuid.New().String()

	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{},
		&mockSecretsService{
			findAllFunc: func(_ context.Context) ([]secrets.Secret, error) {
				return []secrets.Secret{
					{ID: uuid.New().String(), Key: "k1", Value: "v1", ConsumerID: c1},
					{ID: uuid.New().String(), Key: "k2", Value: "v2", ConsumerID: c1},
					{ID: uuid.New().String(), Key: "k3", Value: "v3", ConsumerID: c2},
				}, nil
			},
		},
	)

	rec := doRequest(t, router, http.MethodGet, "/management/secrets", "", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var grouped []consumerSecrets
	raw, err := json.Marshal(decodeResponse(t, rec).Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if err := json.Unmarshal(raw, &grouped); err != nil {
		t.Fatalf("unmarshal grouped: %v", err)
	}

	if len(grouped) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(grouped))
	}
}

func TestManagement_FindSecretsByConsumer_Success(t *testing.T) {
	token := testToken(t)
	consumer := testConsumer()

	expected := []secrets.Secret{
		{ID: uuid.New().String(), Key: "k1", Value: "v1", ConsumerID: consumer.ID},
		{ID: uuid.New().String(), Key: "k2", Value: "v2", ConsumerID: consumer.ID},
	}

	router := setupManagementRouter(
		&mockConsumersService{
			findByIDFunc: func(_ context.Context, id string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{},
		&mockSecretsService{
			findAllFullByConsumerFunc: func(_ context.Context, consumerID string) ([]secrets.Secret, error) {
				if consumerID != consumer.ID {
					t.Errorf("expected consumerID %q, got %q", consumer.ID, consumerID)
				}
				return expected, nil
			},
		},
	)

	rec := doRequest(t, router, http.MethodGet, "/management/secrets?consumerId="+consumer.ID, "", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []secrets.Secret
	raw, err := json.Marshal(decodeResponse(t, rec).Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal secrets: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(got))
	}
}

func TestManagement_FindSecretsByConsumer_ConsumerNotFound(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{
			findByIDFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
		&mockRolesService{},
		&mockSecretsService{},
	)

	rec := doRequest(t, router, http.MethodGet, "/management/secrets?consumerId="+uuid.New().String(), "", token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumer not found" {
		t.Errorf("expected message %q, got %q", "Consumer not found", res.Message)
	}
}

func TestManagement_CreateSecret_ValidationError(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{})

	tests := []struct {
		name string
		body string
	}{
		{name: "missing key", body: `{"value":"v","consumerId":"` + uuid.New().String() + `"}`},
		{name: "missing value", body: `{"key":"k","consumerId":"` + uuid.New().String() + `"}`},
		{name: "missing consumerId", body: `{"key":"k","value":"v"}`},
		{name: "invalid consumerId", body: `{"key":"k","value":"v","consumerId":"not-a-uuid"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, router, http.MethodPost, "/management/secrets", tt.body, token)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}

			res := decodeResponse(t, rec)
			if res.Message == "" {
				t.Error("expected non-empty validation error message")
			}
		})
	}
}

func TestManagement_CreateSecret_Success(t *testing.T) {
	token := testToken(t)
	target := testConsumer()

	expected := &secrets.Secret{
		ID:         uuid.New().String(),
		Key:        "db.password",
		Value:      "s3cret",
		ConsumerID: target.ID,
	}

	var received secrets.CreateSecretRequest
	router := setupManagementRouter(
		&mockConsumersService{
			findByIDFunc: func(_ context.Context, id string) (*consumers.Consumer, error) {
				if id != target.ID {
					t.Errorf("expected FindByID %q, got %q", target.ID, id)
				}
				return target, nil
			},
		},
		&mockRolesService{},
		&mockSecretsService{
			createFunc: func(_ context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error) {
				received = req
				return expected, nil
			},
		},
	)

	body := `{"key":"db.password","value":"s3cret","consumerId":"` + target.ID + `"}`
	rec := doRequest(t, router, http.MethodPost, "/management/secrets", body, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if received.ConsumerID != target.ID {
		t.Errorf("expected ConsumerID %q, got %q", target.ID, received.ConsumerID)
	}
}

func TestManagement_FindAllConsumers_Success(t *testing.T) {
	token := testToken(t)

	expected := []consumers.Consumer{
		{ID: uuid.New().String(), Name: "alpha", Apikey: "alpha-key", RoleID: uuid.New().String(), Active: true},
		{ID: uuid.New().String(), Name: "beta", Apikey: "beta-key", RoleID: uuid.New().String(), Active: false},
	}

	router := setupManagementRouter(
		&mockConsumersService{
			findAllFunc: func(_ context.Context) ([]consumers.Consumer, error) {
				return expected, nil
			},
		},
		&mockRolesService{},
		&mockSecretsService{},
	)

	rec := doRequest(t, router, http.MethodGet, "/management/consumers", "", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []consumers.Consumer
	raw, err := json.Marshal(decodeResponse(t, rec).Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal consumers: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 consumers, got %d", len(got))
	}
}

func TestManagement_CreateConsumer_Success(t *testing.T) {
	token := testToken(t)
	role := testRole()
	expected := testConsumer()

	var received consumers.CreateConsumerRequest
	router := setupManagementRouter(
		&mockConsumersService{
			createFunc: func(_ context.Context, req consumers.CreateConsumerRequest) (*consumers.Consumer, error) {
				received = req
				return expected, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, id string) (*roles.Role, error) {
				if id != role.ID {
					t.Errorf("expected role id %q, got %q", role.ID, id)
				}
				return role, nil
			},
		},
		&mockSecretsService{},
	)

	body := `{"name":"app-1","apikey":"app-1-key","roleId":"` + role.ID + `"}`
	rec := doRequest(t, router, http.MethodPost, "/management/consumers", body, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if received.Name != "app-1" || received.Apikey != "app-1-key" || received.RoleID != role.ID {
		t.Errorf("unexpected request: %+v", received)
	}
}

func TestManagement_CreateConsumer_ValidationError(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{})

	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: `{}`},
		{name: "missing apikey", body: `{"name":"app-1","roleId":"` + uuid.New().String() + `"}`},
		{name: "invalid roleId", body: `{"name":"app-1","apikey":"key","roleId":"not-a-uuid"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, router, http.MethodPost, "/management/consumers", tt.body, token)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}
}

func TestManagement_CreateConsumer_RoleNotFound(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return nil, roles.ErrorNotFound
			},
		},
		&mockSecretsService{},
	)

	body := `{"name":"app-1","apikey":"key","roleId":"` + uuid.New().String() + `"}`
	rec := doRequest(t, router, http.MethodPost, "/management/consumers", body, token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Role not found" {
		t.Errorf("expected message %q, got %q", "Role not found", res.Message)
	}
}

func TestManagement_UpdateConsumer_Success(t *testing.T) {
	token := testToken(t)
	consumer := testConsumer()

	var received consumers.UpdateConsumerRequest
	router := setupManagementRouter(
		&mockConsumersService{
			updateFunc: func(_ context.Context, req consumers.UpdateConsumerRequest) (*consumers.Consumer, error) {
				received = req
				return consumer, nil
			},
		},
		&mockRolesService{},
		&mockSecretsService{},
	)

	rec := doRequest(t, router, http.MethodPut, "/management/consumers/"+consumer.ID, `{"active":false}`, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if received.ID != consumer.ID {
		t.Errorf("expected ID %q, got %q", consumer.ID, received.ID)
	}
	if received.Active == nil || *received.Active {
		t.Errorf("expected Active false, got %+v", received.Active)
	}
}

func TestManagement_UpdateConsumer_NoFields(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{})

	rec := doRequest(t, router, http.MethodPut, "/management/consumers/"+uuid.New().String(), `{}`, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "No fields to update" {
		t.Errorf("expected message %q, got %q", "No fields to update", res.Message)
	}
}

func TestManagement_UpdateConsumer_NotFound(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{
			updateFunc: func(_ context.Context, _ consumers.UpdateConsumerRequest) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
		&mockRolesService{},
		&mockSecretsService{},
	)

	rec := doRequest(t, router, http.MethodPut, "/management/consumers/"+uuid.New().String(), `{"name":"x"}`, token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumer not found" {
		t.Errorf("expected message %q, got %q", "Consumer not found", res.Message)
	}
}

func TestManagement_DeleteConsumer_Success(t *testing.T) {
	token := testToken(t)
	consumer := testConsumer()

	var gotID string
	router := setupManagementRouter(
		&mockConsumersService{
			findByIDFunc: func(_ context.Context, id string) (*consumers.Consumer, error) {
				return consumer, nil
			},
			deleteFunc: func(_ context.Context, id string) error {
				gotID = id
				return nil
			},
		},
		&mockRolesService{},
		&mockSecretsService{
			countByConsumerFunc: func(_ context.Context, _ string) (int, error) {
				return 0, nil
			},
		},
	)

	rec := doRequest(t, router, http.MethodDelete, "/management/consumers/"+consumer.ID, "", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if gotID != consumer.ID {
		t.Errorf("expected id %q, got %q", consumer.ID, gotID)
	}
}

func TestManagement_DeleteConsumer_HasSecrets_Conflict(t *testing.T) {
	token := testToken(t)
	consumer := testConsumer()

	var deleted bool
	router := setupManagementRouter(
		&mockConsumersService{
			findByIDFunc: func(_ context.Context, id string) (*consumers.Consumer, error) {
				return consumer, nil
			},
			deleteFunc: func(_ context.Context, _ string) error {
				deleted = true
				return nil
			},
		},
		&mockRolesService{},
		&mockSecretsService{
			countByConsumerFunc: func(_ context.Context, _ string) (int, error) {
				return 3, nil
			},
		},
	)

	rec := doRequest(t, router, http.MethodDelete, "/management/consumers/"+consumer.ID, "", token)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}

	if deleted {
		t.Error("expected consumer not to be deleted")
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumer has secrets, delete them first" {
		t.Errorf("expected message %q, got %q", "Consumer has secrets, delete them first", res.Message)
	}
}

func TestManagement_DeleteConsumer_NotFound(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{
			findByIDFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
		&mockRolesService{},
		&mockSecretsService{},
	)

	rec := doRequest(t, router, http.MethodDelete, "/management/consumers/"+uuid.New().String(), "", token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestManagement_FindAllRoles_Success(t *testing.T) {
	token := testToken(t)

	expected := []roles.Role{
		{ID: uuid.New().String(), Name: "admin"},
		{ID: uuid.New().String(), Name: "superadmin"},
	}

	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{
			findAllFunc: func(_ context.Context) ([]roles.Role, error) {
				return expected, nil
			},
		},
		&mockSecretsService{},
	)

	rec := doRequest(t, router, http.MethodGet, "/management/roles", "", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []roles.Role
	raw, err := json.Marshal(decodeResponse(t, rec).Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal roles: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(got))
	}
}

func TestManagement_UpdateSecret_NoFieldsToUpdate(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(&mockConsumersService{}, &mockRolesService{}, &mockSecretsService{})

	rec := doRequest(t, router, http.MethodPut, "/management/secrets/"+uuid.New().String(), `{}`, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "No fields to update" {
		t.Errorf("expected message %q, got %q", "No fields to update", res.Message)
	}
}

func TestManagement_UpdateSecret_Success(t *testing.T) {
	token := testToken(t)
	secretID := uuid.New().String()
	newValue := "new-secret"

	expected := &secrets.Secret{
		ID:         secretID,
		Key:        "db.password",
		Value:      newValue,
		ConsumerID: uuid.New().String(),
	}

	var received secrets.UpdateSecretRequest
	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{},
		&mockSecretsService{
			updateFunc: func(_ context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error) {
				received = req
				return expected, nil
			},
		},
	)

	rec := doRequest(t, router, http.MethodPut, "/management/secrets/"+secretID, `{"value":"new-secret"}`, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if received.ID != secretID {
		t.Errorf("expected ID %q, got %q", secretID, received.ID)
	}
	if received.Value == nil || *received.Value != newValue {
		t.Errorf("expected Value %q, got %+v", newValue, received.Value)
	}
}

func TestManagement_UpdateSecret_NotFound(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{},
		&mockSecretsService{
			updateFunc: func(_ context.Context, _ secrets.UpdateSecretRequest) (*secrets.Secret, error) {
				return nil, secrets.ErrNotFound
			},
		},
	)

	rec := doRequest(t, router, http.MethodPut, "/management/secrets/"+uuid.New().String(), `{"key":"k"}`, token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestManagement_DeleteSecret_Success(t *testing.T) {
	token := testToken(t)
	secretID := uuid.New().String()

	var gotID string
	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{},
		&mockSecretsService{
			deleteByIDFunc: func(_ context.Context, id string) error {
				gotID = id
				return nil
			},
		},
	)

	rec := doRequest(t, router, http.MethodDelete, "/management/secrets/"+secretID, "", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if gotID != secretID {
		t.Errorf("expected id %q, got %q", secretID, gotID)
	}
}

func TestManagement_DeleteSecret_NotFound(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{},
		&mockSecretsService{
			deleteByIDFunc: func(_ context.Context, _ string) error {
				return secrets.ErrNotFound
			},
		},
	)

	rec := doRequest(t, router, http.MethodDelete, "/management/secrets/"+uuid.New().String(), "", token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestManagement_ServiceError_InternalServerError(t *testing.T) {
	token := testToken(t)

	router := setupManagementRouter(
		&mockConsumersService{},
		&mockRolesService{},
		&mockSecretsService{
			findAllFunc: func(_ context.Context) ([]secrets.Secret, error) {
				return nil, errors.New("db down")
			},
		},
	)

	rec := doRequest(t, router, http.MethodGet, "/management/secrets", "", token)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
