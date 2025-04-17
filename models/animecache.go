package models

import "gorm.io/gorm"

// Represents a minimal cache or reference to an anime from the external API
type AnimeCache struct {
	gorm.Model // Automatically includes ID, CreatedAt, UpdatedAt, DeletedAt
	// We use the external ID as our primary key for simplicity
	ID            int    `json:"id" gorm:"primaryKey;autoIncrement:false"` // Anilist ID
	Title         string `json:"title" gorm:"index"`                       // Store the primary title for searching/display
	CoverImage    string `json:"cover_image"`                              // URL to the cover image
	Format        string `json:"format"`                                   // e.g., TV, MOVIE, OVA
	TotalEpisodes *int   `json:"total_episodes"`                           // Pointer for nullable/unknown
	// Add other frequently accessed, relatively static fields if needed
	// LastFetched time.Time `json:"-"` // Track when details were last fetched from API (optional)
}
