package routes_test

import (
	"testing"

	"payment-gateway/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	app := fiber.New()
	routes.SetupRoutes(app)

	// Dapatkan semua routes yang telah diregistrasikan
	registeredRoutes := app.GetRoutes()

	// Buat map dari (method + path) untuk kemudahan pengecekan
	routeMap := make(map[string]bool)
	for _, route := range registeredRoutes {
		routeMap[route.Method+" "+route.Path] = true
	}

	// Daftar expected routes yang seharusnya ada
	expectedRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/:provider/callback"},
		{"POST", "/api/login"},
		{"GET", "/api/destinations"},
		{"POST", "/api/destinations"},
		{"PUT", "/api/destinations/:id"},
		{"DELETE", "/api/destinations/:id"},
	}

	// Cek apakah semua expected route telah terdaftar
	for _, expected := range expectedRoutes {
		t.Run(expected.method+" "+expected.path, func(t *testing.T) {
			key := expected.method + " " + expected.path
			_, exists := routeMap[key]
			assert.True(t, exists, "Route %s %s belum terdaftar", expected.method, expected.path)
		})
	}
}
