package kds

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Repository handles DB operations for kitchen stations and tickets.
type Repository struct {
	db *sql.DB
}

type dbtx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// NewRepository creates a new KDS repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func scanStation(row interface{ Scan(...any) error }) (*KitchenStation, error) {
	station := &KitchenStation{}
	err := row.Scan(
		&station.ID,
		&station.BranchID,
		&station.Name,
		&station.IsActive,
		&station.CreatedAt,
		&station.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return station, nil
}

func scanTicket(row interface{ Scan(...any) error }) (*KitchenTicket, error) {
	ticket := &KitchenTicket{}
	var stationID sql.NullInt64
	var stationName sql.NullString
	var tableID sql.NullInt64
	var notes sql.NullString
	var startedAt sql.NullTime
	var completedAt sql.NullTime

	err := row.Scan(
		&ticket.ID,
		&ticket.OrderID,
		&stationID,
		&stationName,
		&ticket.OrderItemID,
		&ticket.BranchID,
		&tableID,
		&ticket.OrderCode,
		&ticket.ItemName,
		&ticket.Quantity,
		&notes,
		&ticket.Status,
		&ticket.Priority,
		&startedAt,
		&completedAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if stationID.Valid {
		value := uint64(stationID.Int64)
		ticket.StationID = &value
	}
	if stationName.Valid {
		ticket.StationName = &stationName.String
	}
	if tableID.Valid {
		value := uint64(tableID.Int64)
		ticket.TableID = &value
	}
	if notes.Valid {
		ticket.Notes = &notes.String
	}
	if startedAt.Valid {
		value := startedAt.Time
		ticket.StartedAt = &value
	}
	if completedAt.Valid {
		value := completedAt.Time
		ticket.CompletedAt = &value
	}
	return ticket, nil
}

const ticketSelect = `
	SELECT kt.id, kt.order_id, kt.station_id, ks.name, kt.order_item_id,
	       o.branch_id, o.table_id, o.order_code, oi.item_name, oi.quantity,
	       oi.notes, kt.status, kt.priority, kt.started_at, kt.completed_at,
	       kt.created_at, kt.updated_at
	FROM kitchen_tickets kt
	JOIN orders o ON o.id = kt.order_id
	JOIN order_items oi ON oi.id = kt.order_item_id
	LEFT JOIN kitchen_stations ks ON ks.id = kt.station_id`

// ListStations returns all stations for a branch.
func (r *Repository) ListStations(ctx context.Context, branchID uint64) ([]KitchenStation, error) {
	const q = `
		SELECT id, branch_id, name, is_active, created_at, updated_at
		FROM kitchen_stations
		WHERE branch_id = ?
		ORDER BY is_active DESC, name ASC`
	rows, err := r.db.QueryContext(ctx, q, branchID)
	if err != nil {
		return nil, fmt.Errorf("list stations: %w", err)
	}
	defer rows.Close()

	stations := make([]KitchenStation, 0)
	for rows.Next() {
		station, err := scanStation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan station: %w", err)
		}
		stations = append(stations, *station)
	}
	return stations, rows.Err()
}

// FindStationByID returns a station scoped to a branch.
func (r *Repository) FindStationByID(ctx context.Context, id, branchID uint64) (*KitchenStation, error) {
	const q = `
		SELECT id, branch_id, name, is_active, created_at, updated_at
		FROM kitchen_stations
		WHERE id = ? AND branch_id = ?
		LIMIT 1`
	return scanStation(r.db.QueryRowContext(ctx, q, id, branchID))
}

// CreateStation inserts a new kitchen station.
func (r *Repository) CreateStation(ctx context.Context, branchID uint64, req *CreateStationRequest) (*KitchenStation, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO kitchen_stations (branch_id, name) VALUES (?, ?)`,
		branchID, strings.TrimSpace(req.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("create station: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindStationByID(ctx, uint64(id), branchID)
}

// ListTickets returns kitchen tickets for a branch filtered by station and status.
func (r *Repository) ListTickets(ctx context.Context, branchID uint64, filter TicketListFilter) ([]KitchenTicket, error) {
	query := ticketSelect + `
		 WHERE o.branch_id = ?`
	args := []any{branchID}

	if filter.StationID != nil {
		query += ` AND kt.station_id = ?`
		args = append(args, *filter.StationID)
	}
	if filter.Status != nil {
		query += ` AND kt.status = ?`
		args = append(args, *filter.Status)
	} else {
		query += ` AND kt.status IN ('queued', 'in_progress', 'done')`
	}
	query += ` ORDER BY FIELD(kt.status, 'queued', 'in_progress', 'done', 'cancelled'), kt.priority ASC, kt.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]KitchenTicket, 0)
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, *ticket)
	}
	return tickets, rows.Err()
}

// FindTicketByID returns a ticket scoped to a branch.
func (r *Repository) FindTicketByID(ctx context.Context, id, branchID uint64) (*KitchenTicket, error) {
	row := r.db.QueryRowContext(ctx, ticketSelect+`
		 WHERE kt.id = ? AND o.branch_id = ?
		 LIMIT 1`, id, branchID)
	return scanTicket(row)
}

func findTicketByIDWithQuerier(ctx context.Context, querier dbtx, id uint64) (*KitchenTicket, error) {
	row := querier.QueryRowContext(ctx, ticketSelect+`
		 WHERE kt.id = ?
		 LIMIT 1`, id)
	return scanTicket(row)
}

// UpdateTicketStatus updates the status of a kitchen ticket.
func (r *Repository) UpdateTicketStatus(ctx context.Context, id, branchID uint64, status string) (*KitchenTicket, error) {
	var setClause string
	switch status {
	case "queued":
		setClause = `kt.status = 'queued', kt.started_at = NULL, kt.completed_at = NULL`
	case "in_progress":
		setClause = `kt.status = 'in_progress', kt.started_at = COALESCE(kt.started_at, CURRENT_TIMESTAMP), kt.completed_at = NULL`
	case "done":
		setClause = `kt.status = 'done', kt.started_at = COALESCE(kt.started_at, CURRENT_TIMESTAMP), kt.completed_at = CURRENT_TIMESTAMP`
	case "cancelled":
		setClause = `kt.status = 'cancelled', kt.completed_at = CURRENT_TIMESTAMP`
	default:
		return nil, fmt.Errorf("status ticket tidak valid")
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE kitchen_tickets kt
		JOIN orders o ON o.id = kt.order_id
		SET `+setClause+`
		WHERE kt.id = ? AND o.branch_id = ?`,
		id, branchID,
	)
	if err != nil {
		return nil, fmt.Errorf("update ticket status: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, sql.ErrNoRows
	}
	return r.FindTicketByID(ctx, id, branchID)
}

// EnsureTicketsForOrderTx creates missing kitchen tickets for all order items in an order.
func (r *Repository) EnsureTicketsForOrderTx(ctx context.Context, tx *sql.Tx, orderID uint64) ([]KitchenTicket, error) {
	var defaultStationID sql.NullInt64
	_ = tx.QueryRowContext(ctx, `
		SELECT ks.id
		FROM orders o
		JOIN kitchen_stations ks ON ks.branch_id = o.branch_id
		WHERE o.id = ? AND ks.is_active = 1
		ORDER BY ks.id ASC
		LIMIT 1`, orderID,
	).Scan(&defaultStationID)

	rows, err := tx.QueryContext(ctx, `
		SELECT oi.id
		FROM order_items oi
		LEFT JOIN kitchen_tickets kt ON kt.order_item_id = oi.id
		WHERE oi.order_id = ? AND kt.id IS NULL
		ORDER BY oi.id ASC`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("ensure tickets: list missing: %w", err)
	}
	defer rows.Close()

	itemIDs := make([]uint64, 0)
	for rows.Next() {
		var orderItemID uint64
		if err := rows.Scan(&orderItemID); err != nil {
			return nil, fmt.Errorf("ensure tickets: scan missing item: %w", err)
		}
		itemIDs = append(itemIDs, orderItemID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	created := make([]KitchenTicket, 0, len(itemIDs))
	for _, orderItemID := range itemIDs {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO kitchen_tickets (order_id, station_id, order_item_id)
			VALUES (?, ?, ?)`,
			orderID, nullableInt64(defaultStationID), orderItemID,
		)
		if err != nil {
			return nil, fmt.Errorf("ensure tickets: insert: %w", err)
		}
		insertedID, _ := res.LastInsertId()
		ticket, err := findTicketByIDWithQuerier(ctx, tx, uint64(insertedID))
		if err != nil {
			return nil, fmt.Errorf("ensure tickets: reload: %w", err)
		}
		created = append(created, *ticket)
	}
	return created, nil
}

// CancelTicketsForOrderTx marks all non-cancelled kitchen tickets in an order as cancelled.
func (r *Repository) CancelTicketsForOrderTx(ctx context.Context, tx *sql.Tx, orderID uint64) ([]KitchenTicket, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM kitchen_tickets
		WHERE order_id = ? AND status <> 'cancelled'
		ORDER BY id ASC`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel tickets: list: %w", err)
	}
	defer rows.Close()

	ticketIDs := make([]uint64, 0)
	for rows.Next() {
		var ticketID uint64
		if err := rows.Scan(&ticketID); err != nil {
			return nil, fmt.Errorf("cancel tickets: scan: %w", err)
		}
		ticketIDs = append(ticketIDs, ticketID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ticketIDs) == 0 {
		return []KitchenTicket{}, nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE kitchen_tickets
		SET status = 'cancelled', completed_at = CURRENT_TIMESTAMP
		WHERE order_id = ? AND status <> 'cancelled'`, orderID,
	); err != nil {
		return nil, fmt.Errorf("cancel tickets: update: %w", err)
	}

	updated := make([]KitchenTicket, 0, len(ticketIDs))
	for _, ticketID := range ticketIDs {
		ticket, err := findTicketByIDWithQuerier(ctx, tx, ticketID)
		if err != nil {
			return nil, fmt.Errorf("cancel tickets: reload: %w", err)
		}
		updated = append(updated, *ticket)
	}
	return updated, nil
}

func nullableInt64(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}
