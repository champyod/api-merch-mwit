package authHandler

import (
	"api-merch-mwit/database"
	"api-merch-mwit/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

func InitGoogleOauth() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:        []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:      google.Endpoint,
	}
}

func GoogleLogin(c *fiber.Ctx) error {
	InitGoogleOauth()
	url := googleOauthConfig.AuthCodeURL("state")
	return c.Redirect(url)
}

func GoogleCallback(c *fiber.Ctx) error {
	InitGoogleOauth()
	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to exchange token"})
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info"})
	}
	defer resp.Body.Close()

	var profile struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode user info"})
	}

	db := database.DB
	var customer model.Customer
	result := db.Where("google_id = ?", profile.ID).First(&customer)

	if result.Error != nil {
		// Create new customer
		customer = model.Customer{
			GoogleID:  profile.ID,
			Email:     profile.Email,
			Name:      profile.Name,
			AvatarURL: profile.Picture,
			Role:      "customer",
		}
		if profile.Email == os.Getenv("GOOGLE_ROOT_EMAIL") {
			customer.Role = "admin"
		}
		db.Create(&customer)
	} else {
		// Update existing customer profile
		customer.Email = profile.Email
		customer.Name = profile.Name
		customer.AvatarURL = profile.Picture
		if profile.Email == os.Getenv("GOOGLE_ROOT_EMAIL") {
			customer.Role = "admin"
		}
		db.Save(&customer)
	}

	// Generate JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uuid": customer.UUID,
		"role": customer.Role,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HTTPOnly: true,
		Secure:   os.Getenv("NODE_ENV") == "production",
		SameSite: "Lax",
		Path:     "/",
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return c.Redirect(frontendURL + "/auth/callback")
}

func GetMe(c *fiber.Ctx) error {
	// The uuid is injected into c.Locals by JWT middleware
	uuid := c.Locals("customerUUID")
	if uuid == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var customer model.Customer
	if err := database.DB.Where("uuid = ?", uuid).First(&customer).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}

	return c.JSON(fiber.Map{
		"uuid":       customer.UUID,
		"name":       customer.Name,
		"email":      customer.Email,
		"avatar_url": customer.AvatarURL,
		"role":       customer.Role,
	})
}

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   os.Getenv("NODE_ENV") == "production",
		SameSite: "Lax",
		Path:     "/",
	})
	return c.JSON(fiber.Map{"success": true})
}