package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter mengembalikan middleware untuk membatasi jumlah request per IP
func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,             // Maksimal 100 request
		Expiration: 1 * time.Minute, // per 1 menit
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "Terlalu banyak request, silakan coba lagi nanti.",
			})
		},
	})
}
