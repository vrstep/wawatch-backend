package config

import (
	"github.com/vrstep/wawatch-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	db, err := gorm.Open(postgres.Open("postgres://postgres:postgres@localhost:5432/wawatchdb"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&models.User{})

	DB = db
}
