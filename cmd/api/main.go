package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"secrets-vault/internal/auth"
	"secrets-vault/internal/config"
	"secrets-vault/internal/consumers"
	"secrets-vault/internal/database"
	"secrets-vault/internal/management"
	"secrets-vault/internal/roles"
	"secrets-vault/internal/secrets"
	"secrets-vault/internal/users"
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
	usersRepo := users.NewUsersRepository(conn)

	consumersService := consumers.NewConsumersService(consumersRepo)
	secretsService := secrets.NewSecretsService(secretsRepo)
	usersService := users.NewUsersService(usersRepo)

	tokens, err := auth.NewTokenService(cfg.AppConfig.AuthSecret, time.Duration(cfg.AppConfig.AuthTTL)*time.Hour)
	if err != nil {
		log.Fatal(err)
	}

	seedDefaultAdmin(context.Background(), usersService)

	secretsHandler := secrets.NewSecretsHandler(secretsService, consumersService)
	authHandler := auth.NewAuthHandler(usersService, tokens)
	managementHandler := management.NewManagementHandler(consumersService, roleRepo, secretsService, tokens)

	mux := http.NewServeMux()
	secrets.RegisterRoutes(mux, secretsHandler)
	auth.RegisterRoutes(mux, authHandler)
	management.RegisterRoutes(mux, managementHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
