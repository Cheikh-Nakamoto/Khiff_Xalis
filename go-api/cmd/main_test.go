package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/brvm/go-api/internal/handler"
	"github.com/brvm/go-api/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

const testJWTSecret = "test-jwt-secret-for-qa"

// buildTestApp creates a minimal Fiber app with all routes wired to real
// handlers backed by nil dependencies. Every test exercises only the
// validation layer, which returns 400/401 before touching any store.
func buildTestApp() *fiber.App {
	app := fiber.New()
	app.Use(middleware.SecurityHeaders)

	app.Get("/health", handler.HealthCheck)

	api := app.Group("/api/v1")

	// Auth — nil service: validation returns 400 before AuthService is called.
	authH := handler.NewAuthHandler(nil)
	api.Post("/auth/register", authH.Register)
	api.Post("/auth/login", authH.Login)

	// Market — nil repo: GetTickers is static; GetMarketData validates
	// the ticker regex before querying the database.
	marketH := handler.NewMarketHandler(nil)
	api.Get("/market/tickers", marketH.GetTickers)
	api.Get("/market/data/:ticker", marketH.GetMarketData)

	// Orders — behind JWT; nil repo/redis: side validation returns 400
	// before any storage call.
	portfolioH := handler.NewPortfolioHandler(nil, nil, nil)
	protected := api.Group("", middleware.JWTMiddleware(testJWTSecret))
	protected.Post("/orders", portfolioH.CreateOrder)

	return app
}

// ---------------------------------------------------------------------------
// 1. Health endpoint
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	app := buildTestApp()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
}

// ---------------------------------------------------------------------------
// 2. Security headers
// ---------------------------------------------------------------------------

func TestSecurityHeaders(t *testing.T) {
	app := buildTestApp()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"X-Frame-Options present", "X-Frame-Options", "DENY"},
		{"X-Content-Type-Options present", "X-Content-Type-Options", "nosniff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resp.Header.Get(tt.header)
			if got != tt.expected {
				t.Errorf("header %s: expected %q, got %q", tt.header, tt.expected, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. GET /api/v1/market/tickers — static list, no DB required
// ---------------------------------------------------------------------------

func TestGetTickers(t *testing.T) {
	app := buildTestApp()

	req := httptest.NewRequest("GET", "/api/v1/market/tickers", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	count, ok := body["count"].(float64)
	if !ok || count == 0 {
		t.Errorf("expected non-zero count, got %v", body["count"])
	}

	data, ok := body["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Error("expected non-empty ticker list in data field")
	}
}

// ---------------------------------------------------------------------------
// 4. POST /api/v1/auth/register — missing-field validation
// ---------------------------------------------------------------------------

func TestRegisterValidation(t *testing.T) {
	app := buildTestApp()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"missing password", `{"email":"test@example.com","first_name":"John","last_name":"Doe"}`},
		{"missing email", `{"password":"password123","first_name":"John","last_name":"Doe"}`},
		{"missing first_name", `{"email":"test@example.com","password":"password123","last_name":"Doe"}`},
		{"missing last_name", `{"email":"test@example.com","password":"password123","first_name":"John"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 400 {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. POST /api/v1/auth/login — missing-field validation
// ---------------------------------------------------------------------------

func TestLoginValidation(t *testing.T) {
	app := buildTestApp()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"missing password", `{"email":"test@example.com"}`},
		{"missing email", `{"password":"password123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 400 {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 6. GET /api/v1/market/data/:ticker — invalid ticker format
//    TickerRe = ^[A-Z]{2,10}$
// ---------------------------------------------------------------------------

func TestInvalidTickerFormat(t *testing.T) {
	app := buildTestApp()

	tests := []struct {
		name   string
		ticker string
	}{
		{"exclamation mark", "INVALID!"},
		{"lowercase letters", "snts"},
		{"single char", "A"},
		{"contains digit", "AB1"},
		{"contains space", "AB CD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/market/data/"+tt.ticker, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 400 {
				t.Errorf("ticker %q: expected 400, got %d", tt.ticker, resp.StatusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. POST /api/v1/orders — invalid side validation (needs JWT)
// ---------------------------------------------------------------------------

func TestCreateOrderValidation(t *testing.T) {
	app := buildTestApp()

	// Generate a valid JWT so the request passes the auth middleware.
	token, err := middleware.GenerateAccessToken(testJWTSecret, "user-1", "test@example.com")
	if err != nil {
		t.Fatalf("generate test JWT: %v", err)
	}

	tests := []struct {
		name string
		body string
	}{
		{"invalid side", `{"ticker":"SNTS","side":"INVALID","quantity":100,"price":1000,"order_type":"LIMIT"}`},
		{"empty side", `{"ticker":"SNTS","side":"","quantity":100,"price":1000,"order_type":"LIMIT"}`},
		{"lowercase side", `{"ticker":"SNTS","side":"buy","quantity":100,"price":1000,"order_type":"LIMIT"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 400 {
				t.Errorf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}
