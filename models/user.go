package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string `json:"username" gorm:"unique;not null"`
	Password       string `json:"password" gorm:"not null"`
	Email          string `json:"email" gorm:"unique;not null"`
	Role           string `json:"role" gorm:"not null"`
	ProfilePicture string `json:"profile_picture" gorm:"default:'default.jpg'"`
}
