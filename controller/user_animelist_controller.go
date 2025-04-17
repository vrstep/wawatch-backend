package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/models"
)

// GetUserAnimeList retrieves a user's anime list
func GetUserAnimeList(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userModel := user.(models.User)

	// Optional status filter
	status := c.Query("status")

	var query = config.DB.Where("user_id = ?", userModel.ID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var list []models.UserAnimeList
	query.Find(&list)

	// Fetch anime details for each list entry
	var result []gin.H
	for _, item := range list {
		var anime models.AnimeCache
		if err := config.DB.First(&anime, item.AnimeExternalID).Error; err != nil {
			// Skip entries where we can't find the anime cache
			continue
		}

		result = append(result, gin.H{
			"id":            item.ID,
			"status":        item.Status,
			"score":         item.Score,
			"progress":      item.Progress,
			"start_date":    item.StartDate,
			"end_date":      item.EndDate,
			"notes":         item.Notes,
			"rewatch_count": item.RewatchCount,
			"anime": gin.H{
				"id":             anime.ID,
				"title":          anime.Title,
				"cover_image":    anime.CoverImage,
				"format":         anime.Format,
				"total_episodes": anime.TotalEpisodes,
			},
		})
	}

	c.JSON(http.StatusOK, result)
}

// AddToAnimeList adds or updates an anime in the user's list
func AddToAnimeList(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userModel := user.(models.User)

	var input struct {
		AnimeID      int        `json:"anime_id" binding:"required"`
		Status       string     `json:"status" binding:"required"`
		Score        *int       `json:"score"`
		Progress     int        `json:"progress"`
		StartDate    *time.Time `json:"start_date"`
		EndDate      *time.Time `json:"end_date"`
		Notes        string     `json:"notes"`
		RewatchCount int        `json:"rewatch_count"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		models.Watching:   true,
		models.Completed:  true,
		models.Planned:    true,
		models.Dropped:    true,
		models.Paused:     true,
		models.Rewatching: true,
	}

	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	// Check if anime exists in cache, if not fetch it
	var animeCache models.AnimeCache
	if err := config.DB.First(&animeCache, input.AnimeID).Error; err != nil {
		// Fetch from AniList if not in cache
		anime, err := anilistClient.GetAnimeByID(input.AnimeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Anime not found"})
			return
		}

		// Create cache entry
		animeCache = anime.ToAnimeCache()
		config.DB.Create(&animeCache)
	}

	// Check if entry already exists
	var existingEntry models.UserAnimeList
	result := config.DB.Where("user_id = ? AND anime_external_id = ?", userModel.ID, input.AnimeID).First(&existingEntry)

	if result.RowsAffected > 0 {
		// Update existing entry
		existingEntry.Status = input.Status
		existingEntry.Score = input.Score
		existingEntry.Progress = input.Progress
		existingEntry.StartDate = input.StartDate
		existingEntry.EndDate = input.EndDate
		existingEntry.Notes = input.Notes
		existingEntry.RewatchCount = input.RewatchCount

		if err := config.DB.Save(&existingEntry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update list entry"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "List entry updated",
			"data":    existingEntry,
		})
		return
	}

	// Create new entry
	newEntry := models.UserAnimeList{
		UserID:          userModel.ID,
		AnimeExternalID: input.AnimeID,
		Status:          input.Status,
		Score:           input.Score,
		Progress:        input.Progress,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Notes:           input.Notes,
		RewatchCount:    input.RewatchCount,
	}

	if err := config.DB.Create(&newEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Anime added to list",
		"data":    newEntry,
	})
}

// UpdateListEntry updates a specific entry in the user's anime list
func UpdateListEntry(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userModel := user.(models.User)
	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID"})
		return
	}

	var entry models.UserAnimeList
	if err := config.DB.First(&entry, entryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	// Verify ownership
	if entry.UserID != userModel.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to edit this entry"})
		return
	}

	var input struct {
		Status       string     `json:"status"`
		Score        *int       `json:"score"`
		Progress     *int       `json:"progress"`
		StartDate    *time.Time `json:"start_date"`
		EndDate      *time.Time `json:"end_date"`
		Notes        *string    `json:"notes"`
		RewatchCount *int       `json:"rewatch_count"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	if input.Status != "" {
		// Validate status
		validStatuses := map[string]bool{
			models.Watching:   true,
			models.Completed:  true,
			models.Planned:    true,
			models.Dropped:    true,
			models.Paused:     true,
			models.Rewatching: true,
		}

		if !validStatuses[input.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}

		entry.Status = input.Status
	}

	if input.Score != nil {
		entry.Score = input.Score
	}

	if input.Progress != nil {
		entry.Progress = *input.Progress
	}

	if input.StartDate != nil {
		entry.StartDate = input.StartDate
	}

	if input.EndDate != nil {
		entry.EndDate = input.EndDate
	}

	if input.Notes != nil {
		entry.Notes = *input.Notes
	}

	if input.RewatchCount != nil {
		entry.RewatchCount = *input.RewatchCount
	}

	if err := config.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Entry updated successfully",
		"data":    entry,
	})
}

// Add this to anime_controller.go
// GetAnimeInUserList checks if an anime is in the user's list and returns its status
func GetAnimeInUserList(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userModel := user.(models.User)
	animeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anime ID"})
		return
	}

	var entry models.UserAnimeList
	result := config.DB.Where("user_id = ? AND anime_external_id = ?", userModel.ID, animeID).First(&entry)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"in_list": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"in_list":  true,
		"status":   entry.Status,
		"progress": entry.Progress,
		"score":    entry.Score,
	})
}

// DeleteListEntry removes an anime from the user's list
func DeleteListEntry(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userModel := user.(models.User)
	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID"})
		return
	}

	var entry models.UserAnimeList
	if err := config.DB.First(&entry, entryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	// Verify ownership
	if entry.UserID != userModel.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this entry"})
		return
	}

	if err := config.DB.Delete(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted successfully"})
}
