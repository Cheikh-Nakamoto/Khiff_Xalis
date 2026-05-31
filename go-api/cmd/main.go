// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	brvmgrpc "github.com/brvm/go-api/internal/grpc"
	"github.com/brvm/go-api/internal/handler"
	"github.com/brvm/go-api/internal/middleware"
	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/repository"
	"github.com/brvm/go-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	jwtSecret := model.JWTSecret
	engineAddr := model.GetEnv("ENGINE_ADDR", "localhost:50051")

	// PostgreSQL
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_URL")})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer rdb.Close()

	// Rust Engine gRPC client (optional — API degrades gracefully if engine is down)
	engineClient, err := brvmgrpc.NewEngineClient(engineAddr)
	if err != nil {
		log.Printf("WARNING: Rust Engine not reachable at %s — signal endpoints will use DB fallback: %v", engineAddr, err)
		engineClient = nil
	} else {
		log.Printf("Connected to Rust Engine at %s", engineAddr)
	}
	defer func() {
		if engineClient != nil {
			engineClient.Close()
		}
	}()

	// Wire repositories
	userRepo := &repository.UserRepo{DB: db}
	marketRepo := &repository.MarketRepo{DB: db}
	portfolioRepo := &repository.PortfolioRepo{DB: db}

	// Wire services
	authService := &service.AuthService{Users: userRepo, JWTSecret: jwtSecret}
	signalService := &service.SignalService{MarketRepo: marketRepo, Redis: rdb, Engine: engineClient}

	// Wire handlers
	authH := handler.NewAuthHandler(authService)
	marketH := handler.NewMarketHandler(marketRepo)
	signalH := handler.NewSignalHandler(signalService)
	portfolioH := handler.NewPortfolioHandler(portfolioRepo, marketRepo, rdb)

	// Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "BRVM Trading Engine API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("ERROR: %s %s - %v", c.Method(), c.Path(), err)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		},
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{Format: "[${time}] ${status} - ${method} ${path}\n"}))
	app.Use(middleware.SecurityHeaders)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     model.GetEnv("CORS_ORIGINS", "http://localhost:3000"),
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        model.RateLimit,
		Expiration: model.RateWindow,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
	}))

	// Health
	app.Get("/health", handler.HealthCheck)

	api := app.Group("/api/v1")

	// Public auth routes
	api.Post("/auth/register", authH.Register)
	api.Post("/auth/login", authH.Login)
	api.Post("/auth/refresh", authH.Refresh)

	// Public market data (read-only)
	api.Get("/market/tickers", marketH.GetTickers)
	api.Get("/market/data/:ticker", marketH.GetMarketData)
	api.Get("/market/latest/:ticker", marketH.GetLatestData)
	api.Get("/fundamental/:ticker", marketH.GetFundamental)
	api.Get("/signals/:ticker", signalH.GetSignal)
	api.Get("/signals/scan", signalH.ScanSignals)

	// Protected routes (JWT required)
	protected := api.Group("", middleware.JWTMiddleware(jwtSecret))
	protected.Get("/portfolio", portfolioH.GetPortfolio)
	protected.Get("/orders", portfolioH.GetOrders)
	protected.Post("/orders", portfolioH.CreateOrder)

	// WebSocket with token validation
	api.Get("/signals/ws", websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		if token == "" {
			c.Close()
			return
		}
		t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !t.Valid {
			c.Close()
			return
		}
		signalH.WSSignalHandler(c)
	}))

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + model.Port); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatal(err)
	}
}
