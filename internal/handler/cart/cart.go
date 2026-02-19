package cartHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetCart(c *fiber.Ctx) error {
	db := database.DB
	customerUUID := c.Locals("customerUUID").(string)

	var cart model.Cart
	err := db.Preload("Items.Item.Images").Where("customer_uuid = ?", customerUUID).First(&cart).Error
	if err == gorm.ErrRecordNotFound {
		// Create an empty cart for the customer if one doesn't exist
		cart = model.Cart{
			UUID:         uuid.New().String(),
			CustomerUUID: customerUUID,
		}
		db.Create(&cart)
		return c.JSON(fiber.Map{"hasError": false, "payload": []model.CartItem{}})
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"hasError": true, "errorMessage": "Failed to fetch cart"})
	}

	return c.JSON(fiber.Map{"hasError": false, "payload": cart.Items})
}

func UpdateCart(c *fiber.Ctx) error {
	db := database.DB
	customerUUID := c.Locals("customerUUID").(string)

	var inputItems []struct {
		ItemID   uint   `json:"item_id"`
		Size     string `json:"size"`
		Color    string `json:"color"`
		Quantity int    `json:"quantity"`
	}

	if err := c.BodyParser(&inputItems); err != nil {
		return c.Status(400).JSON(fiber.Map{"hasError": true, "errorMessage": "Invalid input"})
	}

	// Find or create cart
	var cart model.Cart
	err := db.Where("customer_uuid = ?", customerUUID).First(&cart).Error
	if err == gorm.ErrRecordNotFound {
		cart = model.Cart{
			UUID:         uuid.New().String(),
			CustomerUUID: customerUUID,
		}
		if err := db.Create(&cart).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"hasError": true, "errorMessage": "Failed to create cart"})
		}
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"hasError": true, "errorMessage": "Database error"})
	}

	// Transactional update: clear and re-insert items for simplicity
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cart_uuid = ?", cart.UUID).Unscoped().Delete(&model.CartItem{}).Error; err != nil {
			return err
		}

		if len(inputItems) > 0 {
			cartItems := make([]model.CartItem, len(inputItems))
			for i, ii := range inputItems {
				cartItems[i] = model.CartItem{
					CartUUID: cart.UUID,
					ItemID:   ii.ItemID,
					Size:     ii.Size,
					Color:    ii.Color,
					Quantity: ii.Quantity,
				}
			}
			if err := tx.Create(&cartItems).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"hasError": true, "errorMessage": "Failed to update cart items"})
	}

	return c.JSON(fiber.Map{"hasError": false, "payload": "Cart updated"})
}
