package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeSearch struct{}

func (fakeSearch) Search(context.Context, string) ([]ai.Source, error) {
	return []ai.Source{{Title: "Ref", URL: "http://ref", Snippet: "s"}}, nil
}

type fakeLLM struct{ fail bool }

func (f fakeLLM) Complete(context.Context, string, string) (string, error) {
	if f.fail {
		return "", errors.New("down")
	}
	return `{"title":"Artikel AI","category":"Teknologi","excerpt":"Ringkas","content":"Isi artikel"}`, nil
}

func setup(t *testing.T, llm fakeLLM) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppKey: "kunci", AdminUsername: "admin", AdminPassword: "admin"}
	db, err := database.NewTestDB()
	assert.NoError(t, err)
	assert.NoError(t, database.Migrate(db))
	assert.NoError(t, database.SeedAdmin(db, cfg))
	return NewRouter(cfg, db, &Deps{Searcher: fakeSearch{}, Completer: llm})
}

func do(r *gin.Engine, method, path, token string, body interface{}) (int, map[string]interface{}) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func login(t *testing.T, r *gin.Engine, u, p string) string {
	code, res := do(r, "POST", "/api/auth/login", "", gin.H{"username": u, "password": p})
	assert.Equal(t, 200, code)
	return res["data"].(map[string]interface{})["token"].(string)
}

func TestFullFlow(t *testing.T) {
	r := setup(t, fakeLLM{})

	// register + duplikat + validasi
	code, _ := do(r, "POST", "/api/auth/register", "", gin.H{"username": "giva", "email": "giva@x.id", "password": "rahasia"})
	assert.Equal(t, 201, code)
	code, _ = do(r, "POST", "/api/auth/register", "", gin.H{"username": "giva", "email": "giva@x.id", "password": "rahasia"})
	assert.Equal(t, 409, code)
	code, _ = do(r, "POST", "/api/auth/register", "", gin.H{"username": "g"})
	assert.Equal(t, 400, code)

	// login gagal
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva", "password": "salah"})
	assert.Equal(t, 401, code)
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{})
	assert.Equal(t, 400, code)

	tok := login(t, r, "giva", "rahasia")

	// profil
	code, _ = do(r, "GET", "/api/auth/profile", tok, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "PUT", "/api/auth/profile", tok, gin.H{"username": "giva2", "email": "giva@x.id", "password": "baru123"})
	assert.Equal(t, 200, code)
	code, _ = do(r, "PUT", "/api/auth/profile", tok, gin.H{"username": "admin", "email": "giva@x.id"})
	assert.Equal(t, 409, code)
	code, _ = do(r, "PUT", "/api/auth/profile", tok, gin.H{})
	assert.Equal(t, 400, code)
	code, _ = do(r, "POST", "/api/auth/logout", tok, nil)
	assert.Equal(t, 200, code)

	// buat artikel (slug unik)
	body := gin.H{"title": "Belajar Go", "category": "Teknologi", "content": "isi artikel", "is_published": true}
	code, res := do(r, "POST", "/api/posts", tok, body)
	assert.Equal(t, 201, code)
	slug := res["data"].(map[string]interface{})["slug"].(string)
	assert.Equal(t, "belajar-go", slug)
	code, res = do(r, "POST", "/api/posts", tok, gin.H{"title": "Belajar Go", "content": "draf"})
	assert.Equal(t, 201, code)
	assert.Equal(t, "belajar-go-2", res["data"].(map[string]interface{})["slug"])
	id := fmt.Sprint(res["data"].(map[string]interface{})["id"])
	code, _ = do(r, "POST", "/api/posts", tok, gin.H{"title": ""})
	assert.Equal(t, 400, code)

	// publik
	code, res = do(r, "GET", "/api/public/posts?q=belajar&category=Teknologi", "", nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res["data"], 1)
	code, _ = do(r, "GET", "/api/public/posts/belajar-go", "", nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/public/posts/belajar-go-2", "", nil)
	assert.Equal(t, 404, code)
	code, _ = do(r, "GET", "/api/public/categories", "", nil)
	assert.Equal(t, 200, code)

	// penulis
	code, res = do(r, "GET", "/api/posts/stats", tok, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(2), res["data"].(map[string]interface{})["total"])
	assert.Equal(t, float64(1), res["data"].(map[string]interface{})["draft"])
	code, _ = do(r, "GET", "/api/posts", tok, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/posts/"+id, tok, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/posts/999", tok, nil)
	assert.Equal(t, 404, code)
	code, _ = do(r, "GET", "/api/posts/abc", tok, nil)
	assert.Equal(t, 400, code)

	pub := true
	code, res = do(r, "PUT", "/api/posts/"+id, tok, gin.H{"title": "Judul Baru", "content": "isi baru", "is_published": pub})
	assert.Equal(t, 200, code)
	assert.Equal(t, "judul-baru", res["data"].(map[string]interface{})["slug"])
	code, _ = do(r, "PUT", "/api/posts/"+id, tok, gin.H{"title": ""})
	assert.Equal(t, 400, code)
	code, _ = do(r, "PUT", "/api/posts/999", tok, gin.H{"title": "x", "content": "y"})
	assert.Equal(t, 404, code)

	code, res = do(r, "PATCH", "/api/posts/"+id+"/toggle-publish", tok, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, false, res["data"].(map[string]interface{})["is_published"])
	code, _ = do(r, "PATCH", "/api/posts/999/toggle-publish", tok, nil)
	assert.Equal(t, 404, code)

	// AI generator
	code, res = do(r, "POST", "/api/posts/ai-generate", tok, gin.H{"topic": "Golang terbaru"})
	assert.Equal(t, 200, code)
	assert.Len(t, res["data"].(map[string]interface{})["sources"], 1)
	code, _ = do(r, "POST", "/api/posts/ai-generate", tok, gin.H{})
	assert.Equal(t, 400, code)

	// admin
	code, _ = do(r, "GET", "/api/admin/users", tok, nil)
	assert.Equal(t, 403, code)
	adm := login(t, r, "admin", "admin")
	code, res = do(r, "GET", "/api/admin/users", adm, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res["data"], 2)
	code, _ = do(r, "GET", "/api/admin/posts", adm, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/admin/users/2/posts", adm, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/admin/users/x/posts", adm, nil)
	assert.Equal(t, 400, code)

	// hapus
	code, _ = do(r, "DELETE", "/api/posts/"+id, tok, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "DELETE", "/api/posts/"+id, tok, nil)
	assert.Equal(t, 404, code)
	code, _ = do(r, "DELETE", "/api/posts/abc", tok, nil)
	assert.Equal(t, 400, code)
}

func TestAIFailure(t *testing.T) {
	r := setup(t, fakeLLM{fail: true})
	do(r, "POST", "/api/auth/register", "", gin.H{"username": "giva", "email": "g@x.id", "password": "rahasia"})
	tok := login(t, r, "giva", "rahasia")
	code, _ := do(r, "POST", "/api/posts/ai-generate", tok, gin.H{"topic": "Golang"})
	assert.Equal(t, http.StatusBadGateway, code)
}

func TestDocsAndNotFound(t *testing.T) {
	r := setup(t, fakeLLM{})
	code, _ := do(r, "GET", "/api.json", "", nil)
	assert.Equal(t, 200, code)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/swagger/index.html", nil))
	assert.Equal(t, 200, w.Code)
	code, res := do(r, "GET", "/tidak-ada", "", nil)
	assert.Equal(t, 404, code)
	assert.Equal(t, "Endpoint not found", res["message"])
}
