package handler

import (
	"log"

	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	AuthService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	// Validation
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "All fields are required"})
	}
	if len(req.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	resp, err := h.AuthService.Register(c.Context(), req)
	if err != nil {
		switch err {
		case service.ErrEmailExists:
			return c.Status(409).JSON(fiber.Map{"error": "Email already registered"})
		default:
			log.Printf("Register error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
		}
	}

	return c.Status(201).JSON(resp)
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}

	resp, err := h.AuthService.Login(c.Context(), req)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		default:
			log.Printf("Login error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}
	}

	return c.JSON(resp)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	accessToken, err := h.AuthService.Refresh(c.Context(), body.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidRefreshToken:
			return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
		default:
			log.Printf("Refresh error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"expires_in":   900,
	})
}
