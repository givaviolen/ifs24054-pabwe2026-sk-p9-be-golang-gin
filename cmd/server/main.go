package main

import (
	"log"

	"app/internal/config"
	"app/internal/database"
	"app/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("gagal migrasi: %v", err)
	}
	if err := database.SeedAdmin(db, cfg); err != nil {
		log.Fatalf("gagal seeding admin: %v", err)
	}

	r := server.NewRouter(cfg, db, nil)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
