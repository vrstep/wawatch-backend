package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/api"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/models"
)

var anilistClient = api.NewAniListClient()

// SearchAnime handles anime search requests
func SearchAnime(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	results, total, err := anilistClient.SearchAnime(query, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search anime: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"perPage":     perPage,
			"totalPages":  (total + perPage - 1) / perPage,
			"hasNextPage": page*perPage < total,
		},
	})
}

// GetAnimeDetails fetches detailed information about an anime
func GetAnimeDetails(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anime ID"})
		return
	}

	// First, check if we have this in cache
	var cache models.AnimeCache
	if err := config.DB.First(&cache, id).Error; err == nil {
		// We have it in cache, but still need detailed info
	}

	// Get detailed info from AniList
	anime, err := anilistClient.GetAnimeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch anime details: " + err.Error()})
		return
	}

	// Update or create cache entry
	cacheEntry := anime.ToAnimeCache()
	config.DB.Save(&cacheEntry)

	// Get watch providers
	var providers []models.WatchProvider
	config.DB.Where("anime_id = ?", id).Find(&providers)

	c.JSON(http.StatusOK, gin.H{
		"anime":     anime,
		"providers": providers,
	})
}

// AddWatchProvider adds a new watch provider for an anime
func AddWatchProvider(c *gin.Context) {
	var provider models.WatchProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the anime exists in our cache
	var animeCache models.AnimeCache
	if err := config.DB.First(&animeCache, provider.AnimeID).Error; err != nil {
		// Fetch from AniList if not in cache
		anime, err := anilistClient.GetAnimeByID(int(provider.AnimeID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Anime not found"})
			return
		}

		// Create cache entry
		cacheEntry := anime.ToAnimeCache()
		config.DB.Create(&cacheEntry)
	}

	// Save the provider
	if err := config.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save provider"})
		return
	}

	c.JSON(http.StatusOK, provider)
}
