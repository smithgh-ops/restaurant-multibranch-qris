# RestoQRIS — Manajemen Restoran Multi-Cabang dengan QRIS

Platform manajemen restoran multi-cabang berbasis web dengan dukungan pembayaran QRIS (Quick Response Code Indonesian Standard). Dibangun dengan SvelteKit, Go Gin, MySQL 8, dan Redis.

---

## Prasyarat

| Alat            | Versi minimal |
|-----------------|---------------|
| Go              | 1.24          |
| Node.js         | 20            |
| npm             | 10            |
| Docker          | 24            |
| Docker Compose  | 2.24          |
| MySQL (opsional)| 8.0           |
| Redis (opsional)| 7             |

---

## Pengaturan Environment

```bash
cp .env.example .env
# Sesuaikan nilai di .env sesuai kebutuhan lokal Anda
```

> **Penting:** Jangan pernah commit file `.env` dengan kredensial nyata ke version control.

---

## Menjalankan dengan Docker Compose

```bash
# Build dan jalankan semua layanan (web, api, mysql, redis)
docker compose up --build

# Akses aplikasi:
# Frontend : http://localhost:3000
# API      : http://localhost:8080
# Health   : http://localhost:8080/health
```

Untuk menjalankan di background:

```bash
docker compose up -d --build
```

Untuk menghentikan:

```bash
docker compose down
```

Untuk menghapus volume (reset database):

```bash
docker compose down -v
```

---

## Menjalankan Lokal Tanpa Docker

### Backend (Go API)

```bash
cd apps/api

# Salin dan isi environment variable
export DB_HOST=localhost DB_USER=resto DB_PASSWORD=... DB_NAME=resto_db
export REDIS_ADDR=localhost:6379

# Jalankan server
go run ./cmd/api
# Server berjalan di http://localhost:8080
```

### Frontend (SvelteKit)

```bash
cd apps/web

npm install
npm run dev
# Frontend berjalan di http://localhost:5173
```

---

## Pendekatan Migrasi Database

Migrasi database tersimpan sebagai file SQL bernomor urut di `apps/api/migrations/`:

```
001_create_organizations_and_branches.sql
002_create_roles_and_users.sql
003_create_dining_areas_and_tables.sql
004_create_menu.sql
005_create_orders.sql
006_create_payments.sql
007_create_kitchen.sql
008_create_audit_logs.sql
009_seed_roles.sql
```

Saat menggunakan Docker Compose, MySQL akan **otomatis** menjalankan semua file SQL di `apps/api/migrations/` pada saat pertama kali container dibuat (via `docker-entrypoint-initdb.d`).

Untuk lingkungan produksi, disarankan menggunakan alat migrasi seperti [golang-migrate](https://github.com/golang-migrate/migrate) atau [goose](https://github.com/pressly/goose).

---

## Struktur Proyek

```
.
├── apps/
│   ├── api/                        # Go Gin backend
│   │   ├── cmd/api/main.go         # Entry point
│   │   ├── internal/
│   │   │   ├── config/             # Konfigurasi dari env vars
│   │   │   ├── database/           # Koneksi MySQL & Redis
│   │   │   ├── handler/            # HTTP handler (health, info, dst.)
│   │   │   ├── middleware/         # Logger, Recovery, CORS
│   │   │   ├── router/             # Registrasi route
│   │   │   └── server/             # HTTP server lifecycle
│   │   ├── migrations/             # File SQL migrasi terurut
│   │   ├── Dockerfile
│   │   └── go.mod
│   └── web/                        # SvelteKit frontend
│       ├── src/
│       │   ├── lib/api/client.ts   # Typed API client
│       │   └── routes/
│       │       ├── +page.svelte    # Halaman login
│       │       └── dashboard/      # Dashboard shell & halaman utama
│       ├── Dockerfile
│       └── package.json
├── docs/
│   └── PRD_Restaurant_Multi_Cabang_QRIS.md
├── .github/
│   └── workflows/ci.yml            # GitHub Actions CI
├── compose.yaml                    # Docker Compose
├── .env.example                    # Template environment variables
└── README.md
```

---

## Endpoint Utama

| Method | Path            | Deskripsi                          |
|--------|-----------------|------------------------------------|
| GET    | `/health`       | Status kesehatan server            |
| GET    | `/api/v1/info`  | Informasi aplikasi (nama, versi)   |

Endpoint selanjutnya (auth, branches, menus, orders, payments, kitchen, reports) akan ditambahkan pada sprint berikutnya.

---

## CI/CD

GitHub Actions menjalankan dua job paralel pada setiap push/PR ke `main` dan `develop`:

- **Go API**: `go vet` + `go build`
- **SvelteKit Web**: `npm ci` + `svelte-kit sync` + `svelte-check` + `npm run build`

---

## Catatan & Keterbatasan

- Autentikasi JWT belum diimplementasi; endpoint auth akan hadir di sprint berikutnya.
- Integrasi gateway QRIS nyata belum ada; interface adapter sudah disiapkan.
- WebSocket untuk KDS real-time belum diimplementasi.
- Migrasi menggunakan `docker-entrypoint-initdb.d` yang hanya berjalan sekali saat volume baru; untuk incremental migration di produksi gunakan `golang-migrate` atau `goose`.

---

## Dokumentasi

- [PRD — Product Requirements Document](docs/PRD_Restaurant_Multi_Cabang_QRIS.md)

