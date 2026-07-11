package report

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository handles database queries for sales reports.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new report repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetSalesSummary returns aggregated totals for completed orders in the
// given date range, scoped to the organisation.  branchID is optional.
func (r *Repository) GetSalesSummary(
	ctx context.Context,
	orgID uint64,
	branchID *uint64,
	dateFrom, dateTo string,
) (*SalesSummary, error) {
	q := `
		SELECT
		    COALESCE(CAST(SUM(o.total_amount) AS CHAR), '0.00') AS total_revenue,
		    COUNT(*)                                             AS order_count,
		    COALESCE(CAST(AVG(o.total_amount) AS CHAR), '0.00') AS avg_order_value
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE b.organization_id = ?
		  AND o.status = 'completed'
		  AND DATE(o.created_at) BETWEEN ? AND ?`
	args := []any{orgID, dateFrom, dateTo}
	if branchID != nil {
		q += " AND o.branch_id = ?"
		args = append(args, *branchID)
	}

	s := &SalesSummary{}
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(
		&s.TotalRevenue, &s.OrderCount, &s.AvgOrderValue,
	); err != nil {
		return nil, fmt.Errorf("get sales summary: %w", err)
	}
	return s, nil
}

// GetDailySales returns a day-by-day revenue and order count breakdown for
// completed orders in the given date range.
func (r *Repository) GetDailySales(
	ctx context.Context,
	orgID uint64,
	branchID *uint64,
	dateFrom, dateTo string,
) ([]DailySales, error) {
	q := `
		SELECT
		    DATE_FORMAT(o.created_at, '%Y-%m-%d')      AS sale_date,
		    CAST(SUM(o.total_amount) AS CHAR)           AS revenue,
		    COUNT(*)                                    AS order_count
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE b.organization_id = ?
		  AND o.status = 'completed'
		  AND DATE(o.created_at) BETWEEN ? AND ?`
	args := []any{orgID, dateFrom, dateTo}
	if branchID != nil {
		q += " AND o.branch_id = ?"
		args = append(args, *branchID)
	}
	q += " GROUP BY sale_date ORDER BY sale_date ASC"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("get daily sales: %w", err)
	}
	defer rows.Close()

	var result []DailySales
	for rows.Next() {
		var d DailySales
		if err := rows.Scan(&d.Date, &d.Revenue, &d.OrderCount); err != nil {
			return nil, fmt.Errorf("scan daily sales: %w", err)
		}
		result = append(result, d)
	}
	if result == nil {
		result = []DailySales{}
	}
	return result, rows.Err()
}

// GetTopItems returns the best-selling menu items by total quantity sold,
// limited to `limit` rows.
func (r *Repository) GetTopItems(
	ctx context.Context,
	orgID uint64,
	branchID *uint64,
	dateFrom, dateTo string,
	limit int,
) ([]TopItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	q := `
		SELECT
		    oi.menu_item_id,
		    oi.item_name,
		    SUM(oi.quantity)                          AS total_qty,
		    CAST(SUM(oi.subtotal) AS CHAR)            AS total_revenue
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN branches b ON b.id = o.branch_id
		WHERE b.organization_id = ?
		  AND o.status = 'completed'
		  AND DATE(o.created_at) BETWEEN ? AND ?`
	args := []any{orgID, dateFrom, dateTo}
	if branchID != nil {
		q += " AND o.branch_id = ?"
		args = append(args, *branchID)
	}
	q += fmt.Sprintf(" GROUP BY oi.menu_item_id, oi.item_name ORDER BY total_qty DESC LIMIT %d", limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("get top items: %w", err)
	}
	defer rows.Close()

	var result []TopItem
	for rows.Next() {
		var it TopItem
		if err := rows.Scan(&it.MenuItemID, &it.ItemName, &it.TotalQty, &it.TotalRevenue); err != nil {
			return nil, fmt.Errorf("scan top item: %w", err)
		}
		result = append(result, it)
	}
	if result == nil {
		result = []TopItem{}
	}
	return result, rows.Err()
}
