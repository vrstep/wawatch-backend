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
