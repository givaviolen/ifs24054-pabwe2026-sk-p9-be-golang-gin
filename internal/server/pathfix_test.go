package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func probe(target string) (path, query string) {
	h := RestoreOriginalPath(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", target, nil))
	return
}

func TestRestoreOriginalPath(t *testing.T) {
	p, q := probe("/api/index?__p=/api/public/posts&q=go")
	assert.Equal(t, "/api/public/posts", p)
	assert.Equal(t, "q=go", q)

	p, _ = probe("/api/index?__p=/")
	assert.Equal(t, "/", p)

	p, _ = probe("/api/index?__p=$1")
	assert.Equal(t, "/api/index", p)

	p, _ = probe("/api/index")
	assert.Equal(t, "/", p)

	p, _ = probe("/api/public/categories")
	assert.Equal(t, "/api/public/categories", p)
}
