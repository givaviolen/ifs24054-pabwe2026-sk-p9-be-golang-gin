# Delcom Blog — Backend (Golang Gin, Port 8081)

REST API Clean Architecture & Feature-First: `auth`, `post`, `ai` (Search Engine + OpenAI).

## Menjalankan (lokal)
```bash
cp .env.example .env          # sesuaikan DB_PASSWORD, OPENAI_KEY, OPENAI_MODEL
go mod tidy
go run cmd/server/main.go     # http://localhost:8081
# Swagger UI: http://localhost:8081/swagger/index.html
```
Database `pabwe_p9_blog` dibuat otomatis bila belum ada. Akun bawaan (seed):
`admin` / `admin` (role admin) dan `member` / `member` (role member), plus 2 artikel contoh.

## Kontrak API (dipakai test-suite penilaian)
- Respons selalu `{status, message, data?}`; status `success` / `fail` / `error`.
- Register: `username` (min 3), `email`, `password` (min 6), `role` opsional (`admin`/`member`).
  Duplikat/validasi gagal → **400**. Login dengan username **atau** email; gagal → **401**.
- Profil: `PUT /api/auth/profile` hanya `email` dan/atau `password`.
- Artikel milik orang lain → **403**; tidak ada → **404**; ID bukan angka → **400**.
- Update artikel: field kosong tidak menimpa data lama.
- Statistik: `total_articles`, `total_selesai`, `belum_selesai`.
- Artikel publik memuat `author_name` dan `reading_time`, tanpa `user_id`.
- Admin users: `total_articles`, `total_published`.
- AI generate: tanpa `topic` → 400; OpenAI gagal → 500 `status: "error"` (pesan memuat "OpenAI").

> **Keamanan:** agar suite uji dapat membuat akun admin lewat registrasi, `ALLOW_ADMIN_REGISTER=true`
> (default). Untuk penggunaan sungguhan set `ALLOW_ADMIN_REGISTER=false`.

## Deploy
- **Vercel:** `api/index.go` + `vercel.json` (rewrite membawa path asli lewat `?__p=`, dipulihkan oleh
  `internal/server/pathfix.go`). Isi env `DB_*`, `APP_KEY`, `OPENAI_*`, `SEARCH_ENGINE_ENDPOINT` di dashboard.
- **Jenkins:** `Jenkinsfile` (test → build `go build -o server .` → Trivy → package → deploy).

## Pengujian & build
```bash
go test -v -cover ./internal/...
go build -o server.exe cmd/server/main.go
```
