package server

import (
	"net/http"
	"strings"
)

// RestoreOriginalPath membungkus handler agar path asli request dipulihkan.
//
// Di Vercel, rewrite (vercel.json) meneruskan path TUJUAN (/api/index) ke fungsi,
// bukan path yang diminta klien. vercel.json menyertakan path asli lewat query
// "__p"; middleware ini mengembalikannya sebelum request masuk ke router Gin.
// Tanpa "__p" (lokal / server biasa) request tidak diubah.
func RestoreOriginalPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if vals, ok := q["__p"]; ok {
			q.Del("__p")
			r.URL.RawQuery = q.Encode()

			p := ""
			if len(vals) > 0 {
				p = vals[0]
			}
			// Placeholder yang tidak terganti (":p", "$1") diabaikan.
			if p != "" && !strings.HasPrefix(p, ":") && !strings.HasPrefix(p, "$") {
				r.URL.Path = "/" + strings.TrimPrefix(p, "/")
				r.URL.RawPath = ""
				r.RequestURI = r.URL.RequestURI()
			}
		} else if r.URL.Path == "/api/index" || r.URL.Path == "/api/index.go" {
			r.URL.Path = "/"
		}
		next.ServeHTTP(w, r)
	})
}
