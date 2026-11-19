package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app/internal/features/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func engine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/me", Auth("k"), func(c *gin.Context) { c.String(200, c.GetString("role")) })
	r.GET("/admin", Auth("k"), RequireRole("admin"), func(c *gin.Context) { c.String(200, "ok") })
	r.GET("/limited", RateLimiter(2, time.Minute), func(c *gin.Context) { c.String(200, "ok") })
	return r
}

func call(r *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAuthAndRole(t *testing.T) {
	r := engine()
	assert.Equal(t, 401, call(r, "GET", "/me", "").Code)
	assert.Equal(t, 401, call(r, "GET", "/me", "salah").Code)

	member, _ := auth.GenerateToken("k", &auth.User{ID: 1, Role: "member"})
	admin, _ := auth.GenerateToken("k", &auth.User{ID: 2, Role: "admin"})
	assert.Equal(t, 200, call(r, "GET", "/me", member).Code)
	assert.Equal(t, 403, call(r, "GET", "/admin", member).Code)
	assert.Equal(t, 200, call(r, "GET", "/admin", admin).Code)
}

func TestCORSPreflight(t *testing.T) {
	w := call(engine(), http.MethodOptions, "/me", "")
	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRateLimiter(t *testing.T) {
	r := engine()
	assert.Equal(t, 200, call(r, "GET", "/limited", "").Code)
	assert.Equal(t, 200, call(r, "GET", "/limited", "").Code)
	assert.Equal(t, 429, call(r, "GET", "/limited", "").Code)
}
