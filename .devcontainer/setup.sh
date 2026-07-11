#!/usr/bin/env bash
# setup.sh — Dijalankan SEKALI saat Codespace pertama kali dibuat.
# Tugasnya: siapkan .env, instal dependensi, build API, dan seed demo admin.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║   RestoQRIS — Inisialisasi Codespace     ║"
echo "╚══════════════════════════════════════════╝"
echo ""

# ── 1. Siapkan .env ──────────────────────────────────────────────────────────
if [ ! -f .env ]; then
  cp .env.example .env
  echo "✓ .env dibuat dari .env.example"
else
  echo "✓ .env sudah ada"
fi

# Muat env vars
set -a && source .env && set +a

# ── 2. Mulai layanan infrastruktur ───────────────────────────────────────────
echo "  Memulai MySQL & Redis..."
docker compose up -d mysql redis
echo "✓ MySQL & Redis dimulai"

# ── 3. Tunggu MySQL siap ─────────────────────────────────────────────────────
echo "  Menunggu MySQL siap (maks 90 detik)..."
for i in $(seq 1 30); do
  if docker compose exec -T mysql \
       mysqladmin ping -h localhost \
       -u root -p"${MYSQL_ROOT_PASSWORD:-rootpassword}" \
       --silent 2>/dev/null; then
    echo "✓ MySQL siap"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "✗ MySQL tidak merespons setelah 90 detik. Cek: docker compose logs mysql"
    exit 1
  fi
  sleep 3
done

# ── 4. Build Go API ──────────────────────────────────────────────────────────
echo "  Download Go dependencies..."
cd apps/api
go mod download
mkdir -p .devbuild
echo "  Build Go API binary..."
go build -o .devbuild/api ./cmd/api
echo "✓ Go API berhasil di-build"
cd "$ROOT"

# ── 5. Install Node.js dependencies ─────────────────────────────────────────
echo "  Install Node.js dependencies..."
cd apps/web && npm ci --silent && cd "$ROOT"
echo "✓ Node.js deps terpasang"

# ── 6. Seed demo admin ───────────────────────────────────────────────────────
echo "  Membuat akun admin demo..."
cd apps/api
go run ./cmd/seed \
  --org  "Demo Restoran" \
  --slug "demo-restoran" \
  --name "Admin Demo" \
  --email "admin@demo.com" \
  --pass  "demo12345" 2>/dev/null && echo "✓ Admin demo dibuat" || echo "ℹ Admin demo sudah ada"
cd "$ROOT"

# ── Selesai ───────────────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════╗"
echo "║   ✓ Setup selesai!                       ║"
echo "╠══════════════════════════════════════════╣"
echo "║   Login demo:                            ║"
echo "║     Email : admin@demo.com               ║"
echo "║     Pass  : demo12345                    ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "  App akan otomatis berjalan setelah ini."
echo "  Port 5173 akan terbuka di browser secara otomatis."
echo ""
