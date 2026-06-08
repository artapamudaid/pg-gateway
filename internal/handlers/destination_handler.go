package handlers

import (
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"

	"github.com/gofiber/fiber/v2"
)

// GET /api/destinations
func GetDestinations(c *fiber.Ctx) error {
	destService := services.NewDestinationService()
	destinations, _ := destService.GetAll()
	return c.JSON(fiber.Map{"data": destinations})
}

// POST /api/destinations
func CreateDestination(c *fiber.Ctx) error {
	dest := new(models.Destination)
	if err := c.BodyParser(dest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "error": err.Error()})
	}

	destService := services.NewDestinationService()
	if err := destService.Create(dest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan data", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Destinasi berhasil ditambahkan", "data": dest})
}

// PUT /api/destinations/:id
func UpdateDestination(c *fiber.Ctx) error {
	id := c.Params("id")
	dest := new(models.Destination)

	if err := c.BodyParser(dest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid"})
	}

	destService := services.NewDestinationService()
	updatedDest, err := destService.Update(id, dest)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Data tidak ditemukan"})
	}

	return c.JSON(fiber.Map{"message": "Destinasi berhasil diperbarui", "data": updatedDest})
}

// DELETE /api/destinations/:id
func DeleteDestination(c *fiber.Ctx) error {
	id := c.Params("id")
	destService := services.NewDestinationService()
	
	if err := destService.Delete(id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Data tidak ditemukan"})
	}

	return c.JSON(fiber.Map{"message": "Destinasi berhasil dihapus"})
}
