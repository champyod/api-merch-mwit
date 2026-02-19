package brandHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"

	"github.com/gofiber/fiber/v2"
)

func GetBrands(c *fiber.Ctx) error {
	db := database.DB
	var brands []model.Brand

	db.Find(&brands)

	return c.JSON(fiber.Map{"hasError": false, "metadata": nil, "errorMessage": "", "payload": brands})
}

func AddBrand(c *fiber.Ctx) error {
	db := database.DB
	type Input struct {
		Name string `json:"name"`
	}
	input := new(Input)
	if err := c.BodyParser(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	brand := model.Brand{Name: input.Name}
	if err := db.Create(&brand).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(fiber.Map{"success": true})
}
