package router

import (
	authHandler "api-merch-mwit/internal/handler/auth"
	brandHandler "api-merch-mwit/internal/handler/brand"
	orderHandler "api-merch-mwit/internal/handler/order"
	pageHandler "api-merch-mwit/internal/handler/page"
	paymentHandler "api-merch-mwit/internal/handler/payment"
	preorderHandler "api-merch-mwit/internal/handler/preorder"
	productHandler "api-merch-mwit/internal/handler/product"
	siteHandler "api-merch-mwit/internal/handler/site"
	cartHandler "api-merch-mwit/internal/handler/cart"
	"api-merch-mwit/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New())

	api := app.Group("/api")

	// Auth (public)
	auth := api.Group("/auth")
	auth.Get("/google", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
	auth.Get("/me", middleware.JWTAuth, authHandler.GetMe)
	auth.Post("/logout", middleware.JWTAuth, authHandler.Logout)

	// Public Products
	api.Get("/products", productHandler.GetItems)
	api.Get("/products/:itemId", productHandler.GetItem)

	// Preorders (with optional auth)
	api.Post("/preorders", middleware.OptionalJWTAuth, preorderHandler.AddPreorder)

	// Customer orders & cart (JWT protected)
	me := api.Group("/me", middleware.JWTAuth)
	me.Get("/orders", orderHandler.GetMyOrders)
	me.Get("/orders/:id", orderHandler.GetMyOrder)
	me.Get("/cart", cartHandler.GetCart)
	me.Post("/cart", cartHandler.UpdateCart)

	// Admin (JWT + role=admin)
	admin := api.Group("/admin", middleware.JWTAuth, middleware.AdminOnly)

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
	admin.Put("/preorders/:id/status", preorderHandler.UpdateStatus)
	admin.Put("/preorders/:preorderId/complete", preorderHandler.CompletePreorder)

	// Other template routes
	admin.Post("/brands", brandHandler.AddBrand)
	admin.Get("/brands", brandHandler.GetBrands)
	
	admin.Get("/pages", pageHandler.GetPages)
	admin.Post("/pages", pageHandler.AddPage)
	
	admin.Get("/site", siteHandler.GetSite)
	admin.Post("/site", siteHandler.EditSite)
}
