package database

import (
	"testing"

	"app/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestMigrateAndSeed(t *testing.T) {
	db, err := NewTestDB()
	assert.NoError(t, err)
	assert.NoError(t, Migrate(db))

	cfg := &config.Config{AdminUsername: "admin", AdminPassword: "admin"}
	assert.NoError(t, SeedAdmin(db, cfg))
	assert.NoError(t, SeedAdmin(db, cfg)) // kedua kali: tidak duplikat

	var n int64
	db.Table("users").Count(&n)
	assert.Equal(t, int64(1), n)
}

func TestConnectInvalid(t *testing.T) {
	_, err := Connect(&config.Config{DBHost: "127.0.0.1", DBPort: "1", DBSSLMode: "disable"})
	assert.Error(t, err)
}
