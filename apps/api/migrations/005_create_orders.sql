-- Migration: 005_create_orders.sql
-- Description: Orders and order items with variant/addon line items

CREATE TABLE IF NOT EXISTS orders (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    branch_id       BIGINT UNSIGNED NOT NULL,
    table_id        BIGINT UNSIGNED NULL,  -- NULL for takeaway/delivery
    order_code      VARCHAR(30)     NOT NULL,
    order_type      ENUM('dine_in','takeaway','delivery') NOT NULL DEFAULT 'dine_in',
    status          ENUM('pending','confirmed','preparing','ready','completed','cancelled') NOT NULL DEFAULT 'pending',
    notes           TEXT            NULL,
    subtotal        DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    tax_amount      DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    service_charge  DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    total_amount    DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    created_by      BIGINT UNSIGNED NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_order_code (order_code),
    KEY idx_orders_branch_id (branch_id),
    KEY idx_orders_table_id (table_id),
    KEY idx_orders_status (status),
    KEY idx_orders_created_at (created_at),
    CONSTRAINT fk_orders_branch FOREIGN KEY (branch_id) REFERENCES branches           (id) ON DELETE RESTRICT,
    CONSTRAINT fk_orders_table  FOREIGN KEY (table_id)  REFERENCES restaurant_tables  (id) ON DELETE SET NULL,
    CONSTRAINT fk_orders_user   FOREIGN KEY (created_by) REFERENCES users             (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_items (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id        BIGINT UNSIGNED NOT NULL,
    menu_item_id    BIGINT UNSIGNED NOT NULL,
    variant_id      BIGINT UNSIGNED NULL,
    item_name       VARCHAR(150)    NOT NULL,  -- snapshot at order time
    unit_price      DECIMAL(12,2)   NOT NULL,
    quantity        SMALLINT UNSIGNED NOT NULL DEFAULT 1,
    subtotal        DECIMAL(12,2)   NOT NULL,
    notes           VARCHAR(300)    NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_order_items_order_id (order_id),
    KEY idx_order_items_menu_item_id (menu_item_id),
    CONSTRAINT fk_order_items_order     FOREIGN KEY (order_id)     REFERENCES orders        (id) ON DELETE CASCADE,
    CONSTRAINT fk_order_items_menu_item FOREIGN KEY (menu_item_id) REFERENCES menu_items    (id) ON DELETE RESTRICT,
    CONSTRAINT fk_order_items_variant   FOREIGN KEY (variant_id)   REFERENCES menu_variants (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_item_addons (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_item_id   BIGINT UNSIGNED NOT NULL,
    addon_id        BIGINT UNSIGNED NOT NULL,
    addon_name      VARCHAR(100)    NOT NULL,  -- snapshot
    unit_price      DECIMAL(12,2)   NOT NULL,
    quantity        TINYINT UNSIGNED NOT NULL DEFAULT 1,
    subtotal        DECIMAL(12,2)   NOT NULL,
    PRIMARY KEY (id),
    KEY idx_oia_order_item (order_item_id),
    CONSTRAINT fk_oia_order_item FOREIGN KEY (order_item_id) REFERENCES order_items  (id) ON DELETE CASCADE,
    CONSTRAINT fk_oia_addon      FOREIGN KEY (addon_id)      REFERENCES menu_addons  (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
