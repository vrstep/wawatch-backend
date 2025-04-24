package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock" // Import api package
	"github.com/vrstep/wawatch-backend/models"
)

// Mock AniList Client (Place this at the top or in a helper)
type MockAniListClient struct {
	mock.Mock
}

func (m *MockAniListClient) GetAnimeByID(id int) (*models.AnimeDetails, error) {
	args := m.Called(id)
	var details *models.AnimeDetails
	if args.Get(0) != nil {
		details = args.Get(0).(*models.AnimeDetails)
	}
	return details, args.Error(1)
}

func (m *MockAniListClient) SearchAnime(query string, page int, perPage int) ([]models.AnimeCache, int, error) {
	args := m.Called(query, page, perPage)
	var animes []models.AnimeCache
	if args.Get(0) != nil {
		animes = args.Get(0).([]models.AnimeCache)
	}
	return animes, args.Int(1), args.Error(2)
}

func (m *MockAniListClient) GetPopularAnime(page int, perPage int) ([]models.AnimeCache, int, error) {
	args := m.Called(page, perPage)
	var animes []models.AnimeCache
	if args.Get(0) != nil {
		animes = args.Get(0).([]models.AnimeCache)
	}
	return animes, args.Int(1), args.Error(2)
}

func (m *MockAniListClient) GetTrendingAnime(page int, perPage int) ([]models.AnimeCache, int, error) {
	args := m.Called(page, perPage)
	var animes []models.AnimeCache
	if args.Get(0) != nil {
		animes = args.Get(0).([]models.AnimeCache)
	}
	return animes, args.Int(1), args.Error(2)
}

func (m *MockAniListClient) GetAnimeBySeason(year int, season string, page int, perPage int) ([]models.AnimeCache, int, error) {
	args := m.Called(year, season, page, perPage)
	var animes []models.AnimeCache
	if args.Get(0) != nil {
		animes = args.Get(0).([]models.AnimeCache)
	}
	return animes, args.Int(1), args.Error(2)
}

// Test GetPopularAnime Endpoint
func TestGetPopularAnime(t *testing.T) {
	// Setup Mock API Client
	mockAPI := new(MockAniListClient)
	SetAniListClient(mockAPI) // Inject mock

	// Setup Gin
	router := SetupGin()

	// Mock API Response
	page := 1
	perPage := 5
	total := 10
	mockResults := []models.AnimeCache{
		{ID: 1, Title: "Popular Anime 1"},
		{ID: 2, Title: "Popular Anime 2"},
	}
	mockAPI.On("GetPopularAnime", page, perPage).Return(mockResults, total, nil)

	// Setup Route
	router.GET("/anime/popular", GetPopularAnime)

	// Perform Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/anime/popular?page="+strconv.Itoa(page)+"&perPage="+strconv.Itoa(perPage), nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	// Check meta
	meta := responseBody["meta"].(map[string]interface{})
	assert.Equal(t, float64(total), meta["total"])
	assert.Equal(t, float64(page), meta["page"])
	assert.Equal(t, float64(perPage), meta["perPage"])
	assert.Equal(t, float64(2), meta["totalPages"]) // (10 + 5 - 1) / 5 = 2
	assert.True(t, meta["hasNextPage"].(bool))      // 1*5 < 10

	// Check data
	data := responseBody["data"].([]interface{})
	assert.Len(t, data, len(mockResults))
	firstItem := data[0].(map[string]interface{})
	assert.Equal(t, float64(mockResults[0].ID), firstItem["id"])
	assert.Equal(t, mockResults[0].Title, firstItem["title"])

	// Verify mock expectations
	mockAPI.AssertExpectations(t)
}

// Test GetTrendingAnime Endpoint (Similar structure to GetPopularAnime)
func TestGetTrendingAnime(t *testing.T) {
	mockAPI := new(MockAniListClient)
	SetAniListClient(mockAPI)
	router := SetupGin()

	page, perPage, total := 1, 3, 7
	mockResults := []models.AnimeCache{{ID: 10, Title: "Trending 1"}}
	mockAPI.On("GetTrendingAnime", page, perPage).Return(mockResults, total, nil)

	router.GET("/anime/trending", GetTrendingAnime)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/anime/trending?page="+strconv.Itoa(page)+"&perPage="+strconv.Itoa(perPage), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// ... add more assertions for body content ...
	mockAPI.AssertExpectations(t)
}

// Test GetAnimeBySeason Endpoint
func TestGetAnimeBySeason(t *testing.T) {
	mockAPI := new(MockAniListClient)
	SetAniListClient(mockAPI)
	router := SetupGin()

	year, season := 2024, "SPRING"
	page, perPage, total := 1, 2, 5
	mockResults := []models.AnimeCache{{ID: 20, Title: "Spring Anime"}}
	mockAPI.On("GetAnimeBySeason", year, season, page, perPage).Return(mockResults, total, nil)

	router.GET("/anime/season/:year/:season", GetAnimeBySeason)

	w := httptest.NewRecorder()
	reqURL := "/anime/season/" + strconv.Itoa(year) + "/" + strings.ToLower(season) + "?page=" + strconv.Itoa(page) + "&perPage=" + strconv.Itoa(perPage)
	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// ... add more assertions for body content ...
	mockAPI.AssertExpectations(t)
}

// Test GetAnimeRecommendations Endpoint (Uses GetPopularAnime mock for now)
func TestGetAnimeRecommendations(t *testing.T) {
	mockAPI := new(MockAniListClient)
	SetAniListClient(mockAPI)
	router := SetupGin()

	page, perPage, total := 1, 10, 20 // Default perPage is 10 for recommendations
	mockResults := []models.AnimeCache{{ID: 30, Title: "Recommended Anime"}}
	// Mocking GetPopularAnime as it's the placeholder
	mockAPI.On("GetPopularAnime", page, perPage).Return(mockResults, total, nil)

	router.GET("/anime/recommendations", func(c *gin.Context) {
		// Simulate auth if needed by actual implementation
		// c.Set("user", models.User{Model: gorm.Model{ID: 1}})
		GetAnimeRecommendations(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/anime/recommendations?page="+strconv.Itoa(page)+"&perPage="+strconv.Itoa(perPage), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// ... add more assertions for body content ...
	mockAPI.AssertExpectations(t)
}
