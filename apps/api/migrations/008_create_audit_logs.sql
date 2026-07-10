-- Migration: 008_create_audit_logs.sql
-- Description: Audit trail for important system events

CREATE TABLE IF NOT EXISTS audit_logs (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id         BIGINT UNSIGNED NULL,
    branch_id       BIGINT UNSIGNED NULL,
    entity_type     VARCHAR(50)     NOT NULL,
    entity_id       BIGINT UNSIGNED NULL,
    action          VARCHAR(50)     NOT NULL,  -- e.g. 'create', 'update', 'delete', 'login'
    old_values      JSON            NULL,
    new_values      JSON            NULL,
    ip_address      VARCHAR(45)     NULL,
    user_agent      VARCHAR(300)    NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_audit_logs_user (user_id),
    KEY idx_audit_logs_branch (branch_id),
    KEY idx_audit_logs_entity (entity_type, entity_id),
    KEY idx_audit_logs_created_at (created_at),
    CONSTRAINT fk_audit_logs_user   FOREIGN KEY (user_id)   REFERENCES users    (id) ON DELETE SET NULL,
    CONSTRAINT fk_audit_logs_branch FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
