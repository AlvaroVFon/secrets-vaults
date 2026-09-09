package main

import (
	"log"
	"net/http"

	"secrets-vault/internal/config"
	"secrets-vault/internal/consumers"
	"secrets-vault/internal/database"
	"secrets-vault/internal/management"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/secrets"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := database.Connect(&cfg.DBConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	roleRepo := roles.NewRoleRepository(conn)
	consumersRepo := consumers.NewConsumersRepository(conn)
	secretsRepo := secrets.NewSecretsRepository(conn)

	consumersService := consumers.NewConsumersService(consumersRepo)
	secretsService := secrets.NewSecretsService(secretsRepo)

	secretsHandler := secrets.NewSecretsHandler(secretsService, consumersService)
	managementHandler := management.NewManagementHandler(consumersService, roleRepo, secretsService)

	mux := http.NewServeMux()
	secrets.RegisterRoutes(mux, secretsHandler)
	management.RegisterRoutes(mux, managementHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
