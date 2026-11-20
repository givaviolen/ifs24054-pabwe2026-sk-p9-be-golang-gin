package database

import (
	"testing"

	"app/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestEnsureSchemaAndSeed(t *testing.T) {
	db, err := NewTestDB()
	assert.NoError(t, err)

	cfg := &config.Config{AdminUsername: "admin", AdminPassword: "admin"}
	assert.NoError(t, EnsureSchema(db, cfg))
	assert.NoError(t, EnsureSchema(db, cfg)) // kedua kali: sudah siap, tidak duplikat

	var users, posts int64
	db.Table("users").Count(&users)
	db.Table("posts").Count(&posts)
	assert.Equal(t, int64(2), users)
	assert.Equal(t, int64(2), posts)

	assert.NoError(t, Seed(db, cfg)) // idempoten
	db.Table("posts").Count(&posts)
	assert.Equal(t, int64(2), posts)
}

func TestConnectSQLite(t *testing.T) {
	db, err := Connect(&config.Config{DBConnection: "sqlite", DBDatabase: "file::memory:"})
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestConnectInvalidPostgres(t *testing.T) {
	_, err := Connect(&config.Config{DBConnection: "pgsql", DBHost: "127.0.0.1", DBPort: "1", DBSSLMode: "disable"})
	assert.Error(t, err)
}

func TestEnsureDatabaseInvalid(t *testing.T) {
	err := EnsureDatabase(&config.Config{DBHost: "127.0.0.1", DBPort: "1", DBSSLMode: "disable"})
	assert.Error(t, err)
}
