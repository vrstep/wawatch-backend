package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/api"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/models"
)

// var anilistClient = api.NewAniListClient()

// Use the interface type instead of the concrete type
var anilistClient api.AniListAPI

// SetAniListClient allows injecting a client (real or mock)
// This function will be used by tests to inject a mock client.
func SetAniListClient(client api.AniListAPI) {
	anilistClient = client
}

// Initialize with the real client by default when the package loads.
func init() {
	SetAniListClient(api.NewAniListClient())
}

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

// GetPopularAnime fetches popular anime from AniList
func GetPopularAnime(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	results, total, err := anilistClient.GetPopularAnime(page, perPage) // Needs implementation in api/anilist.go
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch popular anime: " + err.Error()})
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

// GetTrendingAnime fetches trending anime from AniList
func GetTrendingAnime(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	results, total, err := anilistClient.GetTrendingAnime(page, perPage) // Needs implementation in api/anilist.go
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trending anime: " + err.Error()})
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

// GetAnimeBySeason fetches anime for a specific year and season from AniList
func GetAnimeBySeason(c *gin.Context) {
	yearParam := c.Param("year")
	seasonParam := strings.ToUpper(c.Param("season")) // WINTER, SPRING, SUMMER, FALL

	year, err := strconv.Atoi(yearParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year format"})
		return
	}

	validSeasons := map[string]bool{"WINTER": true, "SPRING": true, "SUMMER": true, "FALL": true}
	if !validSeasons[seasonParam] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid season. Use WINTER, SPRING, SUMMER, or FALL"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	results, total, err := anilistClient.GetAnimeBySeason(year, seasonParam, page, perPage) // Needs implementation in api/anilist.go
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch anime by season: " + err.Error()})
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

// GetAnimeRecommendations fetches recommendations (placeholder, uses popular for now)
func GetAnimeRecommendations(c *gin.Context) {
	// _, exists := c.Get("user") // Get user if needed for personalized recommendations
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
	// 	return
	// }
	// userModel := userInterface.(models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "10")) // Fewer recommendations usually

	// Ideally, call a specific recommendation function in anilistClient
	// For now, let's reuse popular as a placeholder
	// results, total, err := anilistClient.GetAnimeRecommendations(userModel.ID, page, perPage)
	results, total, err := anilistClient.GetPopularAnime(page, perPage) // Placeholder
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recommendations: " + err.Error()})
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
