package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders adds standard security headers to every response.
func SecurityHeaders(c *fiber.Ctx) error {
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-XSS-Protection", "1; mode=block")
	c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Set("Content-Security-Policy", "default-src 'self'")
	return c.Next()
}
