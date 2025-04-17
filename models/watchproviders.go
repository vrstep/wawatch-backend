package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WatchProvider represents where an anime can be watched
type WatchProvider struct {
	gorm.Model             // Automatically includes ID, CreatedAt, UpdatedAt, DeletedAt
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnimeID      int       `gorm:"not null" json:"anime_id"` // External ID from AniList
	ProviderName string    `gorm:"not null" json:"provider_name"`
	ProviderURL  string    `json:"provider_url"`
	Region       string    `gorm:"size:2" json:"region"` // Country code
	IsSub        bool      `gorm:"default:false" json:"is_sub"`
	IsDub        bool      `gorm:"default:false" json:"is_dub"`
	LastUpdated  time.Time `json:"last_updated"`

	// Relationships
	AnimeCache AnimeCache `gorm:"foreignKey:AnimeID" json:"-"`
}
