package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vrstep/wawatch-backend/models"
	"gorm.io/gorm"
)

// Test GetUserAnimeListStats Endpoint
func TestGetUserAnimeListStats(t *testing.T) {
	mock, cleanup := SetupTestDB(t)
	defer cleanup()
	router := SetupGin()

	mockUserID := uint(1)
	mockUser := models.User{Model: gorm.Model{ID: mockUserID}}

	score8 := 8
	score9 := 9

	// Mock DB Expectations
	rows := sqlmock.NewRows([]string{"id", "user_id", "anime_external_id", "status", "score", "progress"}).
		AddRow(1, mockUserID, 101, models.Watching, &score8, 5).
		AddRow(2, mockUserID, 102, models.Completed, &score9, 12).
		AddRow(3, mockUserID, 103, models.Completed, &score9, 24).
		AddRow(4, mockUserID, 104, models.Planned, nil, 0).
		AddRow(5, mockUserID, 105, models.Watching, nil, 1) // No score

	mock.ExpectQuery(EscapeQuery(`SELECT * FROM "user_anime_lists" WHERE user_id = $1 AND "user_anime_lists"."deleted_at" IS NULL`)).
		WithArgs(mockUserID).
		WillReturnRows(rows)

	// Setup Route
	router.GET("/animelist/stats", func(c *gin.Context) {
		c.Set("user", mockUser)
		GetUserAnimeListStats(c)
	})

	// Perform Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/animelist/stats", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	assert.Equal(t, float64(5), responseBody["total_anime"])
	assert.Equal(t, float64(5+12+24+0+1), responseBody["episodes_watched"]) // 42

	// Calculate expected mean score: (8 + 9 + 9) / 3 = 26 / 3 = 8.666...
	assert.InDelta(t, 8.666, responseBody["mean_score"], 0.001)

	statusCounts := responseBody["status_counts"].(map[string]interface{})
	assert.Equal(t, float64(2), statusCounts[models.Watching])
	assert.Equal(t, float64(2), statusCounts[models.Completed])
	assert.Equal(t, float64(1), statusCounts[models.Planned])
	assert.Equal(t, float64(0), statusCounts[models.Dropped])
	assert.Equal(t, float64(0), statusCounts[models.Paused])
	assert.Equal(t, float64(0), statusCounts[models.Rewatching])

	assert.NoError(t, mock.ExpectationsWereMet())
}
