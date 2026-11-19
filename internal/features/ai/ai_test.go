package ai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeSearch struct {
	out []Source
	err error
}

func (f fakeSearch) Search(context.Context, string) ([]Source, error) { return f.out, f.err }

type fakeLLM struct {
	out string
	err error
}

func (f fakeLLM) Complete(context.Context, string, string) (string, error) { return f.out, f.err }

const goodJSON = "```json\n{\"title\":\"Judul\",\"category\":\"Teknologi\",\"excerpt\":\"Ringkas\",\"content\":\"Isi [1]\"}\n```"

func TestGenerateWithSources(t *testing.T) {
	svc := NewService(fakeSearch{out: []Source{{Title: "A", URL: "http://a", Snippet: "s"}}}, fakeLLM{out: goodJSON})
	g, err := svc.Generate(context.Background(), GenerateRequest{Topic: "Go", Category: "Bisnis"})
	assert.NoError(t, err)
	assert.Equal(t, "Bisnis", g.Category)
	assert.Len(t, g.Sources, 1)
	assert.Contains(t, g.Content, "## Sumber Rujukan")
}

func TestGenerateSearchFailsOrDisabled(t *testing.T) {
	off := false
	svc := NewService(fakeSearch{err: errors.New("down")}, fakeLLM{out: goodJSON})
	g, err := svc.Generate(context.Background(), GenerateRequest{Topic: "Go"})
	assert.NoError(t, err)
	assert.Empty(t, g.Sources)
	assert.Equal(t, "Teknologi", g.Category)

	g, err = svc.Generate(context.Background(), GenerateRequest{Topic: "Go", IncludeWebSearch: &off})
	assert.NoError(t, err)
	assert.NotContains(t, g.Content, "Sumber Rujukan")
}

func TestGenerateErrors(t *testing.T) {
	_, err := NewService(fakeSearch{}, fakeLLM{err: errors.New("x")}).Generate(context.Background(), GenerateRequest{Topic: "Go"})
	assert.Error(t, err)
	for _, bad := range []string{"tanpa json", "{rusak", `{"title":"","content":""}`} {
		_, err = NewService(fakeSearch{}, fakeLLM{out: bad}).Generate(context.Background(), GenerateRequest{Topic: "Go"})
		assert.ErrorIs(t, err, ErrBadModelOutput)
	}
}

func TestSearchClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "go", r.URL.Query().Get("q"))
		_, _ = w.Write([]byte(`{"data":[{"title":"T","link":"http://l","description":"d"}]}`))
	}))
	defer srv.Close()
	out, err := NewSearchClient(srv.URL).Search(context.Background(), "go")
	assert.NoError(t, err)
	assert.Equal(t, Source{Title: "T", URL: "http://l", Snippet: "d"}, out[0])

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer bad.Close()
	_, err = NewSearchClient(bad.URL).Search(context.Background(), "go")
	assert.Error(t, err)
}

func TestOpenAIClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer kunci", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"halo"}}]}`))
	}))
	defer srv.Close()
	out, err := NewOpenAIClient(srv.URL+"/", "kunci", "m").Complete(context.Background(), "s", "u")
	assert.NoError(t, err)
	assert.Equal(t, "halo", out)

	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"choices":[]}`)) }))
	defer empty.Close()
	_, err = NewOpenAIClient(empty.URL, "k", "m").Complete(context.Background(), "s", "u")
	assert.Error(t, err)
}
