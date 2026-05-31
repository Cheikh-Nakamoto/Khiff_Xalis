package handler

import (
	"context"
	"log"

	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// SignalHandler handles signal endpoints.
type SignalHandler struct {
	SignalService *service.SignalService
}

// NewSignalHandler creates a new SignalHandler.
func NewSignalHandler(signalService *service.SignalService) *SignalHandler {
	return &SignalHandler{SignalService: signalService}
}

// GetSignal handles GET /api/v1/signals/:ticker.
func (h *SignalHandler) GetSignal(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if !model.TickerRe.MatchString(ticker) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ticker format"})
	}

	// Check Redis cache
	cached := h.SignalService.GetCachedSignalString(c.Context(), ticker)
	if cached != "" {
		return c.JSON(fiber.Map{"cached": true, "ticker": ticker, "signal": cached})
	}

	_, sr, err := h.SignalService.GetSignal(c.Context(), ticker)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No signal data available. Collector may not have run yet."})
	}

	return c.JSON(sr)
}

// ScanSignals handles GET /api/v1/signals/scan.
func (h *SignalHandler) ScanSignals(c *fiber.Ctx) error {
	minScore := c.QueryFloat("min_score", 60)
	signalType := c.Query("signal_type")

	validSignals := map[string]bool{"BUY": true, "STRONG_BUY": true, "SELL": true, "STRONG_SELL": true, "HOLD": true, "": true}
	if !validSignals[signalType] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid signal_type"})
	}

	results, err := h.SignalService.ScanSignals(c.Context(), minScore, signalType)
	if err != nil {
		log.Printf("DB query error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database query failed"})
	}

	return c.JSON(fiber.Map{"count": len(results), "data": results})
}

// WSSignalHandler handles WebSocket connections for real-time signals.
func (h *SignalHandler) WSSignalHandler(c *websocket.Conn) {
	pubsub := h.SignalService.SubscribeRealtime(context.Background())
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			break
		}
	}
}
