package table

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Repository handles DB operations for dining areas and tables.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new table repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Dining Areas ──────────────────────────────────────────────────────────────

func scanDiningArea(row interface{ Scan(...any) error }) (*DiningArea, error) {
	da := &DiningArea{}
	var desc sql.NullString
	err := row.Scan(&da.ID, &da.BranchID, &da.Name, &desc, &da.IsActive, &da.CreatedAt, &da.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		da.Description = &desc.String
	}
	return da, nil
}

// ListDiningAreas returns all dining areas for a branch.
func (r *Repository) ListDiningAreas(ctx context.Context, branchID uint64) ([]DiningArea, error) {
	const q = `
		SELECT id, branch_id, name, description, is_active, created_at, updated_at
		FROM dining_areas
		WHERE branch_id = ?
		ORDER BY name`
	rows, err := r.db.QueryContext(ctx, q, branchID)
	if err != nil {
		return nil, fmt.Errorf("list dining areas: %w", err)
	}
	defer rows.Close()

	var areas []DiningArea
	for rows.Next() {
		da, err := scanDiningArea(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dining area: %w", err)
		}
		areas = append(areas, *da)
	}
	if areas == nil {
		areas = []DiningArea{}
	}
	return areas, rows.Err()
}

// FindDiningAreaByID returns a dining area scoped to a branch.
func (r *Repository) FindDiningAreaByID(ctx context.Context, id, branchID uint64) (*DiningArea, error) {
	const q = `
		SELECT id, branch_id, name, description, is_active, created_at, updated_at
		FROM dining_areas WHERE id = ? AND branch_id = ? LIMIT 1`
	return scanDiningArea(r.db.QueryRowContext(ctx, q, id, branchID))
}

// CreateDiningArea inserts a new dining area.
func (r *Repository) CreateDiningArea(ctx context.Context, branchID uint64, req *CreateDiningAreaRequest) (*DiningArea, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO dining_areas (branch_id, name, description) VALUES (?, ?, ?)`,
		branchID, req.Name, req.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("create dining area: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindDiningAreaByID(ctx, uint64(id), branchID)
}

// ── Tables ────────────────────────────────────────────────────────────────────

func scanTable(row interface{ Scan(...any) error }) (*Table, error) {
	t := &Table{}
	err := row.Scan(&t.ID, &t.DiningAreaID, &t.BranchID, &t.TableNumber, &t.Capacity, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// ListTables returns all tables for a branch.
func (r *Repository) ListTables(ctx context.Context, branchID uint64) ([]Table, error) {
	const q = `
		SELECT id, dining_area_id, branch_id, table_number, capacity, is_active, created_at, updated_at
		FROM restaurant_tables
		WHERE branch_id = ?
		ORDER BY dining_area_id, table_number`
	rows, err := r.db.QueryContext(ctx, q, branchID)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		t, err := scanTable(rows)
		if err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, *t)
	}
	if tables == nil {
		tables = []Table{}
	}
	return tables, rows.Err()
}

// FindTableByID returns a table scoped to a branch.
func (r *Repository) FindTableByID(ctx context.Context, id, branchID uint64) (*Table, error) {
	const q = `
		SELECT id, dining_area_id, branch_id, table_number, capacity, is_active, created_at, updated_at
		FROM restaurant_tables WHERE id = ? AND branch_id = ? LIMIT 1`
	return scanTable(r.db.QueryRowContext(ctx, q, id, branchID))
}

// CreateTable inserts a new table.
func (r *Repository) CreateTable(ctx context.Context, branchID uint64, req *CreateTableRequest) (*Table, error) {
	// Verify dining area belongs to branch
	var daID uint64
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM dining_areas WHERE id = ? AND branch_id = ? LIMIT 1`,
		req.DiningAreaID, branchID,
	).Scan(&daID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("area makan tidak ditemukan pada cabang ini")
	}
	if err != nil {
		return nil, fmt.Errorf("create table: verify dining area: %w", err)
	}

	capacity := req.Capacity
	if capacity == 0 {
		capacity = 4
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO restaurant_tables (dining_area_id, branch_id, table_number, capacity) VALUES (?, ?, ?, ?)`,
		req.DiningAreaID, branchID, req.TableNumber, capacity,
	)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindTableByID(ctx, uint64(id), branchID)
}

// UpdateTable applies partial updates to a table.
func (r *Repository) UpdateTable(ctx context.Context, id, branchID uint64, req *UpdateTableRequest) (*Table, error) {
	clauses := []string{}
	args := []any{}

	if req.TableNumber != nil {
		clauses = append(clauses, "table_number = ?")
		args = append(args, *req.TableNumber)
	}
	if req.Capacity != nil {
		clauses = append(clauses, "capacity = ?")
		args = append(args, *req.Capacity)
	}
	if req.IsActive != nil {
		clauses = append(clauses, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	if len(clauses) == 0 {
		return r.FindTableByID(ctx, id, branchID)
	}

	query := "UPDATE restaurant_tables SET " + strings.Join(clauses, ", ") + " WHERE id = ? AND branch_id = ?"
	args = append(args, id, branchID)
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("update table: %w", err)
	}
	return r.FindTableByID(ctx, id, branchID)
}
