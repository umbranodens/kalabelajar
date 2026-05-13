package main

import (
	"context"
	"fmt"
	"log"

	"github.com/umbranodens/kalabelajar/internal/config"
	"github.com/umbranodens/kalabelajar/internal/database"
	"github.com/umbranodens/kalabelajar/internal/repositories"
	"github.com/umbranodens/kalabelajar/internal/seed"
	"github.com/umbranodens/kalabelajar/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	seeder := services.NewSeedService(
		repositories.NewRoleRepository(db),
		repositories.NewUserRepository(db),
	)
	admin, err := seeder.SeedRolesAndSuperAdmin(context.Background())
	if err != nil {
		log.Fatalf("seed roles and super admin: %v", err)
	}
	if cfg.App.Env != "production" {
		if err := seed.LocalDevelopmentData(context.Background(), db); err != nil {
			log.Fatalf("seed local development data: %v", err)
		}
	}

	fmt.Printf("Seed selesai untuk %s. Super Admin siap: %s\n", cfg.App.Env, admin.Email)
}
