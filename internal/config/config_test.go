package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("APP_PORT", "")
	t.Setenv("ALLOW_ADMIN_REGISTER", "")
	cfg := LoadConfig()
	assert.Equal(t, "8081", cfg.AppPort)
	assert.Equal(t, "admin", cfg.AdminUsername)
	assert.True(t, cfg.AllowAdminRegister)
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("ALLOW_ADMIN_REGISTER", "false")
	cfg := LoadConfig()
	assert.Equal(t, "9999", cfg.AppPort)
	assert.False(t, cfg.AllowAdminRegister)
}
