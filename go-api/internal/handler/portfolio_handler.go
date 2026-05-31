package handler

import (
	"fmt"
	"log"

	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// PortfolioHandler handles portfolio and order endpoints.
type PortfolioHandler struct {
	PortfolioRepo *repository.PortfolioRepo
	MarketRepo    *repository.MarketRepo
	Redis         *redis.Client
}

// NewPortfolioHandler creates a new PortfolioHandler.
func NewPortfolioHandler(portfolioRepo *repository.PortfolioRepo, marketRepo *repository.MarketRepo, rdb *redis.Client) *PortfolioHandler {
	return &PortfolioHandler{
		PortfolioRepo: portfolioRepo,
		MarketRepo:    marketRepo,
		Redis:         rdb,
	}
}

// GetPortfolio handles GET /api/v1/portfolio.
func (h *PortfolioHandler) GetPortfolio(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	rows, err := h.PortfolioRepo.GetPortfolio(c.Context(), userID)
	if err != nil {
		log.Printf("DB query error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
	}

	var positions []model.PortfolioPosition
	var totalValue, totalPnL float64

	for _, row := range rows {
		p := model.PortfolioPosition{
			Ticker:      row.Ticker,
			Quantity:    row.Quantity,
			AvgBuyPrice: row.AvgBuyPrice,
		}
		// Get current price
		price, priceErr := h.MarketRepo.CurrentPrice(c.Context(), p.Ticker)
		if priceErr == nil {
			p.CurrentPrice = price
		}

		p.MarketValue = float64(p.Quantity) * p.CurrentPrice
		cost := float64(p.Quantity) * p.AvgBuyPrice
		p.UnrealizedPnL = p.MarketValue - cost
		if cost > 0 {
			p.UnrealizedPnLPct = p.UnrealizedPnL / cost * 100
		}
		totalValue += p.MarketValue
		totalPnL += p.UnrealizedPnL
		positions = append(positions, p)
	}

	if positions == nil {
		positions = []model.PortfolioPosition{}
	}

	return c.JSON(fiber.Map{
		"total_value": totalValue,
		"total_pnl":   totalPnL,
		"positions":   positions,
	})
}

// GetOrders handles GET /api/v1/orders.
func (h *PortfolioHandler) GetOrders(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	status := c.Query("status", "ALL")
	limit := c.QueryInt("limit", 50)
	if limit < 1 || limit > 200 {
		limit = 50
	}

	rows, err := h.PortfolioRepo.GetOrders(c.Context(), userID, status, limit)
	if err != nil {
		log.Printf("DB query error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
	}

	var orders []map[string]interface{}
	for _, o := range rows {
		orders = append(orders, map[string]interface{}{
			"ticker": o.Ticker, "side": o.Side, "quantity": o.Quantity,
			"price": o.Price, "order_type": o.OrderType, "status": o.Status,
			"created_at": o.CreatedAt,
		})
	}

	if orders == nil {
		orders = []map[string]interface{}{}
	}

	return c.JSON(fiber.Map{"count": len(orders), "orders": orders})
}

// CreateOrder handles POST /api/v1/orders.
func (h *PortfolioHandler) CreateOrder(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req model.OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	// Validation
	if req.Ticker == "" || !model.TickerRe.MatchString(req.Ticker) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ticker"})
	}
	if req.Side != "BUY" && req.Side != "SELL" {
		return c.Status(400).JSON(fiber.Map{"error": "Side must be BUY or SELL"})
	}
	if req.Quantity <= 0 || req.Quantity > 1000000 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid quantity"})
	}
	if req.Price <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid price"})
	}
	validOrderTypes := map[string]bool{"LIMIT": true, "MARKET": true, "STOP_LOSS": true, "TAKE_PROFIT": true}
	if !validOrderTypes[req.OrderType] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid order_type"})
	}
	if req.OrderType == "STOP_LOSS" || req.OrderType == "TAKE_PROFIT" {
		if req.StopPrice <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Stop price required for stop/take orders"})
		}
	}

	// Insert order
	if err := h.PortfolioRepo.CreateOrder(c.Context(), userID, req.Ticker, req.Side, req.Quantity, req.Price, req.OrderType); err != nil {
		log.Printf("Order insert error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create order"})
	}

	// Publish to Redis for async processing
	h.Redis.Publish(c.Context(), "orders:new",
		fmt.Sprintf("%s|%s|%s|%d|%f", userID, req.Ticker, req.Side, req.Quantity, req.Price))

	return c.Status(201).JSON(fiber.Map{
		"message": "Order created",
		"ticker":  req.Ticker,
		"side":    req.Side,
		"status":  "PENDING",
	})
}
