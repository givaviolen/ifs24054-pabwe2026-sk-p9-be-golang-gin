package server

import (
	"fmt"
	"log"
	"net/http"

	"app/internal/config"
	"app/internal/database"
)

// Run menjalankan server HTTP (dipakai cmd/server dan main.go di root).
func Run() error {
	cfg := config.LoadConfig()

	// Pengembangan lokal: buat database otomatis bila belum ada (gagal tidak fatal).
	if cfg.DBConnection != "sqlite" {
		if err := database.EnsureDatabase(cfg); err != nil {
			log.Printf("peringatan: tidak dapat memeriksa/membuat database: %v", err)
		}
	}

	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("gagal terhubung ke database: %w", err)
	}
	if err := database.EnsureSchema(db, cfg); err != nil {
		return fmt.Errorf("gagal menyiapkan skema database: %w", err)
	}

	r := NewRouter(cfg, db, nil)
	addr := ":" + cfg.AppPort
	log.Printf("Server berjalan di http://localhost%s", addr)
	return http.ListenAndServe(addr, RestoreOriginalPath(r))
}
