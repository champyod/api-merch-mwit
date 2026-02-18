package paymentHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"
	"github.com/gofiber/fiber/v2"
)

func GetAccounts(c *fiber.Ctx) error {
	db := database.DB
	var accounts []model.PaymentAccount
	db.Find(&accounts)
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
