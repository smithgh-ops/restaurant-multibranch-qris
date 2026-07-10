-- Migration: 004_create_menu.sql
-- Description: Menu categories, items, branch settings, variants, and addons

CREATE TABLE IF NOT EXISTS menu_categories (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    name            VARCHAR(100)    NOT NULL,
    description     VARCHAR(300)    NULL,
    sort_order      SMALLINT        NOT NULL DEFAULT 0,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_menu_categories_org (organization_id),
    CONSTRAINT fk_menu_categories_org FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS menu_items (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    category_id     BIGINT UNSIGNED NOT NULL,
    name            VARCHAR(150)    NOT NULL,
    description     TEXT            NULL,
    base_price      DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    image_url       VARCHAR(500)    NULL,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_menu_items_org (organization_id),
    KEY idx_menu_items_category (category_id),
    CONSTRAINT fk_menu_items_org      FOREIGN KEY (organization_id) REFERENCES organizations   (id) ON DELETE CASCADE,
    CONSTRAINT fk_menu_items_category FOREIGN KEY (category_id)     REFERENCES menu_categories (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Per-branch overrides for menu item availability and price
CREATE TABLE IF NOT EXISTS menu_branch_settings (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    menu_item_id    BIGINT UNSIGNED NOT NULL,
    branch_id       BIGINT UNSIGNED NOT NULL,
    price_override  DECIMAL(12,2)   NULL,
    is_available    TINYINT(1)      NOT NULL DEFAULT 1,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_menu_branch (menu_item_id, branch_id),
    KEY idx_mbs_branch_id (branch_id),
    CONSTRAINT fk_mbs_menu_item FOREIGN KEY (menu_item_id) REFERENCES menu_items (id) ON DELETE CASCADE,
    CONSTRAINT fk_mbs_branch    FOREIGN KEY (branch_id)    REFERENCES branches   (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS menu_variants (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    menu_item_id    BIGINT UNSIGNED NOT NULL,
    name            VARCHAR(100)    NOT NULL,
    price_modifier  DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_menu_variants_item (menu_item_id),
    CONSTRAINT fk_menu_variants_item FOREIGN KEY (menu_item_id) REFERENCES menu_items (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS menu_addons (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    menu_item_id    BIGINT UNSIGNED NOT NULL,
    name            VARCHAR(100)    NOT NULL,
    price           DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_menu_addons_item (menu_item_id),
    CONSTRAINT fk_menu_addons_item FOREIGN KEY (menu_item_id) REFERENCES menu_items (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
