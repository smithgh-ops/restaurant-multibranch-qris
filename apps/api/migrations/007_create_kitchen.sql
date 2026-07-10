-- Migration: 007_create_kitchen.sql
-- Description: Kitchen stations and kitchen tickets

CREATE TABLE IF NOT EXISTS kitchen_stations (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    branch_id   BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(100)    NOT NULL,
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_kitchen_stations_branch (branch_id),
    CONSTRAINT fk_kitchen_stations_branch FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS kitchen_tickets (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id        BIGINT UNSIGNED NOT NULL,
    station_id      BIGINT UNSIGNED NULL,
    order_item_id   BIGINT UNSIGNED NOT NULL,
    status          ENUM('queued','in_progress','done','cancelled') NOT NULL DEFAULT 'queued',
    priority        TINYINT UNSIGNED NOT NULL DEFAULT 5,
    started_at      DATETIME        NULL,
    completed_at    DATETIME        NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_kitchen_tickets_order (order_id),
    KEY idx_kitchen_tickets_station (station_id),
    KEY idx_kitchen_tickets_status (status),
    CONSTRAINT fk_kt_order      FOREIGN KEY (order_id)      REFERENCES orders           (id) ON DELETE CASCADE,
    CONSTRAINT fk_kt_station    FOREIGN KEY (station_id)    REFERENCES kitchen_stations (id) ON DELETE SET NULL,
    CONSTRAINT fk_kt_order_item FOREIGN KEY (order_item_id) REFERENCES order_items      (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
