package handler

import (
	"log"
	"time"

	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// MarketHandler handles market-data endpoints.
type MarketHandler struct {
	MarketRepo *repository.MarketRepo
}

// NewMarketHandler creates a new MarketHandler.
func NewMarketHandler(marketRepo *repository.MarketRepo) *MarketHandler {
	return &MarketHandler{MarketRepo: marketRepo}
}

// GetTickers handles GET /api/v1/market/tickers (static list).
func (h *MarketHandler) GetTickers(c *fiber.Ctx) error {
	tickers := model.BRVMTickers()
	return c.JSON(fiber.Map{"count": len(tickers), "data": tickers})
}

// GetMarketData handles GET /api/v1/market/data/:ticker.
func (h *MarketHandler) GetMarketData(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if !model.TickerRe.MatchString(ticker) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ticker format"})
	}

	limit := c.QueryInt("limit", 100)
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	data, total, err := h.MarketRepo.GetMarketDataPaginated(c.Context(), ticker, limit, offset)
	if err != nil {
		log.Printf("DB query error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
	}

	totalPages := (int(total) + limit - 1) / limit

	return c.JSON(fiber.Map{
		"ticker":      ticker,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"data":        data,
	})
}

// GetLatestData handles GET /api/v1/market/latest/:ticker.
func (h *MarketHandler) GetLatestData(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if !model.TickerRe.MatchString(ticker) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ticker format"})
	}

	d, err := h.MarketRepo.GetLatestData(c.Context(), ticker)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No data found for ticker"})
	}

	return c.JSON(d)
}

// GetFundamental handles GET /api/v1/fundamental/:ticker.
func (h *MarketHandler) GetFundamental(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if !model.TickerRe.MatchString(ticker) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ticker format"})
	}

	f, err := h.MarketRepo.GetFundamentals(c.Context(), ticker)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No fundamental data found"})
	}

	return c.JSON(f)
}

// HealthCheck handles GET /health.
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
}
