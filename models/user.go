package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string `json:"username" gorm:"unique;not null"`
	Password       string `json:"password" gorm:"not null"`
	Email          string `json:"email" gorm:"unique"`
	Role           string `json:"role"`
	ProfilePicture string `json:"profile_picture" gorm:"default:'default.jpg'"`
}
