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
│   │   ├── cmd/api/main.go         # Entry point API
│   │   ├── cmd/seed/main.go        # CLI bootstrap admin pertama
│   │   ├── internal/
│   │   │   ├── auth/               # Login, refresh, logout, JWT, bcrypt
│   │   │   ├── branch/             # CRUD cabang
│   │   │   ├── menu/               # Kategori, item, setting per-cabang
│   │   │   ├── organization/       # Data organisasi
│   │   │   ├── config/             # Konfigurasi dari env vars
│   │   │   ├── database/           # Koneksi MySQL & Redis
│   │   │   ├── handler/            # HTTP handler (health, info)
│   │   │   ├── middleware/         # Logger, Recovery, CORS, Auth JWT
│   │   │   ├── router/             # Registrasi route
│   │   │   └── server/             # HTTP server lifecycle
│   │   ├── migrations/             # File SQL migrasi terurut (001–011)
│   │   ├── Dockerfile
│   │   └── go.mod
│   └── web/                        # SvelteKit frontend
│       ├── src/
│       │   ├── hooks.server.ts     # SSR auth middleware, route protection
│       │   ├── lib/api/client.ts   # Typed API client
│       │   └── routes/
│       │       ├── +page.svelte          # Halaman login (real API)
│       │       ├── +page.server.ts       # Login form action
│       │       ├── logout/               # Logout action endpoint
│       │       └── dashboard/            # Dashboard shell & sub-halaman
│       │           ├── branches/         # Manajemen cabang
│       │           └── menu/             # Manajemen menu & kategori
│       ├── Dockerfile
│       └── package.json
├── docs/
│   ├── PRD_Restaurant_Multi_Cabang_QRIS.md
│   └── PROGRESS.md                 # Catatan progres pengembangan
├── .github/
│   └── workflows/ci.yml            # GitHub Actions CI
├── compose.yaml                    # Docker Compose
├── .env.example                    # Template environment variables
└── README.md
```

---

## Endpoint Utama

| Method | Path                              | Auth | Deskripsi                                |
|--------|-----------------------------------|------|------------------------------------------|
| GET    | `/health`                         | —    | Status kesehatan server                  |
| GET    | `/api/v1/info`                    | —    | Informasi aplikasi (nama, versi)         |
| POST   | `/api/v1/auth/login`              | —    | Login, mendapat access + refresh token   |
| POST   | `/api/v1/auth/refresh`            | —    | Refresh access token (rotasi)            |
| POST   | `/api/v1/auth/logout`             | —    | Logout, revoke refresh token             |
| GET    | `/api/v1/auth/me`                 | JWT  | Profil user + daftar role                |
| GET    | `/api/v1/organization`            | JWT  | Data organisasi pengguna                 |
| GET    | `/api/v1/branches`                | JWT  | Daftar cabang                            |
| POST   | `/api/v1/branches`                | JWT  | Buat cabang baru                         |
| GET    | `/api/v1/branches/:id`            | JWT  | Detail cabang                            |
| PATCH  | `/api/v1/branches/:id`            | JWT  | Update cabang                            |
| GET    | `/api/v1/menu/categories`         | JWT  | Daftar kategori menu                     |
| POST   | `/api/v1/menu/categories`         | JWT  | Buat kategori                            |
| PATCH  | `/api/v1/menu/categories/:id`     | JWT  | Update kategori                          |
| GET    | `/api/v1/menu/items`              | JWT  | Daftar item menu (filter tersedia)       |
| POST   | `/api/v1/menu/items`              | JWT  | Buat item menu                           |
| GET    | `/api/v1/menu/items/:id`          | JWT  | Detail item                              |
| PATCH  | `/api/v1/menu/items/:id`          | JWT  | Update item                              |
| GET    | `/api/v1/menu/items/:id/branches` | JWT  | Setting harga per-cabang                 |
| PUT    | `/api/v1/menu/items/:id/branches/:branch_id` | JWT | Upsert setting per-cabang   |

Endpoint selanjutnya (orders, payments, kitchen, reports) akan ditambahkan pada fase berikutnya.

---

## CI/CD

GitHub Actions menjalankan dua job paralel pada setiap push/PR ke `main`, `develop`, dan `copilot/*`:

- **Go API**: `go vet` + `go build` + `go test ./...`
- **SvelteKit Web**: `npm ci` + `svelte-kit sync` + `svelte-check` + `npm run build`

---

## Membuat Admin Pertama (Bootstrap)

Setelah migrasi database berhasil dijalankan, buat akun admin pertama menggunakan perintah CLI berikut:

```bash
cd apps/api

# Pastikan environment database sudah di-set
export DB_HOST=localhost DB_PORT=3306
export DB_USER=resto DB_PASSWORD=<password_db> DB_NAME=resto_db

go run ./cmd/seed \
  --org  "Nama Restoran Anda" \
  --slug "nama-restoran" \
  --name "Admin Utama" \
  --email "admin@restoran.com" \
  --pass  "password_kuat_anda"
```

> **Keamanan:**
> - Jangan commit password ke source control.
> - Hapus password dari shell history setelah selesai:
>   ```bash
>   history -d $(history 1 | awk '{print $1}')
>   ```
> - Password di-hash menggunakan bcrypt sebelum disimpan — tidak ada plaintext di database.
> - Perintah `seed` hanya untuk bootstrap awal; nonaktifkan atau hapus setelah admin dibuat.

---


- [PRD — Product Requirements Document](docs/PRD_Restaurant_Multi_Cabang_QRIS.md)
- [PROGRESS — Catatan Progres Pengembangan](docs/PROGRESS.md)

---

## Catatan

- Integrasi gateway QRIS nyata belum ada; akan diimplementasi pada Fase 4.
- WebSocket untuk KDS real-time belum diimplementasi; direncanakan pada Fase 5.
- Migrasi menggunakan `docker-entrypoint-initdb.d` yang hanya berjalan sekali saat volume baru; untuk incremental migration di produksi gunakan `golang-migrate` atau `goose`.
