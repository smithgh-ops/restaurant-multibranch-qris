-- Migration: 001_create_organizations_and_branches.sql
-- Description: Core organization and branch structure

CREATE TABLE IF NOT EXISTS organizations (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name        VARCHAR(100)    NOT NULL,
    slug        VARCHAR(100)    NOT NULL,
    logo_url    VARCHAR(500)    NULL,
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_organizations_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS branches (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    name            VARCHAR(100)    NOT NULL,
    slug            VARCHAR(100)    NOT NULL,
    address         TEXT            NULL,
    phone           VARCHAR(20)     NULL,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_branches_slug (slug),
    KEY idx_branches_organization_id (organization_id),
    CONSTRAINT fk_branches_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS branch_settings (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    branch_id       BIGINT UNSIGNED NOT NULL,
    setting_key     VARCHAR(100)    NOT NULL,
    setting_value   TEXT            NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_branch_settings_key (branch_id, setting_key),
    CONSTRAINT fk_branch_settings_branch FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
