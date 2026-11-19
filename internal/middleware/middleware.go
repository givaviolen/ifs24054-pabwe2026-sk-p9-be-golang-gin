package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"app/internal/features/auth"
	"app/internal/response"

	"github.com/gin-gonic/gin"
)

// CORS mengizinkan akses lintas origin (frontend dideploy terpisah).
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Auth memvalidasi Bearer JWT dan menyimpan user_id, username, role di context.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "Token tidak ditemukan")
			return
		}
		claims, err := auth.ParseToken(secret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "Token tidak valid atau kedaluwarsa")
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole membatasi akses hanya untuk role tertentu.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			response.Fail(c, http.StatusForbidden, "Akses ditolak")
			return
		}
		c.Next()
	}
}

// RateLimiter membatasi jumlah request per IP dalam satu jendela waktu.
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	hits := map[string][]time.Time{}
	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()

		mu.Lock()
		recent := hits[ip][:0]
		for _, t := range hits[ip] {
			if now.Sub(t) < window {
				recent = append(recent, t)
			}
		}
		allowed := len(recent) < limit
		if allowed {
			recent = append(recent, now)
		}
		hits[ip] = recent
		mu.Unlock()

		if !allowed {
			response.Fail(c, http.StatusTooManyRequests, "Terlalu banyak permintaan, coba lagi nanti")
			return
		}
		c.Next()
	}
}
