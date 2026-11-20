// Package handler adalah entry point serverless untuk Vercel.
package handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"app/internal/config"
	"app/internal/database"
	"app/internal/server"
)

var (
	mu     sync.Mutex
	router http.Handler
)

// initRouter menyiapkan koneksi DB, skema, dan router sekali per instance.
// Jika gagal, request berikutnya akan mencoba lagi.
func initRouter() (http.Handler, error) {
	mu.Lock()
	defer mu.Unlock()
	if router != nil {
		return router, nil
	}

	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}
	// Serverless: banyak instance paralel, jaga jumlah koneksi per instance tetap kecil.
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(3)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}
	if err := database.EnsureSchema(db, cfg); err != nil {
		return nil, err
	}

	router = server.RestoreOriginalPath(server.NewRouter(cfg, db, nil))
	return router, nil
}

// Handler dipanggil Vercel untuk setiap request.
func Handler(w http.ResponseWriter, r *http.Request) {
	rt, err := initRouter()
	if err != nil {
		log.Printf("init gagal: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status":"error","message":"Server gagal inisialisasi"}`))
		return
	}
	rt.ServeHTTP(w, r)
}
