package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vrstep/wawatch-backend/config"
	"github.com/vrstep/wawatch-backend/models"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	users := []models.User{}
	config.DB.Find(&users)
	c.JSON(200, users)
}

func CreateUser(c *gin.Context) {
	user := models.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&user)
	c.JSON(200, user)
}

func DeleteUser(c *gin.Context) {
	user := models.User{}
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	config.DB.Delete(&user)
	c.JSON(200, gin.H{"message": "User deleted"})
}

func UpdateUser(c *gin.Context) {
	user := models.User{}
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	config.DB.Save(&user)
	c.JSON(200, user)
}

func GetUser(c *gin.Context) {
	user := models.User{}
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	c.JSON(200, user)
}

func Signup(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email,omitempty"` // Optional email field
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: body.Username,
		Password: string(hash),
	}

	// Only set email if provided
	if body.Email != "" {
		user.Email = body.Email
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User created successfully"})
}

func Login(c *gin.Context) {
	var body struct {
		Username string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	user := models.User{}
	if err := config.DB.Where("username = ?", body.Username).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": jwt.TimeFunc().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Auth", tokenString, 3600, "/", "localhost", false, true)

	c.JSON(200, gin.H{
		"token": tokenString,
		"user":  user,
	})

	c.JSON(200, gin.H{"message": "Login successful"})
}

func Validate(c *gin.Context) {
	user, _ := c.Get(("user"))

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}

// GetMyProfile retrieves the profile of the currently logged-in user
func GetMyProfile(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user := userInterface.(models.User)

	// Return user data, excluding the password hash
	c.JSON(http.StatusOK, gin.H{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email,
		"role":            user.Role,
		"profile_picture": user.ProfilePicture,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
	})
}

// UpdateMyProfile allows the logged-in user to update their profile
func UpdateMyProfile(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	currentUser := userInterface.(models.User)

	var input struct {
		Email          *string `json:"email"` // Use pointers to detect if field was provided
		ProfilePicture *string `json:"profile_picture"`
		// Add password update fields if needed (e.g., CurrentPassword, NewPassword)
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Find the user again to ensure we have the latest version
	var userToUpdate models.User
	if err := config.DB.First(&userToUpdate, currentUser.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields if they were provided in the request
	if input.Email != nil {
		// Add validation for email format if desired
		userToUpdate.Email = *input.Email
	}
	if input.ProfilePicture != nil {
		userToUpdate.ProfilePicture = *input.ProfilePicture
	}

	// Handle password update separately if implementing
	// e.g., check current password, hash new password

	if err := config.DB.Save(&userToUpdate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile", "details": err.Error()})
		return
	}

	// Return updated user data (excluding password)
	c.JSON(http.StatusOK, gin.H{
		"id":              userToUpdate.ID,
		"username":        userToUpdate.Username,
		"email":           userToUpdate.Email,
		"role":            userToUpdate.Role,
		"profile_picture": userToUpdate.ProfilePicture,
		"created_at":      userToUpdate.CreatedAt,
		"updated_at":      userToUpdate.UpdatedAt,
	})
}

// GetUserPublicAnimeList retrieves another user's anime list (public view)
func GetUserPublicAnimeList(c *gin.Context) {
	username := c.Param("username")

	var targetUser models.User
	if err := config.DB.Where("username = ?", username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Fetch the user's list (similar logic to GetUserAnimeList)
	var list []models.UserAnimeList
	config.DB.Where("user_id = ?", targetUser.ID).Find(&list)

	var result []gin.H
	for _, item := range list {
		var anime models.AnimeCache
		if err := config.DB.First(&anime, item.AnimeExternalID).Error; err == nil {
			result = append(result, gin.H{
				// Only include publicly relevant fields
				"status":   item.Status,
				"score":    item.Score,
				"progress": item.Progress,
				"anime": gin.H{
					"id":             anime.ID,
					"title":          anime.Title,
					"cover_image":    anime.CoverImage,
					"format":         anime.Format,
					"total_episodes": anime.TotalEpisodes,
				},
			})
		}
	}

	c.JSON(http.StatusOK, result)
}
