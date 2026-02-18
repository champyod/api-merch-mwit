package preorderHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"
	"api-merch-mwit/internal/utils"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetPreorders(c *fiber.Ctx) error {
	db := database.DB
	var preorders []model.Preorder
	if err := db.Preload("Item").Order("created_at DESC").Find(&preorders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch preorders"})
	}
	return c.JSON(fiber.Map{"hasError": false, "payload": preorders})
}

func AddPreorder(c *fiber.Ctx) error {
	db := database.DB
	type Body struct {
		Name   string `json:"name"`
		Social string `json:"social"`
		Size   string `json:"size"`
		Color  string `json:"color"`
		ItemId string `json:"itemId"`
	}
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	itemId, _ := strconv.Atoi(body.ItemId)
	
	var item model.Item
	if err := db.Preload("PaymentAccount").First(&item, itemId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
	}

	preorder := model.Preorder{
		Customer_name: body.Name,
		Social:        body.Social,
		Size:          body.Size,
		Color:         body.Color,
		Item_id:       uint(itemId),
	}
	
	if err := db.Create(&preorder).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	promptpayID := "0812345678"
	if item.PaymentAccount != nil && item.PaymentAccount.PromptpayID != "" {
		promptpayID = item.PaymentAccount.PromptpayID
	}
	
	payload := utils.GeneratePromptPayPayload(promptpayID, float64(item.Price))

	return c.JSON(fiber.Map{
		"preorder":         preorder,
		"payment_payload":  payload,
		"amount":           item.Price,
	})
}

func CompletePreorder(c *fiber.Ctx) error {
	db := database.DB
	preorderId := c.Params("preorderId")

	if err := db.Model(&model.Preorder{}).Where("id = ?", preorderId).Update("completed", 1).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func ExportPreorders(c *fiber.Ctx) error {
	db := database.DB
	var preorders []model.Preorder
	
	// Preload everything for a complete report
	if err := db.Preload("Item.PaymentAccount").Order("created_at DESC").Find(&preorders).Error; err != nil {
		return c.Status(500).SendString("Export failed")
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=preorders_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(c)
	defer writer.Flush()

	// Header
	writer.Write([]string{
		"Order ID", "Date", "Status", "Customer Name", "Social", "Product", "Size", "Color", "Price", "Payment Account", "PromptPay ID",
	})

	for _, p := range preorders {
		status := "Pending"
		if p.Completed == 1 {
			status = "Completed"
		}
		
		price := "0"
		productTitle := "Unknown"
		accountName := "N/A"
		ppId := "N/A"
		
		if p.Item.ID != 0 {
			productTitle = p.Item.Title
			price = fmt.Sprintf("%.2f", p.Item.Price)
			if p.Item.PaymentAccount != nil {
				accountName = p.Item.PaymentAccount.Name
				ppId = p.Item.PaymentAccount.PromptpayID
			}
		}

		writer.Write([]string{
			strconv.FormatUint(uint64(p.ID), 10),
			p.CreatedAt.Format("2006-01-02 15:04"),
			status,
			p.Customer_name,
			p.Social,
			productTitle,
			p.Size,
			p.Color,
			price,
			accountName,
			ppId,
		})
	}

	return nil
}
