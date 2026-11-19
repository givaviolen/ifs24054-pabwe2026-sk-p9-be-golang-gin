package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config menampung seluruh konfigurasi aplikasi dari environment / .env.
type Config struct {
	AppEnv        string
	AppPort       string
	AppKey        string
	AdminUsername string
	AdminPassword string

	DBHost     string
	DBPort     string
	DBDatabase string
	DBUsername string
	DBPassword string
	DBSSLMode  string

	SearchEngineEndpoint string
	OpenAIEndpoint       string
	OpenAIKey            string
	OpenAIModel          string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load membaca .env (jika ada) lalu environment variable.
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:        getenv("APP_ENV", "development"),
		AppPort:       getenv("APP_PORT", "8081"),
		AppKey:        getenv("APP_KEY", "delcom_blog_secret_key_super_secure_2026_random_string_32chars"),
		AdminUsername: getenv("ADMIN_USERNAME", "admin"),
		AdminPassword: getenv("ADMIN_PASSWORD", "admin"),

		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBDatabase: getenv("DB_DATABASE", "pabwe_p9_blog"),
		DBUsername: getenv("DB_USERNAME", "postgres"),
		DBPassword: getenv("DB_PASSWORD", "postgres"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),

		SearchEngineEndpoint: getenv("SEARCH_ENGINE_ENDPOINT", "https://delcom.org/api/v1/search"),
		OpenAIEndpoint:       getenv("OPENAI_ENDPOINT", "https://delcom.org/api/v1"),
		OpenAIKey:            getenv("OPENAI_KEY", "key"),
		OpenAIModel:          getenv("OPENAI_MODEL", "nama-model"),
	}
}
