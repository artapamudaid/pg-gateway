package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"payment-gateway/config"
	"payment-gateway/internal/handlers"
	"payment-gateway/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB menginisialisasi SQLite in-memory untuk testing dan meng-injectnya ke config.DB
func setupTestDB() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}

	// Migrate tabel yang dibutuhkan
	db.AutoMigrate(&models.Destination{}, &models.CallbackLog{})
	
	// Timpa instance DB global
	config.DB = db
}

func TestGetDestinations(t *testing.T) {
	// 1. Setup DB
	setupTestDB()

	// 2. Insert dummy data ke dalam SQLite DB in-memory
	dummyData := models.Destination{
		AppName:     "Sistem Tagihan",
		Environment: "production",
		RoutingCode: "TAGIHAN",
		TargetURL:   "https://tagihan.example.com",
		SecretToken: "secret123",
		ProviderToken: "valid-token-123",
	}
	config.DB.Create(&dummyData)

	// 3. Setup Fiber app dan route
	app := fiber.New()
	app.Get("/api/destinations", handlers.GetDestinations)

	// 4. Lakukan request GET
	req := httptest.NewRequest("GET", "/api/destinations", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 5. Decode response json dan cek isinya
	var response map[string][]models.Destination
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	destinations := response["data"]
	assert.Len(t, destinations, 1, "Harusnya ada 1 data dummy")
	assert.Equal(t, "Sistem Tagihan", destinations[0].AppName)
	assert.Equal(t, "TAGIHAN", destinations[0].RoutingCode)
}

func TestCreateDestination(t *testing.T) {
	// 1. Setup DB (kosong)
	setupTestDB()

	// 2. Setup Fiber app dan route
	app := fiber.New()
	app.Post("/api/destinations", handlers.CreateDestination)

	// 3. Siapkan payload dummy
	newDest := models.Destination{
		AppName:     "Aplikasi Baru",
		Environment: "sandbox",
		RoutingCode: "APP_NEW",
		TargetURL:   "https://new.example.com",
		SecretToken: "supersecret",
		ProviderToken: "valid-token-123",
	}
	body, _ := json.Marshal(newDest)

	// 4. Lakukan request POST dengan payload json
	req := httptest.NewRequest("POST", "/api/destinations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	// 5. Verifikasi apakah data dummy benar-benar masuk ke database
	var count int64
	config.DB.Model(&models.Destination{}).Count(&count)
	assert.Equal(t, int64(1), count, "Data seharusnya bertambah 1 di database")

	var savedDest models.Destination
	config.DB.First(&savedDest)
	assert.Equal(t, "Aplikasi Baru", savedDest.AppName)
	assert.Equal(t, "APP_NEW", savedDest.RoutingCode)
	assert.Equal(t, "sandbox", savedDest.Environment)
}
