package paymentHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"
	"github.com/gofiber/fiber/v2"
)

func GetAccounts(c *fiber.Ctx) error {
	db := database.DB
	var accounts []model.PaymentAccount
	if err := db.Find(&accounts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch accounts"})
	}

	// Calculate analytics for each account
	for i := range accounts {
		var totalOrders int64
		var totalRevenue float64

		// Join preorders through items linked to this account
		db.Table("preorders").
			Joins("JOIN items ON items.id = preorders.item_id").
			Where("items.payment_account_id = ?", accounts[i].ID).
			Count(&totalOrders)

		db.Table("preorders").
			Joins("JOIN items ON items.id = preorders.item_id").
			Where("items.payment_account_id = ?", accounts[i].ID).
			Select("SUM(items.price)").
			Scan(&totalRevenue)

		accounts[i].TotalOrders = totalOrders
		accounts[i].TotalRevenue = totalRevenue
	}

	return c.JSON(accounts)
}

func CreateAccount(c *fiber.Ctx) error {
	db := database.DB
	var account model.PaymentAccount
	if err := c.BodyParser(&account); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.Create(&account)
	return c.Status(201).JSON(account)
}

func DeleteAccount(c *fiber.Ctx) error {
	db := database.DB
	id := c.Params("id")
	db.Delete(&model.PaymentAccount{}, id)
	return c.SendStatus(204)
}