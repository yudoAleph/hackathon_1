# Quick Start Guide - Gin Server Port 9001

## Ringkasan

Server telah berhasil dikonfigurasi dengan:
- **Framework**: Gin (Go)
- **Port**: 9001
- **Health Check**: ✅ Tersedia di `/health`

## Struktur Proyek

```
hackathon_1/
├── cmd/
│   └── server/
│       └── main.go              # Entry point aplikasi
├── internal/
│   └── app/
│       ├── handler.go           # HTTP handlers (diupdate ke Gin)
│       ├── routes/
│       │   └── routes.go        # Route definitions (BARU)
│       ├── service.go           # Business logic
│       ├── repository.go        # Database layer
│       └── model.go             # Data models
├── pkg/
│   ├── config/                  # Configuration management
│   └── db/                      # Database initialization
└── docs/
    └── ROUTES.md                # API documentation (BARU)
```

## File yang Diubah/Dibuat

### ✅ File Baru:
1. `internal/app/routes/routes.go` - Route definitions
2. `docs/ROUTES.md` - API documentation

### ✅ File yang Diupdate:
1. `cmd/server/main.go` - Migrated dari Echo ke Gin, port 9001
2. `internal/app/handler.go` - Handlers diupdate untuk Gin
3. `Makefile` - Ditambahkan commands untuk port 9001
4. `README.md` - Updated dengan info server baru

## Cara Menjalankan

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Run Server
```bash
# Opsi 1: Langsung run
make run

# Opsi 2: Build dulu, lalu run
make build
make start

# Opsi 3: Manual
go run ./cmd/server/main.go
```

### 3. Test Server
```bash
# Health check
curl http://localhost:9001/health

# Ping
curl http://localhost:9001/api/v1/ping
```

## Response Endpoints

### Health Check (`GET /health`)
```json
{
  "status": "healthy",
  "service": "contact-management-api",
  "version": "1.0.0"
}
```

### Ping (`GET /api/v1/ping`)
```json
{
  "message": "pong"
}
```

## Makefile Commands

```bash
make run        # Run aplikasi langsung
make build      # Build binary ke bin/server
make start      # Run binary yang sudah di-build
make kill-9001  # Kill process di port 9001
```

## Integrasi dengan Struktur Existing

### Handler Pattern
Handlers menggunakan Gin context (`*gin.Context`) bukan Echo context:

```go
// Echo (Lama)
func (h *Handler) Ping(c echo.Context) error {
    return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
}

// Gin (Baru)
func (h *Handler) Ping(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
```

### Route Organization
Routes diorganisir di file terpisah `internal/app/routes/routes.go`:

```go
func SetupRoutes(router *gin.Engine, handler *app.Handler) {
    router.GET("/health", healthHandler)
    
    api := router.Group("/api/v1")
    {
        api.GET("/ping", handler.Ping)
        // Tambahkan routes lain di sini
    }
}
```

## Next Steps

1. **Tambah Authentication Middleware** (jika diperlukan)
2. **Tambah Logging Middleware** untuk monitoring
3. **Tambah CORS Configuration** untuk frontend
4. **Implement Swagger Documentation**

## Troubleshooting

### Port sudah digunakan?
```bash
make kill-9001
```

### Server tidak start?
Check logs di `server.log` atau run dengan:
```bash
go run ./cmd/server/main.go
```

### Database connection error?
Pastikan file `.env` sudah dikonfigurasi dengan benar.

## Dependencies

```
github.com/gin-gonic/gin v1.11.0  # Web framework
gorm.io/gorm v1.30.1              # ORM
gorm.io/driver/mysql v1.6.0       # MySQL driver
```

## Kontak & Support

Untuk dokumentasi lengkap API endpoints, lihat:
- `docs/ROUTES.md` - Route documentation
- `README.md` - Project overview
