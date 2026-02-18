package preorderHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"
	"api-merch-mwit/internal/utils"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
)

func GetPreorders(c *fiber.Ctx) error {
	db := database.DB
	var preorders []model.Preorder
	if err := db.Preload("Items.Item").Order("created_at DESC").Find(&preorders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch preorders"})
	}
	return c.JSON(fiber.Map{"hasError": false, "payload": preorders})
}

func AddPreorder(c *fiber.Ctx) error {
	db := database.DB

	// Parse Multipart Form
	name := c.FormValue("name")
	social := c.FormValue("social")
	contact := c.FormValue("contact")
	shippingMethod := c.FormValue("shipping_method")
	address := c.FormValue("address")
	itemsJson := c.FormValue("items")

	var inputItems []struct {
		ItemID   uint   `json:"item_id"`
		Size     string `json:"size"`
		Color    string `json:"color"`
		Quantity int    `json:"quantity"`
	}

	if err := json.Unmarshal([]byte(itemsJson), &inputItems); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid items format"})
	}

	if len(inputItems) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No items in order"})
	}

	// Calculate Total and Create Order Items
	var totalPrice float32
	var orderItems []model.OrderItem
	var firstItemPaymentAccount *model.PaymentAccount

	for _, ii := range inputItems {
		var item model.Item
		if err := db.Preload("PaymentAccount").First(&item, ii.ItemID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("Product %d not found", ii.ItemID)})
		}

		if firstItemPaymentAccount == nil && item.PaymentAccount != nil {
			firstItemPaymentAccount = item.PaymentAccount
		}

		itemPrice := item.Price
		if item.Discount > 0 {
			if item.Discount_type == "dollar" {
				itemPrice -= item.Discount
			} else {
				itemPrice -= (itemPrice * item.Discount / 100)
			}
		}

		orderItems = append(orderItems, model.OrderItem{
			ItemID:   item.ID,
			Size:     ii.Size,
			Color:    ii.Color,
			Quantity: ii.Quantity,
			Price:    itemPrice,
		})
		totalPrice += itemPrice * float32(ii.Quantity)
	}

	shippingCost := float32(0)
	if shippingMethod == "postal" {
		shippingCost = 50
	}
	totalPrice += shippingCost

	// Handle Slip Upload
	file, err := c.FormFile("slip")
	slipURL := ""
	if err == nil {
		// Ensure uploads directory exists
		uploadDir := "./public/uploads"
		os.MkdirAll(uploadDir, os.ModePerm)

		filename := uuid.New().String() + filepath.Ext(file.Filename)
		if err := c.SaveFile(file, filepath.Join(uploadDir, filename)); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save payment slip"})
		}
		slipURL = "/uploads/" + filename
	}

	customerUUID := c.Locals("customerUUID")
	var cuuid *string
	if customerUUID != nil {
		s := customerUUID.(string)
		cuuid = &s
	}

	preorder := model.Preorder{
		CustomerUUID:   cuuid,
		CustomerName:   name,
		Social:         social,
		ContactNumber:  contact,
		ShippingMethod: shippingMethod,
		Address:        address,
		Items:          orderItems,
		TotalPrice:     totalPrice,
		ShippingCost:   shippingCost,
		PaymentSlipURL: slipURL,
		Status:         "placed",
	}

	if err := db.Create(&preorder).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create order"})
	}

	// Generate QR (using the first item's payment account)
	qrPayload := ""
	if firstItemPaymentAccount != nil && firstItemPaymentAccount.PromptpayID != "" {
		qrPayload = utils.GeneratePromptPayPayload(firstItemPaymentAccount.PromptpayID, float64(totalPrice))
	}

	return c.JSON(fiber.Map{
		"preorder":        preorder,
		"payment_payload": qrPayload,
		"amount":          totalPrice,
	})
}

func UpdateStatus(c *fiber.Ctx) error {
	db := database.DB
	id := c.Params("id")

	var input struct {
		Status     string `json:"status"`
		TrackingNo string `json:"tracking_no"`
		Note       string `json:"note"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := db.Model(&model.Preorder{}).Where("id = ?", id).Updates(model.Preorder{
		Status:     input.Status,
		TrackingNo: input.TrackingNo,
		Note:       input.Note,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"success": true})
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

	if err := db.Preload("Items.Item").Order("created_at DESC").Find(&preorders).Error; err != nil {
		return c.Status(500).SendString("Export failed")
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=preorders_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(c)
	defer writer.Flush()

	writer.Write([]string{
		"Order ID", "Date", "Status", "Customer Name", "Social", "Contact", "Shipping", "Address", "Items", "Total Price", "Slip URL",
	})

	for _, p := range preorders {
		status := "Pending"
		if p.Completed == 1 {
			status = "Completed"
		}

		itemsStr := ""
		for _, item := range p.Items {
			itemsStr += fmt.Sprintf("%s (%s/%s) x%d; ", item.Item.Title, item.Size, item.Color, item.Quantity)
		}

		writer.Write([]string{
			strconv.FormatUint(uint64(p.ID), 10),
			p.CreatedAt.Format("2006-01-02 15:04"),
			status, p.CustomerName, p.Social, p.ContactNumber, p.ShippingMethod, p.Address, itemsStr, fmt.Sprintf("%.2f", p.TotalPrice), p.PaymentSlipURL,
		})
	}

	return nil
}
