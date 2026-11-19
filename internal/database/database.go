package database

import (
	"fmt"

	"app/internal/config"
	"app/internal/features/auth"
	"app/internal/features/post"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect membuka koneksi PostgreSQL untuk runtime.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBDatabase, cfg.DBPort, cfg.DBSSLMode,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// NewTestDB membuka SQLite in-memory untuk pengujian.
func NewTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1) // satu koneksi agar database memori tidak hilang
	return db, nil
}

// Migrate membuat/menyesuaikan tabel users dan posts.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&auth.User{}, &post.Post{})
}

// SeedAdmin membuat akun admin dari ADMIN_USERNAME / ADMIN_PASSWORD bila belum ada.
func SeedAdmin(db *gorm.DB, cfg *config.Config) error {
	var n int64
	if err := db.Model(&auth.User{}).Where("username = ?", cfg.AdminUsername).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}
	return db.Create(&auth.User{
		Username: cfg.AdminUsername,
		Email:    cfg.AdminUsername + "@delcom.local",
		Password: hash,
		Role:     "admin",
	}).Error
}
