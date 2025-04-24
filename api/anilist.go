package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vrstep/wawatch-backend/models"
)

const (
	AniListURL = "https://graphql.anilist.co"
	// Default timeout for requests in seconds
	DefaultTimeout = 10
)

type AniListClient struct {
	httpClient *http.Client
}

// NewAniListClient creates a new client for interacting with AniList API
func NewAniListClient() *AniListClient {
	return &AniListClient{
		httpClient: &http.Client{
			Timeout: time.Second * DefaultTimeout,
		},
	}
}

// GetAnimeByID fetches anime details from AniList by ID
func (c *AniListClient) GetAnimeByID(id int) (*models.AnimeDetails, error) {
	query := `
    query ($id: Int) {
        Media(id: $id, type: ANIME) {
            id
            title {
                romaji
                english
                native
            }
            description
            format
            status
            episodes
            duration
            genres
            startDate {
                year
                month
                day
            }
            endDate {
                year
                month
                day
            }
            season
            seasonYear
            coverImage {
                large
                medium
            }
            bannerImage
            averageScore
            popularity
            studios {
                nodes {
                    name
                }
            }
        }
    }
    `

	variables := map[string]interface{}{
		"id": id,
	}

	// Execute the query
	response, err := c.executeQuery(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch anime: %v", err)
	}

	// Parse the response
	var result struct {
		Data struct {
			Media *models.AnimeDetails `json:"Media"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse anime data: %v", err)
	}

	return result.Data.Media, nil
}

// SearchAnime performs a search query on AniList
func (c *AniListClient) SearchAnime(query string, page int, perPage int) ([]models.AnimeCache, int, error) {
	gqlQuery := `
    query ($search: String, $page: Int, $perPage: Int) {
        Page(page: $page, perPage: $perPage) {
            pageInfo {
                total
                currentPage
                lastPage
                hasNextPage
            }
            media(search: $search, type: ANIME, sort: POPULARITY_DESC) {
                id
                title {
                    romaji
                    english
                    native
                }
                coverImage {
                    large
                    medium
                }
                format
                episodes
            }
        }
    }
    `

	variables := map[string]interface{}{
		"search":  query,
		"page":    page,
		"perPage": perPage,
	}

	response, err := c.executeQuery(gqlQuery, variables)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Data struct {
			Page struct {
				PageInfo struct {
					Total int `json:"total"`
				} `json:"pageInfo"`
				Media []struct {
					ID    int `json:"id"`
					Title struct {
						Romaji  string `json:"romaji"`
						English string `json:"english"`
					} `json:"title"`
					CoverImage struct {
						Large string `json:"large"`
					} `json:"coverImage"`
					Format   string `json:"format"`
					Episodes int    `json:"episodes"`
				} `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, 0, fmt.Errorf("failed to parse search results: %v", err)
	}

	// Convert to AnimeCache objects
	animes := make([]models.AnimeCache, len(result.Data.Page.Media))
	for i, media := range result.Data.Page.Media {
		title := media.Title.English
		if title == "" {
			title = media.Title.Romaji
		}

		episodes := media.Episodes
		animes[i] = models.AnimeCache{
			ID:            media.ID,
			Title:         title,
			CoverImage:    media.CoverImage.Large,
			Format:        media.Format,
			TotalEpisodes: &episodes,
		}
	}

	return animes, result.Data.Page.PageInfo.Total, nil
}

// executeQuery handles the execution of GraphQL queries to AniList
func (c *AniListClient) executeQuery(query string, variables map[string]interface{}) ([]byte, error) {
	// Prepare the request body
	reqBody, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	// Create the request
	req, err := http.NewRequest("POST", AniListURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anilist API returned status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

// Helper function to execute paged media queries
func (c *AniListClient) executePagedMediaQuery(query string, variables map[string]interface{}) ([]models.AnimeCache, int, error) {
	response, err := c.executeQuery(query, variables)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Data struct {
			Page struct {
				PageInfo struct {
					Total int `json:"total"`
				} `json:"pageInfo"`
				Media []struct {
					ID    int `json:"id"`
					Title struct {
						Romaji  string `json:"romaji"`
						English string `json:"english"`
					} `json:"title"`
					CoverImage struct {
						Large string `json:"large"`
					} `json:"coverImage"`
					Format   string `json:"format"`
					Episodes *int   `json:"episodes"` // Use pointer for nullable
				} `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, 0, fmt.Errorf("failed to parse paged media results: %v", err)
	}

	// Convert to AnimeCache objects
	animes := make([]models.AnimeCache, len(result.Data.Page.Media))
	for i, media := range result.Data.Page.Media {
		title := media.Title.English
		if title == "" {
			title = media.Title.Romaji
		}

		animes[i] = models.AnimeCache{
			ID:            media.ID,
			Title:         title,
			CoverImage:    media.CoverImage.Large,
			Format:        media.Format,
			TotalEpisodes: media.Episodes, // Assign pointer directly
		}
	}

	return animes, result.Data.Page.PageInfo.Total, nil
}

// GetPopularAnime fetches popular anime
func (c *AniListClient) GetPopularAnime(page int, perPage int) ([]models.AnimeCache, int, error) {
	gqlQuery := `
    query ($page: Int, $perPage: Int) {
        Page(page: $page, perPage: $perPage) {
            pageInfo { total currentPage lastPage hasNextPage }
            media(type: ANIME, sort: POPULARITY_DESC) {
                id
                title { romaji english }
                coverImage { large }
                format
                episodes
            }
        }
    }`
	variables := map[string]interface{}{
		"page":    page,
		"perPage": perPage,
	}
	return c.executePagedMediaQuery(gqlQuery, variables)
}

// GetTrendingAnime fetches trending anime
func (c *AniListClient) GetTrendingAnime(page int, perPage int) ([]models.AnimeCache, int, error) {
	gqlQuery := `
    query ($page: Int, $perPage: Int) {
        Page(page: $page, perPage: $perPage) {
            pageInfo { total currentPage lastPage hasNextPage }
            media(type: ANIME, sort: TRENDING_DESC) {
                id
                title { romaji english }
                coverImage { large }
                format
                episodes
            }
        }
    }`
	variables := map[string]interface{}{
		"page":    page,
		"perPage": perPage,
	}
	return c.executePagedMediaQuery(gqlQuery, variables)
}

// GetAnimeBySeason fetches anime by year and season
func (c *AniListClient) GetAnimeBySeason(year int, season string, page int, perPage int) ([]models.AnimeCache, int, error) {
	gqlQuery := `
    query ($page: Int, $perPage: Int, $season: MediaSeason, $seasonYear: Int) {
        Page(page: $page, perPage: $perPage) {
            pageInfo { total currentPage lastPage hasNextPage }
            media(type: ANIME, season: $season, seasonYear: $seasonYear, sort: POPULARITY_DESC) {
                id
                title { romaji english }
                coverImage { large }
                format
                episodes
            }
        }
    }`
	variables := map[string]interface{}{
		"page":       page,
		"perPage":    perPage,
		"season":     season, // WINTER, SPRING, SUMMER, FALL
		"seasonYear": year,
	}
	return c.executePagedMediaQuery(gqlQuery, variables)
}
