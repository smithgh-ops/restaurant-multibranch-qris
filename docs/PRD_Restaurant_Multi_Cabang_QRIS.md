# PRD — Aplikasi Manajemen Restoran Multi-Cabang dengan QRIS

**Versi:** 1.0-draft  
**Tanggal:** Juli 2025  
**Repositori:** `smithgh-ops/restaurant-multibranch-qris`

---

## 1. Latar Belakang & Tujuan

Restoran modern dengan lebih dari satu cabang menghadapi tantangan dalam menyinkronkan operasional—mulai dari pengelolaan menu, pencatatan pesanan, pembayaran digital, hingga pelaporan terpusat. Proyek ini membangun platform berbasis web yang menjadi tulang punggung operasional seluruh cabang dengan dukungan pembayaran QRIS (Quick Response Code Indonesian Standard).

**Tujuan utama:**

- Satu platform untuk mengelola semua cabang dari satu dasbor.
- POS (Point of Sale) berbasis web yang dapat diakses dari tablet atau desktop.
- Integrasi pembayaran QRIS yang idempoten dan dapat diperluas ke berbagai gateway.
- Kitchen Display System (KDS) real-time untuk dapur.
- Laporan penjualan per-cabang dan keseluruhan.

---

## 2. Pengguna & Peran

| Peran             | Akses Utama                                              |
|-------------------|----------------------------------------------------------|
| `super_admin`     | Seluruh platform lintas organisasi                       |
| `org_admin`       | Seluruh cabang dalam satu organisasi                     |
| `branch_manager`  | Manajemen operasional satu cabang                        |
| `cashier`         | POS dan pembayaran di cabang tertentu                    |
| `kitchen_staff`   | Tampilan KDS di cabang tertentu                          |
| `waiter`          | Ambil pesanan dan kelola meja                            |

---

## 3. Modul MVP

### 3.1 Manajemen Organisasi & Cabang
- CRUD organisasi dan cabang.
- Pengaturan per-cabang (pajak, service charge, jam operasional).
- Pengelolaan area makan dan meja.
- Pembuatan token QR per-meja untuk self-ordering.

### 3.2 Manajemen Pengguna & Role
- Registrasi dan autentikasi pengguna (JWT).
- Penugasan peran per-cabang (satu pengguna dapat memiliki peran berbeda di cabang berbeda).
- Audit log untuk semua perubahan sensitif.

### 3.3 Manajemen Menu
- CRUD kategori, item menu, varian, dan add-on.
- Override harga dan ketersediaan per-cabang.
- Upload gambar item.

### 3.4 Point of Sale (POS)
- Buat pesanan untuk dine-in (pilih meja), takeaway, atau delivery.
- Tambah/edit item, varian, add-on, dan catatan.
- Tampilkan ringkasan tagihan dengan pajak dan service charge.
- Proses pembayaran via QRIS atau metode lain.

### 3.5 Manajemen Pesanan
- Daftar pesanan aktif dan riwayat.
- Filter berdasarkan status, tanggal, dan cabang.
- Update status pesanan (confirmed → preparing → ready → completed).

### 3.6 Kitchen Display System (KDS)
- Tampilan real-time tiket dapur per-stasiun.
- Update status tiket (queued → in_progress → done).
- Notifikasi pesanan baru via WebSocket.

### 3.7 Pembayaran QRIS
- Generate QRIS invoice melalui payment gateway (interface/adapter pattern).
- Terima webhook verifikasi dari gateway sebagai sumber kebenaran status pembayaran.
- Idempotency: `gateway_invoice_id` dan `gateway_event_id` dijamin unik di database.
- Log semua webhook mentah untuk keperluan audit dan replay.

### 3.8 Laporan & Analitik
- Ringkasan penjualan harian/mingguan/bulanan per-cabang.
- Produk terlaris dan jam puncak.
- Ekspor laporan ke CSV.

---

## 4. Arsitektur Sistem

```
┌──────────────────────────────────────────────────────┐
│                    Browser / PWA                     │
│              SvelteKit + TypeScript + Tailwind       │
└──────────────────────┬───────────────────────────────┘
                       │ HTTP / WebSocket
┌──────────────────────▼───────────────────────────────┐
│                  Go Gin API Server                   │
│   /health · /api/v1/{auth,branches,menus,orders,     │
│              payments,kitchen,reports}               │
│              Structured logging (slog)               │
└───────┬───────────────────────┬──────────────────────┘
        │                       │
┌───────▼───────┐     ┌─────────▼─────────┐
│   MySQL 8     │     │    Redis 7         │
│  (Persistent  │     │  (Session cache,   │
│   Data Store) │     │   job queue)       │
└───────────────┘     └────────────────────┘
        │
┌───────▼──────────────────────┐
│   Payment Gateway (adapter)  │
│   QRIS provider (future)     │
└──────────────────────────────┘
```

**Stack:**
- Frontend: SvelteKit 2, TypeScript, Tailwind CSS 3
- Backend: Go 1.24, Gin 1.10
- Database: MySQL 8.0
- Cache: Redis 7
- Kontainerisasi: Docker Compose

---

## 5. Model Data (Ringkasan)

| Tabel                     | Deskripsi                                            |
|---------------------------|------------------------------------------------------|
| `organizations`           | Entitas organisasi/brand restoran                    |
| `branches`                | Cabang-cabang restoran                               |
| `branch_settings`         | Konfigurasi per-cabang                               |
| `roles`                   | Peran sistem                                         |
| `users`                   | Akun pengguna                                        |
| `user_branch_roles`       | Relasi many-to-many user ↔ branch ↔ role             |
| `dining_areas`            | Area makan per-cabang                                |
| `restaurant_tables`       | Meja fisik                                           |
| `qr_table_tokens`         | Token QR per-meja untuk self-order                   |
| `menu_categories`         | Kategori menu tingkat organisasi                     |
| `menu_items`              | Item menu dengan harga dasar                         |
| `menu_branch_settings`    | Override harga/ketersediaan per-cabang               |
| `menu_variants`           | Varian item (ukuran, suhu, dll.)                     |
| `menu_addons`             | Add-on item (topping, ekstra, dll.)                  |
| `orders`                  | Header pesanan                                       |
| `order_items`             | Baris pesanan (snapshot harga saat order)            |
| `order_item_addons`       | Add-on per baris pesanan                             |
| `payment_gateway_configs` | Konfigurasi gateway per-cabang                       |
| `payments`                | Record pembayaran (unique: `gateway_invoice_id`)     |
| `payment_webhook_logs`    | Log webhook mentah (unique: `gateway_event_id`)      |
| `kitchen_stations`        | Stasiun dapur per-cabang                             |
| `kitchen_tickets`         | Tiket masak per-item pesanan                         |
| `audit_logs`              | Log perubahan penting                                |

---

## 6. Keamanan & Kepatuhan

- Autentikasi JWT; secret harus di-set via environment variable, bukan di kode.
- Password di-hash menggunakan bcrypt.
- Kredensial gateway pembayaran **tidak** disimpan sebagai teks biasa; gunakan referensi ke secret manager pada produksi.
- Webhook pembayaran diverifikasi menggunakan signature header dari gateway (belum diimplementasi; siapkan interface).
- CORS dikonfigurasi eksplisit per-environment.
- Semua tabel menggunakan `utf8mb4` untuk dukungan penuh Unicode.

---

## 7. Batasan MVP & Pekerjaan Mendatang

| Item                          | Status MVP       | Catatan                                         |
|-------------------------------|------------------|-------------------------------------------------|
| Autentikasi JWT               | Belum (scaffold) | Endpoint auth akan ditambahkan di sprint berikut |
| Integrasi gateway QRIS nyata  | Belum            | Adapter interface sudah disiapkan               |
| WebSocket KDS                 | Belum            | Struktur paket `ws` disiapkan                   |
| Upload gambar CDN             | Belum            | Gunakan URL eksternal sementara                 |
| Multi-tenancy penuh           | Parsial          | Isolasi organisasi ada; billing belum           |
| PWA / mobile app              | Belum            | SvelteKit mendukung PWA dengan plugin tambahan  |
| Delivery / third-party OTA    | Belum            | Dalam scope roadmap                             |

---

## 8. Definisi Selesai (DoD) untuk Sprint Pertama

- [ ] Repository monorepo dapat di-clone dan dijalankan dengan `docker compose up`.
- [ ] `GET /health` mengembalikan `{"status":"ok"}`.
- [ ] `GET /api/v1/info` mengembalikan nama aplikasi dan environment.
- [ ] Semua migrasi database dijalankan otomatis oleh MySQL container saat pertama kali.
- [ ] Frontend menampilkan halaman login dan dashboard shell.
- [ ] CI (GitHub Actions) build dan check berhasil pada push ke `main`.
