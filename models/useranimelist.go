// filepath: /home/vrstep/Narxoz/Go/wawatch/backend/models/useranimelist.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Define constants for status for better readability and maintenance
const (
	Watching   = "WATCHING"
	Completed  = "COMPLETED"
	Planned    = "PLANNED"
	Dropped    = "DROPPED"
	Paused     = "PAUSED"
	Rewatching = "REWATCHING" // Anilist uses this
)

type UserAnimeList struct {
	gorm.Model
	UserID          uint       `json:"user_id" gorm:"not null;index"`           // Foreign key to User model
	AnimeExternalID int        `json:"anime_external_id" gorm:"not null;index"` // ID from anilist.co
	Status          string     `json:"status" gorm:"type:varchar(20);index"`    // e.g., Watching, Completed, Planned
	Score           *int       `json:"score"`                                   // User's score (0-10 or 0-100 based on anilist?) - Use pointer for nullable
	Progress        int        `json:"progress"`                                // Episodes watched
	StartDate       *time.Time `json:"start_date"`                              // Pointer for nullable
	EndDate         *time.Time `json:"end_date"`                                // Pointer for nullable
	Notes           string     `json:"notes" gorm:"type:text"`
	RewatchCount    int        `json:"rewatch_count" gorm:"default:0"`

	// Optional: Add User navigation property if needed, GORM handles FK automatically
	// User User `gorm:"foreignKey:UserID"`
}
