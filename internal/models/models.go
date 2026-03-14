package models

import "time"

// Data structs derived from Schema.org for maximum AI Agent/SEO compatibility.

// User represents a system user. (Retained from previous schema)
type User struct {
	ID              int
	Username        string
	Email           string
	PasswordHash    string
	GoogleID        string
	FacebookID      string
	Avatar          string // URL or path to the user's profile picture
	ReputationScore int    // Score based on community engagement and contributions
	Role            string // Represents user privileges (e.g., 'user', 'contributor', 'moderator', 'admin')
	CreatedAt       time.Time
}

// Media is the core "Supertype" mapping to Schema.org's CreativeWork, Movie, TVSeries.
// In the database, all subtypes (movies, tv_series) will reference `media.id`.
type Media struct {
	ID              int       `json:"-"`
	MediaType       string    `json:"@type"` // "Movie" or "TVSeries"
	Name            string    `json:"name"`
	Slug            string    `json:"url"`
	Description     string    `json:"description,omitempty"`
	Image           string    `json:"image,omitempty"`
	DatePublished   string    `json:"datePublished,omitempty"` // Maps to release_date or first_air_date
	AggregateRating float64   `json:"aggregateRating,omitempty"`
	CreatedAt       time.Time `json:"-"`
}

// Movie is a subtype of Media.
type Movie struct {
	Media          // Embeds the base Media fields
	ReleaseDate    string  `json:"-"` // Maps to DatePublished internally
	Duration        int     `json:"duration,omitempty"` // In minutes (will be converted to ISO 8601 for JSON-LD)
	ContentRating    string
	TmdbID           int     `json:"-"`
	Genres           []Genre `json:"genre,omitempty"`
	RatingCount      int
	ReviewCount      int
	BestRating       float64
	WorstRating      float64
	IsFamilyFriendly bool
	Budget           string
	BoxOffice        string
	LanguageCode     string
	CountryCode      string
	Tagline          string
	Subtitle string
}

// TVSeries is a subtype of Media.
type TVSeries struct {
	Media            // Embeds the base Media fields
	StartDate        string `json:"-"` // Maps to DatePublished
	EndDate          string `json:"endDate,omitempty"`
	NumberOfSeasons  int    `json:"numberOfSeasons,omitempty"`
	NumberOfEpisodes int    `json:"numberOfEpisodes,omitempty"`
	ContentRating    string
	TmdbID           int    `json:"-"`
	Genres           []Genre `json:"genre,omitempty"`
	RatingCount      int
	ReviewCount      int
	BestRating       float64
	WorstRating      float64
	IsFamilyFriendly bool
	LanguageCode     string
	CountryCode      string
	Tagline string
	Subtitle string
}

// Person represents a cast or crew member (Schema.org Person).
type Person struct {
	ID              int       `json:"-"`
	Name            string    `json:"name"`
	Slug            string    `json:"url"`
	Biography       string    `json:"description,omitempty"`
	Image           string    `json:"image,omitempty"`
	Gender          string    `json:"gender,omitempty"`
	BirthPlace    string    `json:"birthPlace,omitempty"`
	Birthday        string    `json:"birthDate,omitempty"`
	Deathday        string    `json:"deathDate,omitempty"`
	BirthDate       string
	Description     string
	PopularityScore float64
	KnowsLanguage   string
	NationalityCode string
	KnownForDepartment string
	TmdbID          int       `json:"-"`
	KnownFor        string    `json:"jobTitle,omitempty"`
}

// CastMember represents a role a Person played in Media.
type CastMember struct {
	Person
	Character Person `json:"characterName"`
	BillingOrder int
	Order         int    `json:"-"`
}

// CrewMember represents a behind-the-scenes role a Person played in Media.
type CrewMember struct {
	Person
	Job        string `json:"jobTitle"`
	Department string `json:"-"`
}

// Organization represents a production company or similar entity (Schema.org Organization).
type Organization struct {
	ID            int       `json:"-"`
	Name          string    `json:"name"`
	Slug          string    `json:"url"`
	Description   string    `json:"description,omitempty"`
	LogoURL       string    `json:"logo,omitempty"`
	OriginCountry string    `json:"location,omitempty"`
	TmdbID        int       `json:"-"`
}

// Supporting structures
type Watchlist struct {
	ID        int
	UserID    int
	Name      string
	Description string
	IsPublic  bool
	CreatedAt time.Time
}

type WatchlistItem struct {
	ID          int
	WatchlistID int
	MediaID     int
	MediaType   string
	AddedAt     time.Time
}

type Post struct {
	ID        int
	UserID    int
	Content   string
	CreatedAt time.Time
	Author    *User
}

// Missing types required for compilation based on previous struct changes
type Session struct {
	ID        string
	UserID    int
	Expires   time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
}



type Genre struct {
	ID   int
	Name string
	Slug string
}

type TVEpisode struct {
	ID            int       `json:"-"`
	SeriesID      int       `json:"-"`
	SeasonNumber  int       `json:"seasonNumber"`
	EpisodeNumber int       `json:"episodeNumber"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	AirDate       string    `json:"datePublished,omitempty"`
	DatePublished string
	Slug          string
	Image         string
	Duration      int
}

type MovieDetail struct {
	Movie
	Cast       []CastMember
	Directors  []Person
	Writers    []Person
	Genres     []Genre
}

type TVSeriesDetail struct {
	TVSeries
	Series     TVSeries
	Cast       []CastMember
	Creators   []Person
	Genres     []Genre
	Episodes   []TVEpisode
	Directors  []Person
	Writers    []Person
}

type PersonDetail struct {
	Person
	Movies []Movie
	Shows  []TVSeries
}
