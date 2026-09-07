MIGRATE        := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
MIGRATIONS_DIR := migrations
DATABASE_URL   ?= postgres://postgres:postgres@localhost:5432/app?sslmode=disable

.PHONY: help migrate-up migrate-down migrate-down-all migrate-force migrate-version migrate-create migrate-reset test docker-test docker-test-down

help: ## Muestra los comandos disponibles
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

migrate-up: ## Aplica todas las migraciones pendientes
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down: ## Revierte la última migración aplicada
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-down-all: ## Revierte todas las migraciones
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all

migrate-force: ## Fuerza la version de migración: make migrate-force VERSION=1
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)

migrate-version: ## Muestra la version actual de migraciones
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

migrate-create: ## Crea una nueva migración: make migrate-create NAME=create_roles
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-reset: ## Revierte todas las migraciones y las aplica de nuevo
migrate-reset: migrate-down-all migrate-up

docker-test: ## Levanta la infraestructura de test (BD postgres) para integración
	docker compose -f internal/tests/docker-compose.yml --env-file .env.test up -d postgres --wait
	docker compose -f internal/tests/docker-compose.yml --env-file .env.test run --rm migrate

test: ## Levanta la BD de test y ejecuta todos los tests
	$(MAKE) docker-test
	go test -p 1 ./...

docker-test-down: ## Detiene y elimina la infraestructura de test
	docker compose -f internal/tests/docker-compose.yml --env-file .env.test down