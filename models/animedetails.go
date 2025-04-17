package models

// AnimeDetails represents comprehensive information about an anime
type AnimeDetails struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Description string   `json:"description"`
	Format      string   `json:"format"` // TV, MOVIE, OVA, etc.
	Status      string   `json:"status"` // FINISHED, RELEASING, etc.
	Episodes    int      `json:"episodes"`
	Duration    int      `json:"duration"` // Per episode in minutes
	Genres      []string `json:"genres"`
	StartDate   struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"startDate"`
	EndDate struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"endDate"`
	Season     string `json:"season"`
	SeasonYear int    `json:"seasonYear"`
	CoverImage struct {
		Large  string `json:"large"`
		Medium string `json:"medium"`
	} `json:"coverImage"`
	BannerImage  string `json:"bannerImage"`
	AverageScore int    `json:"averageScore"`
	Popularity   int    `json:"popularity"`
	Studios      struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"studios"`
}

// ToAnimeCache converts detailed anime info to a cache entry
func (a *AnimeDetails) ToAnimeCache() AnimeCache {
	return AnimeCache{
		ID:            a.ID,
		Title:         a.Title.English,
		CoverImage:    a.CoverImage.Large,
		Format:        a.Format,
		TotalEpisodes: &a.Episodes,
	}
}
