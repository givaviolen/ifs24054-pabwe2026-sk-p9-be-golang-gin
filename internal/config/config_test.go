package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_PORT", "")
	cfg := Load()
	assert.Equal(t, "8081", cfg.AppPort)
	assert.Equal(t, "admin", cfg.AdminUsername)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("APP_PORT", "9999")
	assert.Equal(t, "9999", Load().AppPort)
}
