-- Migration: 003_create_dining_areas_and_tables.sql
-- Description: Dining areas, restaurant tables, and QR table tokens

CREATE TABLE IF NOT EXISTS dining_areas (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    branch_id   BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(100)    NOT NULL,
    description VARCHAR(200)    NULL,
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_dining_areas_branch_id (branch_id),
    CONSTRAINT fk_dining_areas_branch FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS restaurant_tables (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    dining_area_id  BIGINT UNSIGNED NOT NULL,
    branch_id       BIGINT UNSIGNED NOT NULL,
    table_number    VARCHAR(20)     NOT NULL,
    capacity        TINYINT UNSIGNED NOT NULL DEFAULT 4,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_table_per_area (dining_area_id, table_number),
    KEY idx_restaurant_tables_branch_id (branch_id),
    CONSTRAINT fk_tables_dining_area FOREIGN KEY (dining_area_id) REFERENCES dining_areas (id) ON DELETE CASCADE,
    CONSTRAINT fk_tables_branch      FOREIGN KEY (branch_id)      REFERENCES branches     (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS qr_table_tokens (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    table_id     BIGINT UNSIGNED NOT NULL,
    token        VARCHAR(64)     NOT NULL,
    is_active    TINYINT(1)      NOT NULL DEFAULT 1,
    expires_at   DATETIME        NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_qr_token (token),
    KEY idx_qr_tokens_table_id (table_id),
    CONSTRAINT fk_qr_tokens_table FOREIGN KEY (table_id) REFERENCES restaurant_tables (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
