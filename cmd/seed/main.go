package main

import (
	"fmt"
	"log"

	"github.com/umbranodens/kalabelajar/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	fmt.Printf("Seed command ready for %s environment. Super Admin seed will be added in the auth/model slice.\n", cfg.App.Env)
}
