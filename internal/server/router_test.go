package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
	cfg := &config.Config{AppKey: "kunci", AdminUsername: "admin", AdminPassword: "admin", AllowAdminRegister: true}
	db, err := database.NewTestDB()
	assert.NoError(t, err)
	assert.NoError(t, database.EnsureSchema(db, cfg))
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

func dataOf(res map[string]interface{}) map[string]interface{} {
	d, _ := res["data"].(map[string]interface{})
	return d
}

// daftar membuat akun lalu mengembalikan token dan id user.
func daftar(t *testing.T, r *gin.Engine, name, role string) (string, float64) {
	code, res := do(r, "POST", "/api/auth/register", "", gin.H{
		"username": name, "email": name + "@uji.local", "password": "Password123!", "role": role,
	})
	assert.Equal(t, 201, code)
	d := dataOf(res)
	user := d["user"].(map[string]interface{})
	return d["token"].(string), user["id"].(float64)
}

func TestAuthContract(t *testing.T) {
	r := setup(t, fakeLLM{})

	tok, _ := daftar(t, r, "giva", "")
	code, res := do(r, "POST", "/api/auth/register", "", gin.H{"username": "giva", "email": "giva@uji.local", "password": "Password123!"})
	assert.Equal(t, 400, code)
	assert.Equal(t, "fail", res["status"])
	code, _ = do(r, "POST", "/api/auth/register", "", gin.H{"username": "tanpaemail", "password": "Password123!"})
	assert.Equal(t, 400, code)
	code, _ = do(r, "POST", "/api/auth/register", "", gin.H{"username": "pendek", "email": "p@uji.local", "password": "123"})
	assert.Equal(t, 400, code)
	code, _ = do(r, "POST", "/api/auth/register", "", gin.H{"username": "emailbad", "email": "bukan-email", "password": "Password123!"})
	assert.Equal(t, 400, code)

	// login lewat username dan email
	code, res = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva", "password": "Password123!"})
	assert.Equal(t, 200, code)
	assert.Equal(t, "member", dataOf(res)["user"].(map[string]interface{})["role"])
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva@uji.local", "password": "Password123!"})
	assert.Equal(t, 200, code)
	code, res = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva", "password": "salah"})
	assert.Equal(t, 401, code)
	assert.Nil(t, res["data"])
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva"})
	assert.Equal(t, 400, code)

	// logout tanpa token tetap 200 dan tidak membocorkan token
	code, res = do(r, "POST", "/api/auth/logout", "", gin.H{})
	assert.Equal(t, 200, code)
	b, _ := json.Marshal(res)
	assert.False(t, strings.Contains(strings.ToLower(string(b)), "\"token\""))

	// profil
	code, res = do(r, "GET", "/api/auth/profile", tok, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "giva", dataOf(res)["username"])
	code, _ = do(r, "GET", "/api/auth/profile", "", nil)
	assert.Equal(t, 401, code)
	code, _ = do(r, "GET", "/api/auth/profile", "token.palsu.tidak.valid", nil)
	assert.Equal(t, 401, code)

	code, res = do(r, "PUT", "/api/auth/profile", tok, gin.H{"email": "baru@uji.local", "password": "PasswordBaru2026!"})
	assert.Equal(t, 200, code)
	assert.Equal(t, "baru@uji.local", dataOf(res)["email"])
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva", "password": "PasswordBaru2026!"})
	assert.Equal(t, 200, code)
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{"username": "giva", "password": "Password123!"})
	assert.Equal(t, 401, code)
	code, _ = do(r, "PUT", "/api/auth/profile", tok, gin.H{"email": "bukan-email"})
	assert.Equal(t, 400, code)
	code, _ = do(r, "PUT", "/api/auth/profile", tok, gin.H{"password": "123"})
	assert.Equal(t, 400, code)
	code, _ = do(r, "PUT", "/api/auth/profile", tok, gin.H{"email": "member@delcom.org"})
	assert.Equal(t, 400, code)
	code, _ = do(r, "PUT", "/api/auth/profile", "", gin.H{"email": "x@uji.local"})
	assert.Equal(t, 401, code)
}

func TestPostContract(t *testing.T) {
	r := setup(t, fakeLLM{})
	tok, uid := daftar(t, r, "penulis", "")
	other, _ := daftar(t, r, "lain", "")

	// create
	body := gin.H{"title": "Artikel Uji Golang", "category": "Teknologi", "content": "Isi artikel", "is_published": true}
	code, res := do(r, "POST", "/api/posts", tok, body)
	assert.Equal(t, 201, code)
	d := dataOf(res)
	id := fmt.Sprint(d["id"])
	assert.Equal(t, "artikel-uji-golang", d["slug"])
	assert.Equal(t, uid, d["user_id"])
	assert.NotNil(t, d["published_at"])

	code, res = do(r, "POST", "/api/posts", tok, gin.H{"title": "Artikel Uji Golang", "content": "draf", "is_published": false})
	assert.Equal(t, 201, code)
	assert.Equal(t, "artikel-uji-golang-2", dataOf(res)["slug"])
	assert.Equal(t, "Umum", dataOf(res)["category"])
	assert.Nil(t, dataOf(res)["published_at"])
	draftID := fmt.Sprint(dataOf(res)["id"])

	code, _ = do(r, "POST", "/api/posts", tok, gin.H{"content": "tanpa judul"})
	assert.Equal(t, 400, code)
	code, _ = do(r, "POST", "/api/posts", tok, gin.H{"title": "tanpa konten"})
	assert.Equal(t, 400, code)
	code, res = do(r, "POST", "/api/posts", "", body)
	assert.Equal(t, 401, code)
	assert.Equal(t, "fail", res["status"])

	// read
	code, res = do(r, "GET", "/api/posts", tok, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res["data"], 2)
	code, res = do(r, "GET", "/api/posts/"+id, tok, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "Artikel Uji Golang", dataOf(res)["title"])
	code, res = do(r, "GET", "/api/posts/stats", tok, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(2), dataOf(res)["total_articles"])
	assert.Equal(t, float64(1), dataOf(res)["total_selesai"])
	assert.Equal(t, float64(1), dataOf(res)["belum_selesai"])

	code, _ = do(r, "GET", "/api/posts/"+id, "", nil)
	assert.Equal(t, 401, code)
	code, _ = do(r, "GET", "/api/posts/99999999", tok, nil)
	assert.Equal(t, 404, code)
	code, _ = do(r, "GET", "/api/posts/bukan-angka", tok, nil)
	assert.Equal(t, 400, code)
	code, _ = do(r, "GET", "/api/posts/"+id, other, nil)
	assert.Equal(t, 403, code)

	// update: field kosong diabaikan; kepemilikan dicek
	code, res = do(r, "PUT", "/api/posts/"+id, tok, gin.H{"title": "Judul Baru", "content": "Konten baru"})
	assert.Equal(t, 200, code)
	assert.Equal(t, "Judul Baru", dataOf(res)["title"])
	assert.Equal(t, "Teknologi", dataOf(res)["category"])
	code, _ = do(r, "PUT", "/api/posts/"+id, "", gin.H{"title": "x"})
	assert.Equal(t, 401, code)
	code, _ = do(r, "PUT", "/api/posts/99999999", tok, gin.H{"title": "x"})
	assert.Equal(t, 404, code)
	code, _ = do(r, "PUT", "/api/posts/"+id, other, gin.H{"title": "Dibajak"})
	assert.Equal(t, 403, code)
	code, _ = do(r, "PUT", "/api/posts/abc", tok, gin.H{"title": "x"})
	assert.Equal(t, 400, code)

	// toggle
	code, res = do(r, "PATCH", "/api/posts/"+draftID+"/toggle-publish", tok, nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, true, dataOf(res)["is_published"])
	assert.NotNil(t, dataOf(res)["published_at"])
	code, res = do(r, "PATCH", "/api/posts/"+draftID+"/toggle-publish", tok, nil)
	assert.Equal(t, false, dataOf(res)["is_published"])
	code, _ = do(r, "PATCH", "/api/posts/99999999/toggle-publish", tok, nil)
	assert.Equal(t, 404, code)
	code, _ = do(r, "PATCH", "/api/posts/"+id+"/toggle-publish", "", nil)
	assert.Equal(t, 401, code)
	code, _ = do(r, "PATCH", "/api/posts/"+id+"/toggle-publish", other, nil)
	assert.Equal(t, 403, code)

	// publik: tanpa user_id, ada author_name
	code, res = do(r, "GET", "/api/public/posts?q=konten&category=Teknologi", "", nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res["data"], 1)
	raw, _ := json.Marshal(res["data"])
	assert.False(t, strings.Contains(string(raw), "user_id"))
	assert.Contains(t, string(raw), "author_name")
	code, res = do(r, "GET", "/api/public/posts/artikel-uji-golang", "", nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "penulis", dataOf(res)["author_name"])
	code, _ = do(r, "GET", "/api/public/posts/"+"tidak-ada", "", nil)
	assert.Equal(t, 404, code)
	code, res = do(r, "GET", "/api/public/categories", "", nil)
	assert.Equal(t, 200, code)
	assert.NotEmpty(t, res["data"])

	// delete
	code, _ = do(r, "DELETE", "/api/posts/"+id, "", nil)
	assert.Equal(t, 401, code)
	code, _ = do(r, "DELETE", "/api/posts/"+id, other, nil)
	assert.Equal(t, 403, code)
	code, _ = do(r, "DELETE", "/api/posts/99999999", tok, nil)
	assert.Equal(t, 404, code)
	code, _ = do(r, "DELETE", "/api/posts/abc", tok, nil)
	assert.Equal(t, 400, code)
	code, _ = do(r, "DELETE", "/api/posts/"+id, tok, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/posts/"+id, tok, nil)
	assert.Equal(t, 404, code)
}

func TestAdminContract(t *testing.T) {
	r := setup(t, fakeLLM{})
	member, memberID := daftar(t, r, "anggota", "")
	admin, _ := daftar(t, r, "adminuji", "admin")
	do(r, "POST", "/api/posts", member, gin.H{"title": "Artikel Anggota", "content": "isi", "is_published": true})

	code, res := do(r, "GET", "/api/admin/users", admin, nil)
	assert.Equal(t, 200, code)
	users := res["data"].([]interface{})
	assert.GreaterOrEqual(t, len(users), 3) // admin seed, member seed, anggota, adminuji
	found := false
	for _, u := range users {
		m := u.(map[string]interface{})
		if m["username"] == "anggota" {
			found = true
			assert.Equal(t, float64(1), m["total_articles"])
			assert.Equal(t, float64(1), m["total_published"])
		}
	}
	assert.True(t, found)

	code, res = do(r, "GET", "/api/admin/posts", admin, nil)
	assert.Equal(t, 200, code)
	assert.NotEmpty(t, res["data"])
	code, res = do(r, "GET", fmt.Sprintf("/api/admin/users/%d/posts", int(memberID)), admin, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res["data"], 1)
	code, res = do(r, "GET", "/api/admin/users/99999999/posts", admin, nil)
	assert.Equal(t, 200, code)
	code, _ = do(r, "GET", "/api/admin/users/abc/posts", admin, nil)
	assert.Equal(t, 400, code)

	for _, path := range []string{"/api/admin/users", "/api/admin/posts", "/api/admin/users/1/posts"} {
		code, _ = do(r, "GET", path, "", nil)
		assert.Equal(t, 401, code)
		code, res = do(r, "GET", path, member, nil)
		assert.Equal(t, 403, code)
		assert.Equal(t, "fail", res["status"])
	}

	// akun admin bawaan dari seed bisa login
	code, _ = do(r, "POST", "/api/auth/login", "", gin.H{"username": "admin", "password": "admin"})
	assert.Equal(t, 200, code)
}

func TestAdminRegisterDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppKey: "kunci", AdminUsername: "admin", AdminPassword: "admin", AllowAdminRegister: false}
	db, _ := database.NewTestDB()
	assert.NoError(t, database.EnsureSchema(db, cfg))
	r := NewRouter(cfg, db, &Deps{Searcher: fakeSearch{}, Completer: fakeLLM{}})
	code, res := do(r, "POST", "/api/auth/register", "", gin.H{"username": "mau-admin", "email": "m@uji.local", "password": "Password123!", "role": "admin"})
	assert.Equal(t, 201, code)
	assert.Equal(t, "member", dataOf(res)["user"].(map[string]interface{})["role"])
}

func TestAIContract(t *testing.T) {
	r := setup(t, fakeLLM{})
	tok, _ := daftar(t, r, "penulisai", "")

	code, res := do(r, "POST", "/api/posts/ai-generate", tok, gin.H{"topic": "Golang terbaru"})
	assert.Equal(t, 200, code)
	assert.Len(t, dataOf(res)["sources"], 1)
	assert.NotEmpty(t, dataOf(res)["excerpt"])

	code, res = do(r, "POST", "/api/posts/ai-generate", tok, gin.H{"category": "Teknologi"})
	assert.Equal(t, 400, code)
	assert.Contains(t, strings.ToLower(res["message"].(string)), "topic")
	code, _ = do(r, "POST", "/api/posts/ai-generate", "", gin.H{"topic": "x"})
	assert.Equal(t, 401, code)
	code, _ = do(r, "POST", "/api/posts/ai-generate", "token.palsu.tidak.valid", gin.H{"topic": "x"})
	assert.Equal(t, 401, code)

	// OpenAI gagal -> 500 status "error" berisi kata "OpenAI"
	r2 := setup(t, fakeLLM{fail: true})
	tok2, _ := daftar(t, r2, "penulisai", "")
	code, res = do(r2, "POST", "/api/posts/ai-generate", tok2, gin.H{"topic": "Golang"})
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, "error", res["status"])
	assert.Contains(t, strings.ToLower(res["message"].(string)), "openai")
}

func TestBaseAndNotFound(t *testing.T) {
	r := setup(t, fakeLLM{})
	code, res := do(r, "GET", "/", "", nil)
	assert.Equal(t, 200, code)
	assert.Equal(t, "success", res["status"])
	code, _ = do(r, "GET", "/api.json", "", nil)
	assert.Equal(t, 200, code)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/swagger/index.html", nil))
	assert.Equal(t, 200, w.Code)
	code, res = do(r, "GET", "/api/route-tidak-ada-uji", "", nil)
	assert.Equal(t, 404, code)
	assert.Equal(t, "fail", res["status"])
	assert.Equal(t, "Endpoint not found", res["message"])
}
