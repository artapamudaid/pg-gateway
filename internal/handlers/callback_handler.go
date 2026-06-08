package handlers

import (
	"encoding/json"
	"strings"


	"payment-gateway/internal/services"

	"github.com/gofiber/fiber/v2"
)

func HandleCallback(c *fiber.Ctx) error {
	provider := c.Params("provider")

	var refID string
	var payload []byte
	var tokenRaw string

	switch provider {
	case "flip":
		// Sesuai dokumentasi Flip: x-www-form-urlencoded
		tokenRaw = c.FormValue("token")
		dataRaw := c.FormValue("data")

		// Parsing JSON internal (data) untuk mengambil reference_id
		var dataObj map[string]interface{}
		if err := json.Unmarshal([]byte(dataRaw), &dataObj); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON Payload")
		}

		var ok bool
		refID, ok = dataObj["reference_id"].(string)
		if !ok || refID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("reference_id is required")
		}
		
		payload = c.Body()

	case "midtrans":
		// Contoh untuk midtrans, token biasanya di header (misal: Authorization)
		tokenRaw = c.Get("Authorization")
		
		var dataObj map[string]interface{}
		if err := json.Unmarshal(c.Body(), &dataObj); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON Payload")
		}
		
		var ok bool
		refID, ok = dataObj["order_id"].(string)
		if !ok || refID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("order_id is required")
		}
		payload = c.Body()

	default:
		return c.Status(fiber.StatusBadRequest).SendString("Provider not supported")
	}

	// Ekstrak Routing Code (Misal: "SHOP-INV-001" -> "SHOP")
	parts := strings.SplitN(refID, "-", 2)
	if len(parts) < 2 {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid reference_id format")
	}
	routingCode := parts[0]

	// Cari konfigurasi tujuan menggunakan Service
	destService := services.NewDestinationService()
	dest, err := destService.GetByRoutingCode(routingCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Destination not found")
	}

	// Validasi Token sesuai konfigurasi aplikasi di Database
	if tokenRaw != dest.ProviderToken {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized Token")
	}

	// Sesuai standar: Respon 200 secepatnya agar tidak timeout
	// Kita masukkan ke dalam Job Queue (Go Channel) untuk diproses oleh Worker Pool
	services.JobQueue <- services.CallbackJob{
		ReferenceID: refID,
		Payload:     payload,
	}

	return c.SendStatus(fiber.StatusOK)
}
