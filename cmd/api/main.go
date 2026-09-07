package main

import (
	"context"
	"fmt"
	"log"

	"secrets-vault/internal/config"
	"secrets-vault/internal/database"
	"secrets-vault/internal/roles"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := database.Connect(&cfg.DBConfig)
	if err != nil {
		log.Fatal(err)
	}

	repo := roles.NewRoleRepository(conn)

	role, err := roles.NewRole("Admin")
	if err != nil {
		log.Fatal(err)
	}
	if err := repo.Create(ctx, *role); err != nil {
		log.Fatal(err)
	}
	found, err := repo.FindByID(ctx, role.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*found)
}
