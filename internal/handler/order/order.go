package orderHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"

	"github.com/gofiber/fiber/v2"
)

func GetMyOrders(c *fiber.Ctx) error {
	db := database.DB
	customerUUID := c.Locals("customerUUID")

	if customerUUID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var preorders []model.Preorder
	if err := db.Preload("Items.Item").Where("customer_uuid = ?", customerUUID).Order("created_at DESC").Find(&preorders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}

	return c.JSON(fiber.Map{"hasError": false, "payload": preorders})
}

func GetMyOrder(c *fiber.Ctx) error {
	db := database.DB
	customerUUID := c.Locals("customerUUID")
	id := c.Params("id")

	if customerUUID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var preorder model.Preorder
	if err := db.Preload("Items.Item").Where("id = ? AND customer_uuid = ?", id, customerUUID).First(&preorder).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.JSON(fiber.Map{"hasError": false, "payload": preorder})
}
