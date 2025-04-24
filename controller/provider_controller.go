package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/models"
)

// UpdateWatchProvider updates an existing watch provider entry
// TODO: Add admin authorization check if needed
func UpdateWatchProvider(c *gin.Context) {
	providerIDParam := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID format"})
		return
	}

	var provider models.WatchProvider
	if err := config.DB.First(&provider, "id = ?", providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Watch provider not found"})
		return
	}

	// Bind JSON data to update the provider
	// Use pointers to only update fields that are actually sent
	var input struct {
		ProviderName *string `json:"provider_name"`
		ProviderURL  *string `json:"provider_url"`
		Region       *string `json:"region"`
		IsSub        *bool   `json:"is_sub"`
		IsDub        *bool   `json:"is_dub"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Apply updates
	updated := false
	if input.ProviderName != nil {
		provider.ProviderName = *input.ProviderName
		updated = true
	}
	if input.ProviderURL != nil {
		provider.ProviderURL = *input.ProviderURL
		updated = true
	}
	if input.Region != nil {
		// Add validation for region code if needed
		provider.Region = *input.Region
		updated = true
	}
	if input.IsSub != nil {
		provider.IsSub = *input.IsSub
		updated = true
	}
	if input.IsDub != nil {
		provider.IsDub = *input.IsDub
		updated = true
	}

	if updated {
		provider.LastUpdated = time.Now() // Update the timestamp
		if err := config.DB.Save(&provider).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
			return
		}
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteWatchProvider deletes a watch provider entry
// TODO: Add admin authorization check if needed
func DeleteWatchProvider(c *gin.Context) {
	providerIDParam := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID format"})
		return
	}

	var provider models.WatchProvider
	// Ensure the provider exists before attempting deletion
	if err := config.DB.First(&provider, "id = ?", providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Watch provider not found"})
		return
	}

	// Perform the delete operation
	if err := config.DB.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watch provider deleted successfully"})
}
