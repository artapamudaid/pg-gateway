package main

import (
	"log"
	"os"

	"payment-gateway/config"
	"payment-gateway/internal/middleware"
	"payment-gateway/internal/routes"
	"payment-gateway/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan variabel OS environment")
	}

	// 2. Inisialisasi Database
	config.ConnectDB()

	// 2.5 Inisialisasi Worker Pool (Maksimal 50 Worker secara bersamaan)
	services.StartWorkerPool(50)

	// 3. Inisialisasi Fiber App
	app := fiber.New(fiber.Config{
		// Mempercepat performa Fiber
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
	})

	// Middleware bawaan Fiber untuk log HTTP dan mencegah app mati saat panic
	app.Use(logger.New())
	app.Use(recover.New())
	
	// Middleware Rate Limiter untuk mencegah spam/DDoS
	app.Use(middleware.RateLimiter())

	// 4. Setup Routing
	app.Static("/", "./frontend")
	routes.SetupRoutes(app)

	// 5. Jalankan Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Printf("Payment Gateway berjalan di port %s", port)
	log.Fatal(app.Listen(port))
}
