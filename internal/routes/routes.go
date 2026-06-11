package routes

import (
	"payment-gateway/internal/handlers"
	"payment-gateway/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {

	//Health Check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// =========================================
	// PUBLIC ROUTE (Hanya diakses oleh server Payment Gateway)
	// =========================================
	app.Post("/:provider/callback", handlers.HandleCallback) // Defaults to production
	app.Post("/:env/:provider/callback", handlers.HandleCallback) // e.g. /sandbox/flip/callback

	// =========================================
	// API ROUTES (Diakses oleh Frontend Dashboard)
	// =========================================
	api := app.Group("/api")

	// Endpoint Publik untuk Frontend
	api.Post("/login", handlers.Login)

	// API Group yang dilindungi Middleware JWT
	admin := api.Group("/", middleware.Protected())

	// CRUD Routes untuk Destination Settings
	admin.Get("/destinations", handlers.GetDestinations)
	admin.Post("/destinations", handlers.CreateDestination)
	admin.Put("/destinations/:id", handlers.UpdateDestination)
	admin.Delete("/destinations/:id", handlers.DeleteDestination)
}
