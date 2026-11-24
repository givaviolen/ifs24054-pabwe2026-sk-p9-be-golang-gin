package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"app/internal/config"
	"app/internal/database"
	"app/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	port := strings.TrimSpace(cfg.AppPort)
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = os.Getenv("APP_PORT")
			if port == "" {
				port = "8081"
			}
		}
	}
	cfg.AppPort = port

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Printf("Peringatan: Gagal terhubung ke database: %v. Menggunakan in-memory database...", err)
		db, _ = database.ConnectTestDB()
	}

	if err := database.EnsureSchema(db, cfg); err != nil {
		log.Printf("Peringatan: Gagal menyiapkan database schema: %v", err)
	}

	r := server.SetupRouter(cfg, db)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server berjalan di port %s (addr: %s)", port, addr)
	if err := http.ListenAndServe(addr, server.RestoreOriginalPath(r)); err != nil {
		log.Fatalf("Server berhenti karena error: %v", err)
	}
}
