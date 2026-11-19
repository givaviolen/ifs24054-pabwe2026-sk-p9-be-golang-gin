package post

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	assert.Equal(t, "belajar-go-itu-seru", Slugify("  Belajar Go itu SERU!  "))
	assert.Equal(t, "artikel", Slugify("!!!"))
}

func TestReadingTime(t *testing.T) {
	assert.Equal(t, 1, ReadingTime(""))
	assert.Equal(t, 1, ReadingTime(strings.Repeat("kata ", 200)))
	assert.Equal(t, 2, ReadingTime(strings.Repeat("kata ", 201)))
}

func TestNormalizeCategory(t *testing.T) {
	assert.Equal(t, "Teknologi", NormalizeCategory(" teknologi "))
	assert.Equal(t, "Umum", NormalizeCategory("tidak-ada"))
}

func TestMakeExcerpt(t *testing.T) {
	assert.Equal(t, "a b", makeExcerpt("a\n  b"))
	assert.True(t, strings.HasSuffix(makeExcerpt(strings.Repeat("x", 300)), "..."))
}
