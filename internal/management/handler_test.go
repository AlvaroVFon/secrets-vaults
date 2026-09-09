// Package management
package management

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
	"secrets-vault/internal/roles"
	"secrets-vault/internal/secrets"

	"github.com/google/uuid"
)

type mockConsumersService struct {
	findAllFunc      func(ctx context.Context) ([]consumers.Consumer, error)
	findByApikeyFunc func(ctx context.Context, apikey string) (*consumers.Consumer, error)
	findByIDFunc     func(ctx context.Context, id string) (*consumers.Consumer, error)
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

type mockRolesService struct {
	findByIDFunc func(ctx context.Context, id string) (*roles.Role, error)
}

func (m *mockRolesService) FindByID(ctx context.Context, id string) (*roles.Role, error) {
	return m.findByIDFunc(ctx, id)
}

type mockSecretsService struct {
	createFunc     func(ctx context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error)
	findAllFunc    func(ctx context.Context) ([]secrets.Secret, error)
	updateFunc     func(ctx context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error)
	deleteByIDFunc func(ctx context.Context, id string) error
}

func (m *mockSecretsService) Create(ctx context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error) {
	return m.createFunc(ctx, req)
}

func (m *mockSecretsService) FindAll(ctx context.Context) ([]secrets.Secret, error) {
	return m.findAllFunc(ctx)
}

func (m *mockSecretsService) Update(ctx context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error) {
	return m.updateFunc(ctx, req)
}

func (m *mockSecretsService) DeleteByID(ctx context.Context, id string) error {
	return m.deleteByIDFunc(ctx, id)
}

func testSuperadminConsumer() *consumers.Consumer {
	return &consumers.Consumer{
		ID:     uuid.New().String(),
		Name:   "superadmin-1",
		Apikey: "superadmin-key",
		RoleID: uuid.New().String(),
		Active: true,
	}
}

func testSuperadminRole() *roles.Role {
	return &roles.Role{ID: uuid.New().String(), Name: "superadmin"}
}

func testAdminRole() *roles.Role {
	return &roles.Role{ID: uuid.New().String(), Name: "admin"}
}

func setupManagementRouter(c consumersService, r rolesService, s secretsService) http.Handler {
	handler := NewManagementHandler(c, r, s)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)
	return mux
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

func TestManagement_NoApikey_Unauthorized(t *testing.T) {
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
	}

	for _, route := range routes {
		rec := doRequest(t, router.ServeHTTP, route.method, route.path, route.body, "")

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("(%s %s) expected status %d, got %d", route.method, route.path, http.StatusUnauthorized, rec.Code)
		}

		res := decodeResponse(t, rec)
		if res.Message != "Invalid apikey provided" {
			t.Errorf("(%s %s) expected message %q, got %q", route.method, route.path, "Invalid apikey provided", res.Message)
		}
	}
}

func TestManagement_ConsumerNotFound(t *testing.T) {
	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
		&mockRolesService{},
		&mockSecretsService{},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/management/secrets", "", "some-key")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumer not found" {
		t.Errorf("expected message %q, got %q", "Consumer not found", res.Message)
	}
}

func TestManagement_NonSuperadmin_Forbidden(t *testing.T) {
	consumer := testSuperadminConsumer()
	consumer.Apikey = "admin-key"

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testAdminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/management/consumers"},
		{http.MethodGet, "/management/secrets"},
		{http.MethodPost, "/management/secrets"},
		{http.MethodPut, "/management/secrets/" + uuid.New().String()},
		{http.MethodDelete, "/management/secrets/" + uuid.New().String()},
	}

	for _, route := range routes {
		rec := doRequest(t, router.ServeHTTP, route.method, route.path, "{}", consumer.Apikey)

		if rec.Code != http.StatusForbidden {
			t.Errorf("(%s %s) expected status %d, got %d", route.method, route.path, http.StatusForbidden, rec.Code)
		}

		res := decodeResponse(t, rec)
		if res.Message != "Access denied: superadmin role required" {
			t.Errorf("(%s %s) expected message %q, got %q", route.method, route.path, "Access denied: superadmin role required", res.Message)
		}
	}
}

func TestManagement_FindAllSecretsGroupedByConsumer_Success(t *testing.T) {
	consumer := testSuperadminConsumer()
	c1 := uuid.New().String()
	c2 := uuid.New().String()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
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

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/management/secrets", "", consumer.Apikey)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secrets found successfully" {
		t.Errorf("expected message %q, got %q", "Secrets found successfully", res.Message)
	}

	var grouped []consumerSecrets
	raw, err := json.Marshal(res.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if err := json.Unmarshal(raw, &grouped); err != nil {
		t.Fatalf("unmarshal grouped: %v", err)
	}

	if len(grouped) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(grouped))
	}
	if grouped[0].ConsumerID != c1 || len(grouped[0].Secrets) != 2 {
		t.Errorf("expected consumer %s with 2 secrets, got %+v", c1, grouped[0])
	}
	if grouped[1].ConsumerID != c2 || len(grouped[1].Secrets) != 1 {
		t.Errorf("expected consumer %s with 1 secret, got %+v", c2, grouped[1])
	}
	if grouped[1].Secrets[0].Key != "k3" {
		t.Errorf("expected key k3, got %q", grouped[1].Secrets[0].Key)
	}
}

func TestManagement_FindAllSecretsGroupedByConsumer_NoSecrets(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			findAllFunc: func(_ context.Context) ([]secrets.Secret, error) {
				return []secrets.Secret{}, nil
			},
		},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/management/secrets", "", consumer.Apikey)

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

	if grouped == nil || len(grouped) != 0 {
		t.Errorf("expected empty groups, got %+v", grouped)
	}
}

func TestManagement_CreateSecret_ValidationError(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	tests := []struct {
		name string
		body string
	}{
		{name: "missing key", body: `{"value":"v","consumerId":"` + uuid.New().String() + `"}`},
		{name: "missing value", body: `{"key":"k","consumerId":"` + uuid.New().String() + `"}`},
		{name: "empty key", body: `{"key":"","value":"v","consumerId":"` + uuid.New().String() + `"}`},
		{name: "empty value", body: `{"key":"k","value":"","consumerId":"` + uuid.New().String() + `"}`},
		{name: "missing consumerId", body: `{"key":"k","value":"v"}`},
		{name: "invalid consumerId", body: `{"key":"k","value":"v","consumerId":"not-a-uuid"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, router.ServeHTTP, http.MethodPost, "/management/secrets", tt.body, consumer.Apikey)

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
	consumer := testSuperadminConsumer()
	target := &consumers.Consumer{
		ID:     uuid.New().String(),
		Name:   "app-1",
		Apikey: "app-1-key",
		RoleID: uuid.New().String(),
		Active: true,
	}

	expected := &secrets.Secret{
		ID:         uuid.New().String(),
		Key:        "db.password",
		Value:      "s3cret",
		ConsumerID: target.ID,
	}

	var received secrets.CreateSecretRequest
	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
			findByIDFunc: func(_ context.Context, id string) (*consumers.Consumer, error) {
				if id != target.ID {
					t.Errorf("expected FindByID %q, got %q", target.ID, id)
				}
				return target, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			createFunc: func(_ context.Context, req secrets.CreateSecretRequest) (*secrets.Secret, error) {
				received = req
				return expected, nil
			},
		},
	)

	body := `{"key":"db.password","value":"s3cret","consumerId":"` + target.ID + `"}`
	rec := doRequest(t, router.ServeHTTP, http.MethodPost, "/management/secrets", body, consumer.Apikey)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if received.ConsumerID != target.ID {
		t.Errorf("expected ConsumerID %q, got %q", target.ID, received.ConsumerID)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secret created successfully" {
		t.Errorf("expected message %q, got %q", "Secret created successfully", res.Message)
	}
}

func TestManagement_CreateSecret_ConsumerNotFound(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
			findByIDFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return nil, consumers.ErrorNotFound
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	body := `{"key":"k","value":"v","consumerId":"` + uuid.New().String() + `"}`
	rec := doRequest(t, router.ServeHTTP, http.MethodPost, "/management/secrets", body, consumer.Apikey)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumer not found" {
		t.Errorf("expected message %q, got %q", "Consumer not found", res.Message)
	}
}

func TestManagement_FindAllConsumers_Success(t *testing.T) {
	consumer := testSuperadminConsumer()

	expected := []consumers.Consumer{
		{ID: uuid.New().String(), Name: "alpha", Apikey: "alpha-key", RoleID: uuid.New().String(), Active: true},
		{ID: uuid.New().String(), Name: "beta", Apikey: "beta-key", RoleID: uuid.New().String(), Active: false},
	}

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
			findAllFunc: func(_ context.Context) ([]consumers.Consumer, error) {
				return expected, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/management/consumers", "", consumer.Apikey)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Consumers found successfully" {
		t.Errorf("expected message %q, got %q", "Consumers found successfully", res.Message)
	}

	var got []consumers.Consumer
	raw, err := json.Marshal(res.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal consumers: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 consumers, got %d", len(got))
	}
	if got[0].Name != "alpha" || got[1].Name != "beta" {
		t.Errorf("expected consumers alpha and beta, got %+v", got)
	}
}

func TestManagement_UpdateSecret_NoFieldsToUpdate(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/management/secrets/"+uuid.New().String(), `{}`, consumer.Apikey)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "No fields to update" {
		t.Errorf("expected message %q, got %q", "No fields to update", res.Message)
	}
}

func TestManagement_UpdateSecret_InvalidID_ValidationError(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/management/secrets/not-a-uuid", `{"key":"k"}`, consumer.Apikey)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	res := decodeResponse(t, rec)
	if !strings.Contains(res.Message, "uuid") {
		t.Errorf("expected message to reference uuid validation, got %q", res.Message)
	}
}

func TestManagement_UpdateSecret_ValidationError(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/management/secrets/invalid-id", `{"key":null,"value":null}`, consumer.Apikey)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message == "" {
		t.Error("expected non-empty validation error message")
	}
}

func TestManagement_UpdateSecret_Success(t *testing.T) {
	consumer := testSuperadminConsumer()
	secretID := uuid.New().String()
	newValue := "new-secret"

	expected := &secrets.Secret{
		ID:         secretID,
		Key:        "db.password",
		Value:      newValue,
		ConsumerID: consumer.ID,
	}

	var received secrets.UpdateSecretRequest
	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			updateFunc: func(_ context.Context, req secrets.UpdateSecretRequest) (*secrets.Secret, error) {
				received = req
				return expected, nil
			},
		},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/management/secrets/"+secretID, `{"value":"new-secret"}`, consumer.Apikey)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if received.ID != secretID {
		t.Errorf("expected ID %q, got %q", secretID, received.ID)
	}
	if received.Key != nil {
		t.Errorf("expected Key nil, got %q", *received.Key)
	}
	if received.Value == nil || *received.Value != newValue {
		t.Errorf("expected Value %q, got %+v", newValue, received.Value)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secret updated successfully" {
		t.Errorf("expected message %q, got %q", "Secret updated successfully", res.Message)
	}
}

func TestManagement_UpdateSecret_NotFound(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			updateFunc: func(_ context.Context, _ secrets.UpdateSecretRequest) (*secrets.Secret, error) {
				return nil, secrets.ErrNotFound
			},
		},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/management/secrets/"+uuid.New().String(), `{"key":"k"}`, consumer.Apikey)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secret not found" {
		t.Errorf("expected message %q, got %q", "Secret not found", res.Message)
	}
}

func TestManagement_UpdateSecret_EmptyField_BadRequest(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			updateFunc: func(_ context.Context, _ secrets.UpdateSecretRequest) (*secrets.Secret, error) {
				return nil, secrets.ErrEmptyArgument
			},
		},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodPut, "/management/secrets/"+uuid.New().String(), `{"value":"v"}`, consumer.Apikey)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestManagement_DeleteSecret_NotFound(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			deleteByIDFunc: func(_ context.Context, _ string) error {
				return secrets.ErrNotFound
			},
		},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodDelete, "/management/secrets/"+uuid.New().String(), "", consumer.Apikey)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secret not found" {
		t.Errorf("expected message %q, got %q", "Secret not found", res.Message)
	}
}

func TestManagement_RoleServiceError_InternalServerError(t *testing.T) {
	consumer := testSuperadminConsumer()

	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return nil, errors.New("db down")
			},
		},
		&mockSecretsService{},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodGet, "/management/secrets", "", consumer.Apikey)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestManagement_DeleteSecret_Success(t *testing.T) {
	consumer := testSuperadminConsumer()
	secretID := uuid.New().String()

	var gotID string
	router := setupManagementRouter(
		&mockConsumersService{
			findByApikeyFunc: func(_ context.Context, _ string) (*consumers.Consumer, error) {
				return consumer, nil
			},
		},
		&mockRolesService{
			findByIDFunc: func(_ context.Context, _ string) (*roles.Role, error) {
				return testSuperadminRole(), nil
			},
		},
		&mockSecretsService{
			deleteByIDFunc: func(_ context.Context, id string) error {
				gotID = id
				return nil
			},
		},
	)

	rec := doRequest(t, router.ServeHTTP, http.MethodDelete, "/management/secrets/"+secretID, "", consumer.Apikey)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if gotID != secretID {
		t.Errorf("expected id %q, got %q", secretID, gotID)
	}

	res := decodeResponse(t, rec)
	if res.Message != "Secret deleted successfully" {
		t.Errorf("expected message %q, got %q", "Secret deleted successfully", res.Message)
	}
}
