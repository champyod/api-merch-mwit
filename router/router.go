package router

import (
	"api-merch-mwit/internal/handler/auth"
	"api-merch-mwit/internal/handler/brand"
	"api-merch-mwit/internal/handler/page"
	"api-merch-mwit/internal/handler/payment"
	"api-merch-mwit/internal/handler/preorder"
	"api-merch-mwit/internal/handler/product"
	"api-merch-mwit/internal/handler/site"
	"api-merch-mwit/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New())

	api := app.Group("/api")

	// Auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/register", authHandler.Register)

	// Public Products
	api.Get("/products", productHandler.GetItems)
	api.Get("/products/:itemId", productHandler.GetItem)

	// Preorders
	api.Post("/preorders", preorderHandler.AddPreorder)

	// Admin (Protected)
	admin := api.Group("/admin", middleware.Auth)

	// Product management
	admin.Post("/products", productHandler.AddItem)
	admin.Put("/products/:itemId", productHandler.EditItem)
	admin.Delete("/products/:itemId", productHandler.DeleteItem)

	// Payment Accounts
	admin.Get("/payment-accounts", paymentHandler.GetAccounts)
	admin.Post("/payment-accounts", paymentHandler.CreateAccount)
	admin.Delete("/payment-accounts/:id", paymentHandler.DeleteAccount)

	// Preorder management
	admin.Get("/preorders", preorderHandler.GetPreorders)
	admin.Get("/preorders/export", preorderHandler.ExportPreorders)
	admin.Put("/preorders/:preorderId/complete", preorderHandler.CompletePreorder)

	// Other template routes
	admin.Post("/brands", brandHandler.AddBrand)
	admin.Get("/brands", brandHandler.GetBrands)
	
	admin.Get("/pages", pageHandler.GetPages)
	admin.Post("/pages", pageHandler.AddPage)
	
	admin.Get("/site", siteHandler.GetSite)
	admin.Post("/site", siteHandler.EditSite)
}