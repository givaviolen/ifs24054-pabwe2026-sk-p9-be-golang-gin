package post

import (
	"strings"
	"testing"
	"time"

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
	assert.Equal(t, "a b", MakeExcerpt("a\n  b"))
	assert.True(t, strings.HasSuffix(MakeExcerpt(strings.Repeat("x", 300)), "..."))
}

func TestApplyPublish(t *testing.T) {
	p := &Post{}
	applyPublish(p, true)
	assert.True(t, p.IsPublished)
	assert.NotNil(t, p.PublishedAt)
	first := *p.PublishedAt
	time.Sleep(time.Millisecond)
	applyPublish(p, true)
	assert.Equal(t, first, *p.PublishedAt)
	applyPublish(p, false)
	assert.False(t, p.IsPublished)
	assert.Nil(t, p.PublishedAt)
}

func TestPublicResponseHidesUserID(t *testing.T) {
	res := toPublic(&Post{Title: "T", Content: "isi"})
	assert.Equal(t, "Anonim", res.AuthorName)
	assert.Equal(t, 1, res.ReadingTime)
}
