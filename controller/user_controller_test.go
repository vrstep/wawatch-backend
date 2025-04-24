package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vrstep/wawatch-backend/models"
	"gorm.io/gorm"
)

// Test GetMyProfile Endpoint
func TestGetMyProfile(t *testing.T) {
	// Setup
	_, cleanup := SetupTestDB(t) // Setup mock DB (even if not used directly, good practice)
	defer cleanup()
	router := SetupGin()

	// Mock User Data
	mockUser := models.User{
		Model:          gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Username:       "testuser",
		Email:          "test@example.com",
		Role:           "user",
		ProfilePicture: "pic.jpg",
	}

	// Setup Route and Handler
	router.GET("/profile", func(c *gin.Context) {
		// Simulate middleware setting the user
		c.Set("user", mockUser)
		GetMyProfile(c)
	})

	// Perform Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
	// Simulate authentication (e.g., by setting cookie if middleware checks it,
	// or directly setting context as done above)

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, float64(mockUser.ID), responseBody["id"]) // JSON numbers are float64
	assert.Equal(t, mockUser.Username, responseBody["username"])
	assert.Equal(t, mockUser.Email, responseBody["email"])
	assert.NotContains(t, responseBody, "password") // Ensure password is not returned
}

// Test UpdateMyProfile Endpoint
func TestUpdateMyProfile(t *testing.T) {
	// Setup
	mock, cleanup := SetupTestDB(t)
	defer cleanup()
	router := SetupGin()

	// Mock User Data
	mockUserID := uint(1)
	mockUser := models.User{
		Model:    gorm.Model{ID: mockUserID},
		Username: "testuser",
		Email:    "old@test.com",
	}

	// Input Data
	updateInput := gin.H{
		"email":           "new@test.com",
		"profile_picture": "new_pic.jpg",
	}
	requestBody, _ := json.Marshal(updateInput)

	// Mock DB Expectations
	// 1. Expect GORM to fetch the user first
	mock.ExpectQuery(EscapeQuery(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(int64(mockUserID)).
		// Ensure all columns GORM expects for models.User (including gorm.Model) are provided
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "password", "email", "role", "profile_picture"}).
			AddRow(mockUserID, time.Now(), time.Now(), nil, mockUser.Username, "hashedpassword", mockUser.Email, mockUser.Role, "old_pic.jpg")) // Add dummy values for all fields

	// 2. Expect GORM to begin a transaction, update, and commit
	mock.ExpectBegin()
	// Use int64 for the ID in the WHERE clause
	mock.ExpectExec(EscapeQuery(`UPDATE "users" SET "email"=$1,"profile_picture"=$2,"updated_at"=$3 WHERE "users"."deleted_at" IS NULL AND "id" = $4`)).
		WithArgs(updateInput["email"], updateInput["profile_picture"], MockAnyTime{}, int64(mockUserID)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Route and Handler
	router.PUT("/profile", func(c *gin.Context) {
		c.Set("user", mockUser) // Simulate authenticated user
		UpdateMyProfile(c)
	})

	// Perform Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, updateInput["email"], responseBody["email"])
	assert.Equal(t, updateInput["profile_picture"], responseBody["profile_picture"])

	// Verify all DB expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test GetUserPublicAnimeList Endpoint
func TestGetUserPublicAnimeList(t *testing.T) {
	mock, cleanup := SetupTestDB(t)
	defer cleanup()
	router := SetupGin()

	targetUsername := "publicuser"
	targetUserID := uint(2)
	animeID1 := 101
	animeID2 := 102

	// Mock DB Expectations
	// 1. Find the target user by username
	// Revert to specific username string
	mock.ExpectQuery(EscapeQuery(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(targetUsername). // Reverted to specific string
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(targetUserID, targetUsername))

	// 2. Find the user's anime list entries
	listRows := sqlmock.NewRows([]string{"id", "user_id", "anime_external_id", "status", "score", "progress"}).
		AddRow(10, targetUserID, animeID1, models.Watching, 8, 5).
		AddRow(11, targetUserID, animeID2, models.Completed, 9, 12)
	// Use int64 for user_id if targetUserID is uint
	mock.ExpectQuery(EscapeQuery(`SELECT * FROM "user_anime_lists" WHERE user_id = $1 AND "user_anime_lists"."deleted_at" IS NULL`)).
		WithArgs(int64(targetUserID)). // Use int64
		WillReturnRows(listRows)

	// 3. Find anime cache details for each list item
	animeRows1 := sqlmock.NewRows([]string{"id", "title", "cover_image", "format", "total_episodes"}).
		AddRow(animeID1, "Anime Title 1", "cover1.jpg", "TV", 12)
	// Use int64 for anime ID if animeID1 is int
	mock.ExpectQuery(EscapeQuery(`SELECT * FROM "anime_caches" WHERE "anime_caches"."id" = $1 AND "anime_caches"."deleted_at" IS NULL ORDER BY "anime_caches"."id" LIMIT 1`)).
		WithArgs(int64(animeID1)). // Use int64
		WillReturnRows(animeRows1)

	animeRows2 := sqlmock.NewRows([]string{"id", "title", "cover_image", "format", "total_episodes"}).
		AddRow(animeID2, "Anime Title 2", "cover2.jpg", "MOVIE", 1)
	// Use int64 for anime ID if animeID2 is int
	mock.ExpectQuery(EscapeQuery(`SELECT * FROM "anime_caches" WHERE "anime_caches"."id" = $1 AND "anime_caches"."deleted_at" IS NULL ORDER BY "anime_caches"."id" LIMIT 1`)).
		WithArgs(int64(animeID2)). // Use int64
		WillReturnRows(animeRows2)

	// Setup Route
	router.GET("/users/:username/animelist", GetUserPublicAnimeList)

	// Perform Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/"+targetUsername+"/animelist", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Len(t, responseBody, 2)

	// Check first item
	assert.Equal(t, models.Watching, responseBody[0]["status"])
	assert.Equal(t, float64(8), responseBody[0]["score"]) // JSON numbers
	animeDetails1 := responseBody[0]["anime"].(map[string]interface{})
	assert.Equal(t, float64(animeID1), animeDetails1["id"])
	assert.Equal(t, "Anime Title 1", animeDetails1["title"])

	// Check second item
	assert.Equal(t, models.Completed, responseBody[1]["status"])
	assert.Equal(t, float64(9), responseBody[1]["score"]) // JSON numbers
	animeDetails2 := responseBody[1]["anime"].(map[string]interface{})
	assert.Equal(t, float64(animeID2), animeDetails2["id"])
	assert.Equal(t, "Anime Title 2", animeDetails2["title"])

	assert.NoError(t, mock.ExpectationsWereMet())
}
