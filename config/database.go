package config

import (
	"fmt"
	"log"
	"os"

	"github.com/gokul-ms4/sop-manager/common/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		log.Fatal("Failed to create uuid extension:", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.SopHeading{},
		&models.SopItem{},
		&models.SopChunk{},
	)

	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	DB = db
	log.Println("Database connected and migrated successfully")
}
