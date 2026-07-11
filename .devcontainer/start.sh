#!/usr/bin/env bash
# start.sh — Dijalankan SETIAP KALI Codespace dinyalakan.
# Tugasnya: mulai MySQL, Redis, Go API, dan SvelteKit dev server.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║   RestoQRIS — Memulai Layanan            ║"
echo "╚══════════════════════════════════════════╝"
echo ""

# ── 1. Muat env vars ─────────────────────────────────────────────────────────
if [ -f .env ]; then
  set -a && source .env && set +a
fi

# ── 2. Mulai infrastruktur (idempoten) ───────────────────────────────────────
docker compose up -d mysql redis
echo "✓ MySQL & Redis berjalan"

# ── 3. Tunggu MySQL siap ─────────────────────────────────────────────────────
echo "  Menunggu MySQL..."
for i in $(seq 1 20); do
  if docker compose exec -T mysql \
       mysqladmin ping -h localhost \
       -u root -p"${MYSQL_ROOT_PASSWORD:-rootpassword}" \
       --silent 2>/dev/null; then
    echo "✓ MySQL siap"
    break
  fi
  sleep 3
done

# ── 4. Tentukan URL API untuk browser ────────────────────────────────────────
# Di GitHub Codespaces, browser tidak bisa mengakses localhost secara langsung.
# Gunakan URL publik Codespaces agar VITE_API_BASE_URL bisa dijangkau dari browser.
if [ -n "${CODESPACE_NAME:-}" ] && [ -n "${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN:-}" ]; then
  API_URL="https://${CODESPACE_NAME}-8080.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN}"
  echo "✓ Mode Codespaces — API URL: ${API_URL}"
else
  API_URL="http://localhost:8080"
  echo "✓ Mode lokal — API URL: ${API_URL}"
fi

# Tulis ke .env.local agar Vite dev server membacanya saat startup.
echo "VITE_API_BASE_URL=${API_URL}" > apps/web/.env.local

# ── 5. Mulai Go API ──────────────────────────────────────────────────────────
# Hentikan instance lama jika ada.
pkill -f 'apps/api/.devbuild/api' 2>/dev/null || true
sleep 1

nohup apps/api/.devbuild/api > /tmp/resto-api.log 2>&1 &
API_PID=$!
echo "✓ Go API dimulai (PID ${API_PID})"

# ── 6. Tunggu API siap ────────────────────────────────────────────────────────
echo "  Menunggu API merespons di :8080..."
for i in $(seq 1 20); do
  if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ Go API siap"
    break
  fi
  if [ "$i" -eq 20 ]; then
    echo "⚠ API belum merespons. Cek log: tail -f /tmp/resto-api.log"
  fi
  sleep 2
done

# ── 7. Mulai SvelteKit dev server ─────────────────────────────────────────────
pkill -f 'vite dev' 2>/dev/null || true
sleep 1

cd apps/web
nohup npm run dev -- --host 0.0.0.0 > /tmp/resto-web.log 2>&1 &
WEB_PID=$!
echo "✓ SvelteKit dev server dimulai (PID ${WEB_PID})"
cd "$ROOT"

# ── Selesai ───────────────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════╗"
echo "║   ✓ Semua layanan berjalan!              ║"
echo "╠══════════════════════════════════════════╣"
echo "║   Web : port 5173 (buka dari tab Ports)  ║"
echo "║   API : port 8080                        ║"
echo "╠══════════════════════════════════════════╣"
echo "║   Login demo:                            ║"
echo "║     Email : admin@demo.com               ║"
echo "║     Pass  : demo12345                    ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "  Log API  : tail -f /tmp/resto-api.log"
echo "  Log Web  : tail -f /tmp/resto-web.log"
echo ""
