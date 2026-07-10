-- Migration: 006_create_payments.sql
-- Description: Payment gateway config, payments, and webhook logs (QRIS idempotency)

CREATE TABLE IF NOT EXISTS payment_gateway_configs (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    branch_id       BIGINT UNSIGNED NOT NULL,
    provider        VARCHAR(50)     NOT NULL,   -- e.g. 'qris_midtrans', 'qris_xendit'
    is_active       TINYINT(1)      NOT NULL DEFAULT 0,
    -- Credentials stored as opaque JSON; real secrets must live outside the DB
    -- (e.g., secret manager) and this column holds only key identifiers/references.
    config_json     JSON            NOT NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_pgc_branch_id (branch_id),
    CONSTRAINT fk_pgc_branch FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS payments (
    id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id                BIGINT UNSIGNED NOT NULL,
    gateway_config_id       BIGINT UNSIGNED NULL,
    -- gateway_invoice_id is the external identifier issued by the payment provider.
    -- Unique constraint ensures QRIS idempotency: each invoice can only be recorded once.
    gateway_invoice_id      VARCHAR(200)    NOT NULL,
    payment_method          VARCHAR(50)     NOT NULL DEFAULT 'qris',
    amount                  DECIMAL(12,2)   NOT NULL,
    status                  ENUM('pending','paid','failed','expired','refunded') NOT NULL DEFAULT 'pending',
    paid_at                 DATETIME        NULL,
    qr_code_url             VARCHAR(500)    NULL,
    qr_string               TEXT            NULL,
    expiry_at               DATETIME        NULL,
    created_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_payments_gateway_invoice (gateway_invoice_id),
    KEY idx_payments_order_id (order_id),
    KEY idx_payments_status (status),
    CONSTRAINT fk_payments_order          FOREIGN KEY (order_id)          REFERENCES orders                 (id) ON DELETE RESTRICT,
    CONSTRAINT fk_payments_gateway_config FOREIGN KEY (gateway_config_id) REFERENCES payment_gateway_configs (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Stores every raw webhook event from the payment gateway for audit and replay.
-- gateway_event_id must be unique to guarantee idempotent webhook processing.
CREATE TABLE IF NOT EXISTS payment_webhook_logs (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    payment_id          BIGINT UNSIGNED NULL,
    gateway_event_id    VARCHAR(200)    NOT NULL,
    event_type          VARCHAR(100)    NOT NULL,
    raw_payload         JSON            NOT NULL,
    is_processed        TINYINT(1)      NOT NULL DEFAULT 0,
    processed_at        DATETIME        NULL,
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_webhook_gateway_event (gateway_event_id),
    KEY idx_webhook_payment_id (payment_id),
    KEY idx_webhook_is_processed (is_processed),
    CONSTRAINT fk_webhook_payment FOREIGN KEY (payment_id) REFERENCES payments (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
