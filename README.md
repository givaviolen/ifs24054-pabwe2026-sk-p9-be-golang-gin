# Delcom Blog — Backend (Golang Gin, Port 8081)

REST API Clean Architecture & Feature-First: `auth`, `post`, `ai` (Search Engine + OpenAI).

## Menjalankan
```bash
psql -U postgres -c "CREATE DATABASE pabwe_p9_blog;"
cp .env.example .env
go mod tidy
go run cmd/server/main.go        # http://localhost:8081
# Swagger UI: http://localhost:8081/swagger/index.html
```
Akun admin otomatis dibuat dari `ADMIN_USERNAME` / `ADMIN_PASSWORD` (default admin/admin).

## Pengujian
```bash
go test -v -cover ./internal/...
```

## Build
```bash
go build -o server.exe cmd/server/main.go
```
