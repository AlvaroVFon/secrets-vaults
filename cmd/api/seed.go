package main

import (
	"context"
	"log"
	"os"

	"secrets-vault/internal/users"
)

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin"
)

// seedDefaultAdmin creates the initial panel user when the users table is empty.
// Credentials come from ADMIN_USERNAME / ADMIN_PASSWORD env vars.
func seedDefaultAdmin(ctx context.Context, usersService *users.UsersService) {
	count, err := usersService.Count(ctx)
	if err != nil {
		log.Fatalf("count users: %v", err)
	}
	if count > 0 {
		return
	}

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = defaultAdminUsername
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = defaultAdminPassword
	}

	if _, err := usersService.Create(ctx, users.CreateUserRequest{Username: username, Password: password}); err != nil {
		log.Fatalf("seed default admin: %v", err)
	}

	log.Printf("created default admin user %q", username)
}
