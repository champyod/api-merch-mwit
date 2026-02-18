package database

import (
	"api-merch-mwit/internal/model"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=password dbname=merch_db port=5432 sslmode=disable"
	}

	// Retry connection for docker startup
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying in 5s... (%d/5)", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to PostgreSQL database")

	// Auto Migrate
	err = DB.AutoMigrate(
		&model.User{},
		&model.Brand{},
		&model.Item{},
		&model.Image{},
		&model.Size{},
		&model.Color{},
		&model.Preorder{},
		&model.Page{},
		&model.PaymentAccount{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
}