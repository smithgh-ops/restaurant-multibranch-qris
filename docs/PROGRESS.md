# RestoQRIS — Catatan Progres Pengembangan

> Dokumen ini mencatat status pengerjaan platform RestoQRIS secara kronologis.
> Perbarui dokumen ini setiap kali fase atau fitur baru selesai dikerjakan.

---

## Ringkasan Platform

**RestoQRIS** adalah platform manajemen restoran multi-cabang berbasis web dengan fitur utama:

- Manajemen cabang, menu, dan pesanan terpusat
- Sistem pembayaran QRIS dinamis terintegrasi webhook
- Kitchen Display System (KDS) berbasis WebSocket
- Laporan penjualan dan analitik per-cabang
- Role-Based Access Control (RBAC) berbasis cabang

**Stack teknologi:**

- **Frontend:** SvelteKit + TypeScript + Tailwind CSS (adapter-node)
- **Backend:** Golang + Gin + `database/sql` + MySQL 8
- **Cache & Queue:** Redis
- **Autentikasi:** JWT (access token) + refresh token rotation (SHA-256 hashed)
- **Infra:** Docker Compose, GitHub Actions CI

---

## Status Fase Pengembangan

| Fase       | Deskripsi                                                  | Status          |
| ---------- | ---------------------------------------------------------- | --------------- |
| **Fase 1** | Fondasi monorepo, scaffold awal                            | ✅ Selesai      |
| **Fase 2** | Autentikasi, manajemen cabang & menu, dashboard fungsional | ✅ Selesai      |
| **Fase 3** | POS / manajemen pesanan                                    | ✅ Selesai      |
| **Fase 4** | Integrasi QRIS gateway + webhook                           | ✅ Selesai      |
| **Fase 5** | Kitchen Display System (KDS) real-time                     | ✅ Selesai      |
| **Fase 6** | Laporan & analitik                                         | ✅ Selesai      |
| **Fase 7** | Self-order QR meja, gambar menu, adapter gateway           | ✅ Selesai      |
| **Fase 8** | Manajemen pengguna & pengaturan profil                     | ✅ Selesai      |
| **Fase 9** | Upload gambar menu (lokal)                                 | ✅ Selesai      |

---

## Fase 1 — Fondasi Monorepo ✅

**Branch:** `copilot/initial-monorepo-setup`  
**Status:** Selesai (Draft PR masih terbuka, belum di-merge ke main)

### Yang dikerjakan:

- [x] Inisialisasi monorepo: `apps/web` (SvelteKit) dan `apps/api` (Go + Gin)
- [x] Konfigurasi Docker Compose (MySQL 8, Redis, API, Web)
- [x] Migrasi database 001–009:
  - `001` — organizations, branches, branch_settings
  - `002` — roles, users, user_branch_roles
  - `003` — dining_areas, tables
  - `004` — menu_categories, menu_items, menu_branch_settings, menu_variants, menu_addons
  - `005` — orders, order_items
  - `006` — payments
  - `007` — kitchen_display
  - `008` — audit_logs
  - `009` — seed roles (super_admin, org_admin, branch_manager, cashier, kitchen_staff, waiter)
- [x] Go API: health check, info endpoint, CORS middleware, logger, recovery
- [x] SvelteKit: layout dashboard (sidebar, header), halaman login placeholder
- [x] GitHub Actions CI: build & vet Go, install + check + build SvelteKit
- [x] Dokumentasi: README.md, PRD, .env.example

---

## Fase 2 — Autentikasi, RBAC, Manajemen Cabang & Menu ✅

**Branch:** `copilot/implement-phase-2-authentication-role-management`  
**Target PR:** terhadap `copilot/initial-monorepo-setup`

### Backend (Go + Gin)

#### Autentikasi (`internal/auth/`)

- [x] **JWT access token** (HMAC-SHA256, TTL: 15 menit, dapat dikonfigurasi via `JWT_ACCESS_TOKEN_MINUTES`)
- [x] **Refresh token rotation** — token opaque 32-byte hex, disimpan hashed SHA-256 di database
- [x] **Bcrypt** untuk hashing password user
- [x] Endpoint `POST /api/v1/auth/login`
- [x] Endpoint `POST /api/v1/auth/refresh` (token rotation)
- [x] Endpoint `POST /api/v1/auth/logout` (revoke refresh token)
- [x] Endpoint `GET /api/v1/auth/me` (profil + daftar role user)
- [x] Tidak menyimpan refresh token plaintext; hash SHA-256 sebelum disimpan
- [x] Respons error konsisten tanpa membocorkan detail internal

#### Middleware & Otorisasi (`internal/middleware/`)

- [x] `Authenticate(jwtSecret)` — validasi JWT dari header `Authorization: ******
- [x] `RequireRoles(...)` — middleware pemeriksaan role untuk route terbatas
- [x] Helper `OrgIDFromContext`, `UserIDFromContext` untuk handler

#### Organisasi (`internal/organization/`)

- [x] `GET /api/v1/organization` — data organisasi konteks pengguna yang terautentikasi

#### Cabang (`internal/branch/`)

- [x] `GET /api/v1/branches` — daftar cabang (scoped ke organisasi)
- [x] `POST /api/v1/branches` — buat cabang baru
- [x] `GET /api/v1/branches/:id` — detail cabang
- [x] `PATCH /api/v1/branches/:id` — update sebagian data cabang

#### Menu (`internal/menu/`)

- [x] `GET /api/v1/menu/categories` — daftar kategori
- [x] `POST /api/v1/menu/categories` — buat kategori
- [x] `PATCH /api/v1/menu/categories/:id` — update kategori
- [x] `GET /api/v1/menu/items` — daftar item (filter: `category_id`, `branch_id`, `active`)
- [x] `POST /api/v1/menu/items` — buat item menu
- [x] `GET /api/v1/menu/items/:id` — detail item
- [x] `PATCH /api/v1/menu/items/:id` — update sebagian data item
- [x] `GET /api/v1/menu/items/:id/branches` — daftar setting harga/ketersediaan per-cabang
- [x] `PUT /api/v1/menu/items/:id/branches/:branch_id` — upsert setting per-cabang

#### Migrasi Tambahan

- [x] `010_create_refresh_tokens.sql` — tabel refresh_tokens (token_hash UNIQUE, expires_at, revoked_at)
- [x] `011_alter_user_branch_roles_nullable_branch.sql` — branch_id di user_branch_roles boleh NULL (untuk role global/organisasi)

#### CLI Seed Admin (`cmd/seed/`)

- [x] Perintah Go CLI untuk membuat admin awal secara aman
- [x] Meng-hash password menggunakan bcrypt sebelum disimpan
- [x] Tidak ada kredensial default atau seed user dalam migrasi

#### Unit Tests

- [x] `internal/auth/token_test.go`:
  - Generate dan parse access token
  - Token yang sudah kadaluarsa ditolak
  - Token dengan secret yang salah ditolak
  - Generate refresh token (unik, deterministik)
  - Token malformed ditolak

#### Infrastruktur

- [x] Dependensi baru: `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto` (bcrypt)
- [x] Config: `JWT_ACCESS_TOKEN_MINUTES`, `JWT_REFRESH_TOKEN_DAYS`
- [x] CI: langkah `go test ./...` ditambahkan di workflow
- [x] CI: trigger PR sekarang mencakup branch `copilot/*`

### Frontend (SvelteKit)

#### Autentikasi

- [x] **Login real** via `POST /api/v1/auth/login` menggunakan SvelteKit form action
- [x] **Access token** disimpan di httpOnly cookie (`access_token`, TTL sesuai `expires_in`)
- [x] **Refresh token** disimpan di httpOnly cookie (`refresh_token`, TTL 30 hari)
- [x] **Tidak ada** long-lived token di localStorage
- [x] `hooks.server.ts` — middleware SSR: validasi token, refresh otomatis, proteksi route `/dashboard/*`
- [x] Redirect otomatis dari login jika sudah terautentikasi
- [x] Logout: revoke token di backend, hapus kedua cookie, redirect ke `/`

#### Dashboard

- [x] Dashboard menampilkan data nyata: jumlah cabang, kategori menu, item menu
- [x] Tampilan daftar cabang aktif (preview 3 teratas)
- [x] Kartu modul dengan status aktif/segera-hadir yang akurat

#### Halaman Cabang (`/dashboard/branches`)

- [x] Daftar semua cabang dengan nama, slug, alamat, telepon, status
- [x] Form tambah cabang baru (auto-slug dari nama)
- [x] Form edit cabang (nama, slug, alamat, telepon, status aktif/nonaktif)
- [x] Feedback error/sukses dari server action

#### Halaman Menu (`/dashboard/menu`)

- [x] Tab **Item Menu** dan **Kategori**
- [x] Filter item berdasarkan kategori
- [x] Form tambah kategori baru
- [x] Form tambah item menu baru (pilih kategori, nama, harga, deskripsi)
- [x] Form edit item menu (nama, harga, deskripsi, status)
- [x] Format harga Rupiah (IDR)

#### API Client (`src/lib/api/client.ts`)

- [x] Typed interface: `TokenPair`, `MeResponse`, `Organization`, `Branch`, `MenuCategory`, `MenuItem`, `MenuBranchSetting`
- [x] Fungsi: `api.login`, `api.refresh`, `api.logout`, `api.me`
- [x] Fungsi: `api.organization`, `api.branches.*`, `api.menu.categories.*`, `api.menu.items.*`
- [x] Authorization header diset otomatis dari parameter `token`

---

## Fase 3 — POS & Manajemen Pesanan ✅

**Branch:** `copilot/implement-phase-2-authentication-role-management`

### Backend (Go + Gin)

#### Manajemen Meja (`internal/table/`)

- [x] `GET /api/v1/branches/:branch_id/dining-areas` — daftar area makan
- [x] `POST /api/v1/branches/:branch_id/dining-areas` — buat area makan baru
- [x] `GET /api/v1/branches/:branch_id/tables` — daftar meja per cabang
- [x] `POST /api/v1/branches/:branch_id/tables` — tambah meja baru
- [x] `PATCH /api/v1/branches/:branch_id/tables/:id` — update meja (nomor, kapasitas, status)
- [x] Validasi area makan dan meja termasuk dalam cabang yang benar

#### Manajemen Pesanan (`internal/order/`)

- [x] `GET /api/v1/orders` — daftar pesanan (filter: `branch_id`, `status`)
- [x] `POST /api/v1/orders` — buat pesanan baru dengan item
- [x] `GET /api/v1/orders/:id` — detail pesanan + item
- [x] `PATCH /api/v1/orders/:id/status` — update status pesanan
- [x] Pembuatan dalam transaksi: order + order_items sekaligus
- [x] Resolusi harga: gunakan `price_override` per cabang jika ada, fallback ke `base_price`
- [x] Validasi cabang aktif, meja aktif, dan item menu aktif milik org
- [x] Generate `order_code` unik: `ORD-{YYYYMMDD}-{6 hex chars}`
- [x] Status pesanan: `pending → confirmed → preparing → ready → completed / cancelled`

### Frontend (SvelteKit)

#### Halaman POS (`/dashboard/pos`)

- [x] Pilih cabang aktif
- [x] Pilih tipe pesanan: Makan di Tempat / Takeaway
- [x] Pilih meja (lazy-load daftar meja per cabang saat dibutuhkan)
- [x] Filter item menu berdasarkan kategori
- [x] Grid item menu dengan tombol tambah ke keranjang
- [x] Keranjang: tambah, kurangi item, lihat total
- [x] Catatan pesanan opsional
- [x] Checkout → POST `/api/v1/orders`, tampilkan kode pesanan setelah sukses

#### Halaman Pesanan (`/dashboard/orders`)

- [x] Filter pesanan berdasarkan cabang dan status
- [x] Daftar pesanan dengan status berwarna
- [x] Tombol kemajuan status (Konfirmasi / Mulai Proses / Tandai Siap / Selesaikan)
- [x] Tombol batalkan pesanan (untuk status pending/confirmed)
- [x] Expandable detail pesanan: tabel item, harga satuan, subtotal, total
- [x] Tombol "+ Pesanan Baru" mengarah ke halaman POS

#### API Client (`src/lib/api/client.ts`)

- [x] Type: `DiningArea`, `RestaurantTable`, `Order`, `OrderItem`, `OrderStatus`, `OrderType`, `CreateOrderPayload`
- [x] Fungsi: `api.tables.*` (diningAreas, list, create, update, createDiningArea)
- [x] Fungsi: `api.orders.*` (list, get, create, updateStatus)

---

## Fase 6 — Laporan & Analitik ✅

### Backend (Go + Gin)

- [x] Package `internal/report/` — model, repository, handler
- [x] `GET /api/v1/reports/sales` — ringkasan penjualan (total_revenue, order_count, avg_order_value) + rincian harian; filter: `branch_id`, `date_from`, `date_to`
- [x] `GET /api/v1/reports/top-items` — produk terlaris berdasarkan qty, filter dan limit yang sama
- [x] `GET /api/v1/reports/sales/export` — ekspor penjualan harian ke CSV (UTF-8 BOM agar kompatibel dengan Excel)
- [x] Semua endpoint hanya menghitung pesanan berstatus `completed`

### Frontend (SvelteKit)

- [x] Halaman `/dashboard/reports` dengan filter cabang dan rentang tanggal
- [x] Tiga kartu ringkasan: Total Pendapatan, Jumlah Pesanan, Rata-rata Nilai Pesanan
- [x] Tabel penjualan harian
- [x] Tabel produk terlaris (top 10)
- [x] Tombol ekspor CSV
- [x] Status modul di dashboard diperbarui menjadi "Aktif"

---

## Fase 4 — Integrasi QRIS Gateway & Webhook ✅

**Catatan:** Implementasi menggunakan provider `qris_mock` untuk alur invoice, webhook signature, dan rekonsiliasi. Untuk beralih ke provider produksi (Midtrans, Xendit, dll.), implementasikan interface `Gateway` di `internal/payment/gateway.go` dan daftarkan via `Repository.WithGateway()`.

### Backend (Go + Gin)

- [x] Konfigurasi merchant QRIS per-cabang (credential terenkripsi AES-GCM)
- [x] **Interface `Gateway`** (`internal/payment/gateway.go`) — abstraksi provider QRIS yang dapat ditukar
- [x] **`MockGateway`** (`internal/payment/provider_mock.go`) — implementasi built-in untuk development
- [x] `NewRepository(...).WithGateway(gw)` — registry provider; tambahkan provider produksi tanpa mengubah kode inti
- [x] Endpoint buat invoice QRIS dinamis (`POST /api/v1/orders/:id/payments/qris-invoice`)
- [x] Endpoint webhook dengan validasi signature via `Gateway.NormalizeWebhook()`
- [x] Pencegahan double-payment (idempotency via `gateway_invoice_id`)
- [x] Rekonsiliasi transaksi QRIS (`GET /api/v1/payments/reconciliation`)
- [x] Penanganan status: pending, paid, expired, failed

### Frontend (SvelteKit)

- [x] Tampilkan QR code di halaman POS setelah checkout
- [x] Warning jika invoice QRIS gagal dibuat (pesanan tetap dicatat)
- [x] Konfigurasi gateway QRIS per-cabang via endpoint yang ada

---

## Fase 5 — Kitchen Display System (KDS) ✅

### Backend (Go + Gin)

- [x] Package `internal/kds/` untuk station dapur, ticket dapur, dan event realtime
- [x] WebSocket hub goroutine-safe berbasis `gorilla/websocket` dengan registry client per cabang
- [x] Endpoint `GET /api/v1/branches/:branch_id/kds/stations` — daftar station dapur
- [x] Endpoint `POST /api/v1/branches/:branch_id/kds/stations` — tambah station dapur
- [x] Endpoint `GET /api/v1/branches/:branch_id/kds/tickets` — daftar ticket dapur aktif
- [x] Endpoint `PATCH /api/v1/branches/:branch_id/kds/tickets/:id/status` — update status ticket
- [x] Endpoint `GET /api/v1/kds/ws?branch_id=X&token=Y` — realtime stream ticket dapur
- [x] Auto-create kitchen ticket per `order_item` saat order dikonfirmasi
- [x] Auto-cancel kitchen ticket terkait saat order dibatalkan
- [x] Broadcast event realtime saat ticket baru dibuat atau status ticket berubah
- [x] Unit test untuk subscribe, unsubscribe, dan broadcast pada KDS hub

### Frontend (SvelteKit)

- [x] Halaman `/dashboard/kds` dengan filter cabang dan station
- [x] Tampilan kolom KDS: baru → diproses → siap
- [x] Auto-reconnect WebSocket untuk update ticket real-time
- [x] Highlight visual dan notifikasi audio ringan saat ticket baru masuk
- [x] Form tambah station dapur langsung dari halaman KDS
- [x] Tombol perubahan status ticket dari UI KDS

### Catatan

- [x] Route lama `/dashboard/kitchen` diarahkan ke `/dashboard/kds`
- [x] Schema database existing pada migrasi `007_create_kitchen.sql` digunakan tanpa migrasi baru

---

## Fase 7 — Self-order QR Meja, Gambar Menu, Gateway Adapter ✅

**Branch:** `copilot/implement-phase-2-authentication-role-management`

### Backend (Go + Gin)

#### QRIS Gateway Adapter (`internal/payment/`)

- [x] Interface `Gateway` — abstraksi provider QRIS (invoice + webhook normalization)
- [x] `MockGateway` — implementasi built-in untuk development/testing
- [x] `Repository.WithGateway()` — registry multi-provider; tambahkan Midtrans/Xendit tanpa mengubah kode inti
- [x] `Repository.NormalizeWebhook()` — delegasikan validasi signature + parsing ke provider
- [x] Handler webhook menggunakan `query param branch_id` (tidak lagi bergantung pada body format tertentu)

#### Self-order QR Meja (`internal/selforder/`)

- [x] Migrasi `012_create_qr_table_tokens.sql` — tabel token QR per meja
- [x] `POST /api/v1/branches/:branch_id/tables/:table_id/qr-token` — generate token QR baru (JWT required)
- [x] `GET /api/v1/branches/:branch_id/tables/:table_id/qr-token` — lihat token aktif (JWT required)
- [x] `GET /api/v1/public/table/:token` — menu publik per token (tanpa login)
- [x] `POST /api/v1/public/table/:token/orders` — buat pesanan self-order (tanpa login)
- [x] Token 64-char hex acak, idempotent (deactivate token lama saat generate ulang)
- [x] Harga efektif = override cabang jika ada, fallback ke harga dasar

### Frontend (SvelteKit)

#### Dashboard Meja & QR (`/dashboard/tables`)

- [x] Pilih cabang → tampilkan daftar meja
- [x] Generate QR token per meja (satu klik)
- [x] Tampilkan QR code gambar + URL self-order
- [x] Tombol salin URL dan buka tab baru
- [x] Sidebar: tambah menu item "Meja & QR" (🪑)

#### Halaman Self-order Publik (`/order/[token]`)

- [x] Halaman publik (tidak perlu login), accessible dari QR code
- [x] Tampilkan nama cabang dan nomor meja
- [x] Filter menu per kategori
- [x] Keranjang stiky di bawah layar dengan tombol +/−
- [x] Checkout → POST ke public API, tampilkan kode pesanan sukses

#### Gambar Menu

- [x] Field `image_url` di form tambah dan edit item menu
- [x] Preview thumbnail gambar pada form edit
- [x] POS grid dan self-order menampilkan gambar jika tersedia

---

```bash
# 1. Salin file environment
cp .env.example .env
# Edit .env sesuai kebutuhan lokal Anda

# 2. Jalankan semua service
docker compose up -d

# 3. Jalankan migrasi (jika belum otomatis)
# Lihat instruksi di README.md

# 4. Buat admin pertama
cd apps/api
go run ./cmd/seed \
  --org  "Nama Restoran Anda" \
  --slug "nama-restoran" \
  --name "Admin Utama" \
  --email "admin@restoran.com" \
  --pass  "password_kuat_anda"
# Hindari menyimpan password di shell history (misal awali command dengan spasi
# pada shell yang mendukung HISTCONTROL=ignorespace, atau hapus entri terakhir
# segera setelah selesai dengan `history -d -1`).
```

---

## Fase 8 — Manajemen Pengguna & Pengaturan Profil ✅

**Branch:** `copilot/implement-phase-2-authentication-role-management`

### Backend (Go + Gin)

#### Manajemen Pengguna (`internal/user/`)

- [x] `GET /api/v1/roles` — daftar semua role yang tersedia
- [x] `GET /api/v1/users` — daftar pengguna dalam organisasi (filter: `branch_id`)
- [x] `POST /api/v1/users` — tambah pengguna baru + assign role (org_admin/super_admin only)
- [x] `GET /api/v1/users/:id` — detail pengguna
- [x] `PATCH /api/v1/users/:id` — update nama, email, status aktif (org_admin/super_admin only)
- [x] `DELETE /api/v1/users/:id` — nonaktifkan pengguna / soft delete (org_admin/super_admin only)
- [x] `PUT /api/v1/users/:id/roles` — ganti seluruh role assignment pengguna (org_admin/super_admin only)
- [x] Guard berbasis DB: `IsOrgAdmin()` memvalidasi caller punya role `org_admin` atau `super_admin`

#### Auth — Ganti Password

- [x] `PATCH /api/v1/auth/me/password` — ganti password diri sendiri, validasi password lama dulu
- [x] `ErrWrongPassword` — error khusus untuk password lama tidak cocok

#### Organisasi — Update Profil

- [x] `PATCH /api/v1/organization` — update nama/slug organisasi

### Frontend (SvelteKit)

#### Halaman Pengguna (`/dashboard/users`)

- [x] Tabel pengguna: nama, email, role + cabang, status aktif/nonaktif
- [x] Form tambah pengguna (nama, email, password sementara, pilih role + cabang)
- [x] Form edit pengguna (nama, email, toggle status aktif)
- [x] Form ubah role pengguna
- [x] Tombol nonaktifkan pengguna (dengan konfirmasi)
- [x] Menu sidebar "👥 Pengguna" ditambahkan

#### Halaman Pengaturan (`/dashboard/settings`)

- [x] Kartu info sesi aktif (nama, email, role)
- [x] Form update profil organisasi (nama, slug)
- [x] Form ganti password (validasi password lama, konfirmasi password baru)
- [x] Feedback error/sukses per section (profil & password terpisah)

#### API Client (`src/lib/api/client.ts`)

- [x] Type: `UserWithRoles`, `Role`, `RoleAssignment`, `CreateUserPayload`, `UpdateUserPayload`, `UpdateRolesPayload`
- [x] Fungsi: `api.users.*` (list, get, create, update, deactivate, setRoles)
- [x] Fungsi: `api.roles(token)` — daftar role tersedia
- [x] Fungsi: `api.updateOrganization(token, payload)` — update profil organisasi
- [x] Fungsi: `api.changePassword(token, payload)` — ganti password

---

## Catatan Keamanan

- Jangan commit file `.env` dengan kredensial nyata
- Ganti `JWT_SECRET` dengan string random yang panjang di production
- Set cookie `secure: true` di production (HTTPS)
- Refresh token disimpan sebagai SHA-256 hash — token plaintext hanya dikirim sekali ke client
- Password user di-hash dengan bcrypt (cost default)
- Tidak ada user/password default dalam migrasi

---

## Fase 9 — Upload Gambar Menu ✅

**Branch:** `copilot/implement-phase-2-authentication-role-management`

### Backend (Go + Gin)

#### Config

- [x] `UPLOAD_DIR` — direktori penyimpanan file gambar (default `./uploads`)
- [x] `PUBLIC_BASE_URL` — URL dasar untuk membangun URL publik file (default `http://localhost:8080`)

#### Menu Handler (`internal/menu/`)

- [x] `POST /api/v1/menu/items/:id/image` — upload gambar item menu (multipart/form-data, field `image`)
  - Validasi tipe MIME dari konten file: hanya `image/jpeg`, `image/png`, `image/webp`
  - Batas ukuran: 5 MB
  - Nama file aman: `menu_{id}_{namaasli}.{ext}`
  - Hapus otomatis file lama saat gambar diganti
  - Simpan URL publik di kolom `image_url` di database
- [x] `DELETE /api/v1/menu/items/:id/image` — hapus gambar item menu
  - Set `image_url = NULL` di database
  - Hapus file fisik dari disk
- [x] `WithUpload(uploadDir, publicBaseURL)` — fluent config pada handler
- [x] `Repository.SetItemImage()` — update `image_url` di DB
- [x] `Repository.ClearItemImage()` — clear `image_url` di DB, kembalikan URL lama
- [x] Static serving: `GET /uploads/*` — melayani file gambar yang diunggah

### Frontend (SvelteKit)

#### API Client (`src/lib/api/client.ts`)

- [x] `api.menu.items.uploadImage(token, itemId, file)` — upload file langsung ke API
- [x] `api.menu.items.deleteImage(token, itemId)` — hapus gambar item

#### Halaman Menu (`/dashboard/menu`)

- [x] Form tambah item: input file upload (jpg/png/webp, maks 5 MB) dengan preview thumbnail
- [x] Form edit item: section "Gambar Menu" terpisah dari form update teks
  - Preview gambar saat ini
  - Upload gambar baru (form multipart terpisah)
  - Tombol "Hapus Gambar" (hanya tampil jika gambar ada)
  - Preview gambar baru sebelum diunggah
- [x] Server action `uploadImage` — meneruskan file ke API sebagai multipart
- [x] Server action `deleteImage` — memanggil DELETE endpoint
- [x] Server action `createItem` — membuat item lalu upload gambar jika disertakan (2 langkah)
- [x] Feedback warning jika item berhasil dibuat tapi gambar gagal diunggah

#### Environment

- [x] `.env.example` diperbarui dengan `UPLOAD_DIR` dan `PUBLIC_BASE_URL`
