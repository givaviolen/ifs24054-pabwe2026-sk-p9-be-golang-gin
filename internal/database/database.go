package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"app/internal/config"
	"app/internal/features/auth"
	"app/internal/features/post"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func gormConfig() *gorm.Config {
	return &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
}

// EnsureDatabase membuat database DB_DATABASE jika belum ada (untuk pengembangan lokal).
func EnsureDatabase(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s connect_timeout=5",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBPort, cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), gormConfig())
	if err != nil {
		return err
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	var n int64
	if err := db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", cfg.DBDatabase).Scan(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	name := strings.ReplaceAll(cfg.DBDatabase, `"`, "")
	return db.Exec(`CREATE DATABASE "` + name + `"`).Error
}

// Connect membuka koneksi database (PostgreSQL, atau SQLite bila DB_CONNECTION=sqlite).
func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if cfg.DBConnection == "sqlite" {
		dialector = sqlite.Open(cfg.DBDatabase)
	} else {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta connect_timeout=10",
			cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBDatabase, cfg.DBPort, cfg.DBSSLMode,
		)
		// PreferSimpleProtocol: kompatibel dengan connection pooler (serverless / PgBouncer).
		dialector = postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
	}
	return gorm.Open(dialector, gormConfig())
}

// NewTestDB membuka SQLite in-memory untuk pengujian.
func NewTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), gormConfig())
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

func schemaReady(db *gorm.DB, cfg *config.Config) bool {
	m := db.Migrator()
	if !m.HasTable(&auth.User{}) || !m.HasTable(&post.Post{}) {
		return false
	}
	var n int64
	db.Model(&auth.User{}).Where("username IN ?", []string{cfg.AdminUsername, "member"}).Count(&n)
	return n >= 2
}

// EnsureSchema menjalankan Migrate + Seed hanya bila belum pernah dilakukan
// (aman untuk serverless: cold start tidak berulang kali migrasi).
func EnsureSchema(db *gorm.DB, cfg *config.Config) error {
	if schemaReady(db, cfg) {
		return nil
	}
	if err := Migrate(db); err != nil {
		return err
	}
	return Seed(db, cfg)
}

func ensureUser(db *gorm.DB, username, email, password, role string) (*auth.User, error) {
	var u auth.User
	err := db.Where("username = ?", username).First(&u).Error
	if err == nil {
		return &u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u = auth.User{Username: username, Email: email, Password: hash, Role: role}
	if err := db.Create(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// Seed membuat akun admin, akun member, dan dua artikel contoh bila tabel posts kosong.
func Seed(db *gorm.DB, cfg *config.Config) error {
	admin, err := ensureUser(db, cfg.AdminUsername, cfg.AdminUsername+"@delcom.org", cfg.AdminPassword, "admin")
	if err != nil {
		return err
	}
	member, err := ensureUser(db, "member", "member@delcom.org", "member", "member")
	if err != nil {
		return err
	}

	var n int64
	if err := db.Model(&post.Post{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	now := time.Now()
	posts := []post.Post{
		{
			UserID:      member.ID,
			Title:       "Pengenalan Bahasa Pemrograman Golang Modern",
			Slug:        "pengenalan-bahasa-pemrograman-golang-modern",
			Category:    "Teknologi",
			Excerpt:     "Panduan memahami Golang dan Gin untuk membangun REST API modern.",
			Content:     "# Belajar Golang\n\nGolang adalah bahasa pemrograman open source dari Google yang unggul dalam performa, efisiensi memori, dan konkurensi lewat goroutine. Artikel ini membahas fitur utama Golang dan framework web Gin.",
			IsPublished: true,
			PublishedAt: &now,
		},
		{
			UserID:      admin.ID,
			Title:       "Penerapan AI dalam Kurikulum Pendidikan 2026",
			Slug:        "penerapan-ai-dalam-kurikulum-pendidikan-2026",
			Category:    "Pendidikan",
			Excerpt:     "Transformasi pendidikan tinggi berbasis teknologi AI terkini.",
			Content:     "# Pendidikan Modern Berbasis AI\n\nPendidikan modern mengintegrasikan kecerdasan buatan untuk membantu proses belajar mengajar yang interaktif dan adaptif.",
			IsPublished: true,
			PublishedAt: &now,
		},
	}
	for i := range posts {
		if err := db.Omit("User").Create(&posts[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
