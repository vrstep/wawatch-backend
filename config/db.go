package config

import (
	"log" // Import log package

	"github.com/vrstep/wawatch-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "postgres://postgres:postgres@localhost:5432/wawatchdb" // Use your actual DSN
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Running database auto migration...")
	// Add ALL your models here
	err = db.AutoMigrate(
		&models.User{},
		&models.AnimeCache{},
		&models.UserAnimeList{},
		&models.WatchProvider{},
		// Add any other models you create in the future here
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Database migration completed successfully.")

	DB = db
}
