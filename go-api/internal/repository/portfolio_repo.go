package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PortfolioRepo handles portfolio and order queries.
type PortfolioRepo struct {
	DB *pgxpool.Pool
}

// PortfolioRow is a raw portfolio row before price enrichment.
type PortfolioRow struct {
	Ticker      string
	Quantity    int
	AvgBuyPrice float64
}

// GetPortfolio returns all positions for a user.
func (r *PortfolioRepo) GetPortfolio(ctx context.Context, userID string) ([]PortfolioRow, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT ticker, quantity, avg_buy_price
		 FROM portfolios WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []PortfolioRow
	for rows.Next() {
		var p PortfolioRow
		if err := rows.Scan(&p.Ticker, &p.Quantity, &p.AvgBuyPrice); err != nil {
			continue
		}
		positions = append(positions, p)
	}
	return positions, nil
}

// OrderRow is a raw order row.
type OrderRow struct {
	Ticker      string
	Side        string
	Quantity    int
	Price       float64
	OrderType   string
	Status      string
	CreatedAt   time.Time
}

// GetOrders returns orders for a user, optionally filtered by status.
func (r *PortfolioRepo) GetOrders(ctx context.Context, userID, status string, limit int) ([]OrderRow, error) {
	query := `SELECT ticker, side, quantity, price, order_type, status, created_at
	          FROM orders WHERE user_id = $1`
	args := []interface{}{userID}
	argIdx := 2

	if status != "ALL" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderRow
	for rows.Next() {
		var o OrderRow
		if err := rows.Scan(&o.Ticker, &o.Side, &o.Quantity, &o.Price, &o.OrderType, &o.Status, &o.CreatedAt); err != nil {
			continue
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// CreateOrder inserts a new PENDING order.
func (r *PortfolioRepo) CreateOrder(ctx context.Context, userID, ticker, side string, quantity int, price float64, orderType string) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO orders (user_id, ticker, side, quantity, price, order_type, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'PENDING', NOW())`,
		userID, ticker, side, quantity, price, orderType)
	return err
}
