package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenRoundTrip(t *testing.T) {
	tok, err := GenerateToken("secret", &User{ID: 7, Username: "giva", Role: "member"})
	assert.NoError(t, err)

	claims, err := ParseToken("secret", tok)
	assert.NoError(t, err)
	assert.Equal(t, uint(7), claims.UserID)
	assert.Equal(t, "member", claims.Role)

	_, err = ParseToken("salah", tok)
	assert.Error(t, err)
	_, err = ParseToken("secret", "bukan-token")
	assert.Error(t, err)
}

func TestHashPassword(t *testing.T) {
	h, err := HashPassword("rahasia")
	assert.NoError(t, err)
	assert.NotEqual(t, "rahasia", h)
}

func TestValidEmail(t *testing.T) {
	assert.True(t, validEmail("a@uji.local"))
	assert.False(t, validEmail("bukan-email"))
	assert.False(t, validEmail("Nama <a@uji.local>"))
	assert.False(t, validEmail("a@localhost"))
}
