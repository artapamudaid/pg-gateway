package handlers

import (
	"payment-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid"})
	}

	authService := services.NewAuthService()
	token, err := authService.Login(input.Username, input.Password)
	if err != nil {
		if err.Error() == "username atau password salah" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"token":   token,
		"message": "Login berhasil",
	})
}
