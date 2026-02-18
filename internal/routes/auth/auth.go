package authRoutes

import (
	authHandler "api-merch-mwit/internal/handler/auth"
	"api-merch-mwit/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router) {
	auth := router.Group("/auth")

	auth.Get("/", authHandler.GetUser)
	auth.Post("/login", authHandler.HandleLogin)
	auth.Post("/logout", middleware.AuthMiddleware, authHandler.HandleLogout)
}
