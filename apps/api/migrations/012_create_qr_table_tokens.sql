-- Migration: 012_create_qr_table_tokens.sql
-- Description: QR tokens per restaurant table for self-ordering (no login required)

CREATE TABLE IF NOT EXISTS qr_table_tokens (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    table_id    BIGINT UNSIGNED NOT NULL,
    branch_id   BIGINT UNSIGNED NOT NULL,
    -- token is a cryptographically random 32-byte hex string (64 chars) embedded in the QR code URL.
    token       VARCHAR(64)     NOT NULL,
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    -- expires_at NULL means the token never expires (rotate manually via the dashboard).
    expires_at  DATETIME        NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_qr_table_tokens_token (token),
    KEY idx_qr_table_tokens_table (table_id),
    KEY idx_qr_table_tokens_branch (branch_id),
    CONSTRAINT fk_qrtk_table  FOREIGN KEY (table_id)  REFERENCES restaurant_tables (id) ON DELETE CASCADE,
    CONSTRAINT fk_qrtk_branch FOREIGN KEY (branch_id) REFERENCES branches           (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
