package handlers

import (
	"encoding/json"
	"fmt"
	"strings"


	"payment-gateway/internal/models"
	"payment-gateway/internal/services"

	"github.com/gofiber/fiber/v2"
)

func HandleCallback(c *fiber.Ctx) error {
	provider := c.Params("provider")
	env := c.Params("env")
	if env == "" {
		env = "production"
	}

	var refID string
	var payload []byte
	var tokenRaw string
	var fallbackTitle string

	switch provider {
	case "flip":
		// Sesuai dokumentasi Flip: x-www-form-urlencoded
		tokenRaw = c.FormValue("token")
		dataRaw := c.FormValue("data")

		// Parsing JSON internal (data)
		var dataObj map[string]interface{}
		if err := json.Unmarshal([]byte(dataRaw), &dataObj); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON Payload")
		}

		// Cari routing prefix dari berbagai field yang mungkin dikirim oleh Flip
		// Flip biasanya mengirim: bill_link_id, title, reference, atau id
		var ok bool
		if val, exists := dataObj["reference_id"]; exists && val != nil && val != "" {
			refID = fmt.Sprintf("%v", val)
			ok = true
		} else if val, exists := dataObj["bill_link_id"]; exists && val != nil && val != "" {
			refID = fmt.Sprintf("%v", val)
			ok = true
		} else if val, exists := dataObj["id"]; exists && val != nil && val != "" {
			refID = fmt.Sprintf("%v", val)
			ok = true
		}

		if val, exists := dataObj["bill_title"]; exists && val != nil && val != "" {
			fallbackTitle = fmt.Sprintf("%v", val)
		} else if val, exists := dataObj["title"]; exists && val != nil && val != "" {
			fallbackTitle = fmt.Sprintf("%v", val)
		}

		if !ok || refID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Routing identifier (reference_id/bill_link_id/id) is required")
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
	
	destService := services.NewDestinationService()
	var dest models.Destination
	var err error

	if len(parts) >= 2 {
		routingCode := parts[0]
		dest, err = destService.GetByRoutingCodeAndEnv(routingCode, env)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Destination not found for routing code")
		}
	} else {
		// Fallback untuk testing dari Sandbox Flip atau jika Flip tidak mengirim reference_id
		// Maka kita mencoba mencocokkan bill_title dengan app_name, jika tidak ketemu baru gunakan default
		dests, errList := destService.GetAll()
		if errList != nil || len(dests) == 0 {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid reference_id format and no default destination found in database")
		}

		matched := false
		if fallbackTitle != "" {
			for _, d := range dests {
				if d.Environment == env && strings.Contains(strings.ToLower(fallbackTitle), strings.ToLower(d.AppName)) {
					dest = d
					matched = true
					break
				}
			}
		}

		if !matched {
			// Find the first default for the env
			for _, d := range dests {
				if d.Environment == env {
					dest = d
					matched = true
					break
				}
			}
			if !matched {
				dest = dests[0]
			}
		}
	}

	// Validasi Token sesuai konfigurasi aplikasi di Database
	if tokenRaw != dest.ProviderToken {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized Token")
	}

	// Meneruskan request secara sinkron ke aplikasi tujuan
	statusCode, bodyBytes, err := services.ForwardSync(refID, dest, payload)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString("Gateway Error: Failed to reach destination")
	}

	// Meneruskan respons (status code dan JSON) dari aplikasi tujuan kembali ke Flip
	c.Set("Content-Type", "application/json")
	return c.Status(statusCode).Send(bodyBytes)
}
