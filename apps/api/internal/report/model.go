package report

// SalesSummary aggregates order totals for a given period.
type SalesSummary struct {
	TotalRevenue  string `json:"total_revenue"`
	OrderCount    int64  `json:"order_count"`
	AvgOrderValue string `json:"avg_order_value"`
}

// DailySales holds the revenue and order count for a single calendar day.
type DailySales struct {
	Date       string `json:"date"`
	Revenue    string `json:"revenue"`
	OrderCount int64  `json:"order_count"`
}

// TopItem represents a best-selling menu item by quantity and revenue.
type TopItem struct {
	MenuItemID   uint64 `json:"menu_item_id"`
	ItemName     string `json:"item_name"`
	TotalQty     int64  `json:"total_qty"`
	TotalRevenue string `json:"total_revenue"`
}

// SalesReportResponse is the response for GET /api/v1/reports/sales.
type SalesReportResponse struct {
	Summary SalesSummary `json:"summary"`
	Daily   []DailySales `json:"daily"`
}
