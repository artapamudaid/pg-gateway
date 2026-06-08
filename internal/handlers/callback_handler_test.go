package handlers_test

import (
	"net/http/httptest"
	"net/url"

	"strings"
	"testing"

	"payment-gateway/config"
	"payment-gateway/internal/handlers"
	"payment-gateway/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandleFlipCallback(t *testing.T) {
	// Setup DB in-memory dari fungsi helper di destination_handler_test.go
	// agar goroutine yang memanggil DB di service tidak panic
	setupTestDB()

	// Setup dummy destination in DB for token validation
	config.DB.Create(&models.Destination{
		AppName:       "Test App",
		RoutingCode:   "SHOP",
		TargetURL:     "http://localhost/test",
		SecretToken:   "secret",
		ProviderToken: "valid-token-123",
	})

	// Setup Fiber App
	app := fiber.New()
	app.Post("/:provider/callback", handlers.HandleCallback)

	// List of test cases
	tests := []struct {
		name         string
		token        string
		dataJSON     string
		expectedCode int
	}{
		{
			name:         "Success valid payload",
			token:        "valid-token-123",
			dataJSON:     `{"reference_id": "SHOP-123", "amount": 10000}`,
			expectedCode: fiber.StatusOK,
		},
		{
			name:         "Unauthorized invalid token",
			token:        "wrong-token",
			dataJSON:     `{"reference_id": "SHOP-123", "amount": 10000}`,
			expectedCode: fiber.StatusUnauthorized,
		},
		{
			name:         "Invalid JSON payload",
			token:        "valid-token-123",
			dataJSON:     `{invalid-json}`,
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name:         "Missing reference_id",
			token:        "valid-token-123",
			dataJSON:     `{"amount": 10000}`,
			expectedCode: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sesuai dokumentasi Flip, request body menggunakan form-urlencoded
			form := url.Values{}
			form.Add("token", tt.token)
			form.Add("data", tt.dataJSON)

			// Buat request POST
			req := httptest.NewRequest("POST", "/flip/callback", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Eksekusi request menggunakan app Fiber
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)

			// Verifikasi response status code
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}
