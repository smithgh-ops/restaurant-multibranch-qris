// Command seed creates an initial super-admin user for the RestoQRIS platform.
//
// Usage:
//
//	go run ./cmd/seed \
//	  --org    "Nama Organisasi" \
//	  --slug   "nama-org" \
//	  --name   "Admin Nama" \
//	  --email  "admin@example.com" \
//	  --pass   "password_kuat_anda"
//
// Environment variables DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME must be set
// (or rely on defaults from config.Load).
//
// This command is intentionally not wired into the main API binary; run it
// once during environment bootstrap and do NOT commit passwords to source control.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/config"
)

func main() {
	orgName := flag.String("org", "", "Nama organisasi (wajib)")
	orgSlug := flag.String("slug", "", "Slug organisasi, huruf kecil + tanda hubung (wajib)")
	name := flag.String("name", "", "Nama lengkap admin (wajib)")
	email := flag.String("email", "", "Alamat email admin (wajib)")
	pass := flag.String("pass", "", "Password admin plaintext (wajib, hanya digunakan saat seed)")
	flag.Parse()

	if *orgName == "" || *orgSlug == "" || *name == "" || *email == "" || *pass == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.Load()
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("ping db", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password", "err", err)
		os.Exit(1)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin tx", "err", err)
		os.Exit(1)
	}
	defer tx.Rollback() //nolint:errcheck

	// Upsert organization
	var orgID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM organizations WHERE slug = ? LIMIT 1`, *orgSlug).Scan(&orgID)
	if err == sql.ErrNoRows {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO organizations (name, slug) VALUES (?, ?)`,
			*orgName, strings.ToLower(*orgSlug),
		)
		if err != nil {
			slog.Error("insert org", "err", err)
			os.Exit(1)
		}
		orgID, _ = res.LastInsertId()
		slog.Info("organisasi dibuat", "id", orgID, "slug", *orgSlug)
	} else if err != nil {
		slog.Error("check org", "err", err)
		os.Exit(1)
	} else {
		slog.Info("organisasi sudah ada", "id", orgID)
	}

	// Check if email already exists
	var existingID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, *email).Scan(&existingID)
	if err == nil {
		slog.Warn("user sudah ada dengan email ini", "id", existingID, "email", *email)
		os.Exit(0)
	} else if err != sql.ErrNoRows {
		slog.Error("check user", "err", err)
		os.Exit(1)
	}

	// Insert user
	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (organization_id, name, email, password_hash) VALUES (?, ?, ?, ?)`,
		orgID, *name, *email, string(hash),
	)
	if err != nil {
		slog.Error("insert user", "err", err)
		os.Exit(1)
	}
	userID, _ := res.LastInsertId()
	slog.Info("user dibuat", "id", userID, "email", *email)

	// Find or get super_admin role id
	var roleID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'super_admin' LIMIT 1`).Scan(&roleID); err != nil {
		slog.Error("find super_admin role (pastikan migration 009 sudah dijalankan)", "err", err)
		os.Exit(1)
	}

	// Assign global super_admin role (branch_id = NULL)
	_, err = tx.ExecContext(ctx,
		`INSERT IGNORE INTO user_branch_roles (user_id, branch_id, role_id) VALUES (?, NULL, ?)`,
		userID, roleID,
	)
	if err != nil {
		slog.Error("assign role", "err", err)
		os.Exit(1)
	}

	if err := tx.Commit(); err != nil {
		slog.Error("commit", "err", err)
		os.Exit(1)
	}

	slog.Info("✓ seed selesai", "org_id", orgID, "user_id", userID, "email", *email, "role", "super_admin")
	fmt.Printf("\nAdmin berhasil dibuat:\n  Email : %s\n  User ID: %d\n\nHapus password dari shell history Anda setelah selesai.\n", *email, userID)
}
