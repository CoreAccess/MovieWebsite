// Package models defines the core data structures used throughout the application.
// These structs are used to map database rows into Go objects and pass data to the HTML templates for rendering.
package models

import "time"

// User represents a generalized website user.
// This struct maps directly to the `users` table in the database.
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

// Session represents a server-side session used for authenticating users.
// Instead of storing tokens, we use a database-backed session model.
type Session struct {
	ID        string    // Session ID (The value stored in the client's cookie)
	UserID    int       // Foreign key linking the session back to the specific User
	ExpiresAt time.Time // The expiration time for the session
}

// PasswordResetToken represents a password recovery token
type PasswordResetToken struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
}

// Organization represents schema.org/Organization (Studio, Network)
type Organization struct {
	ID           int
	Name         string
	Slug         string
	Description  string
	Image        string
	Logo         string
	Url          string
	FoundingDate string
}

// Person represents schema.org/Person (Actor, Director)
type Person struct {
	ID            int
	Name          string
	Slug          string
	Gender        string
	BirthDate     string
	BirthPlace    string
	DeathDate     string
	Height        string
	Description   string
	Image         string
	AlsoKnownAs   string
	Awards        string
	KnowsLanguage string
}

// Character represents schema.org/Person (Fictional)
type Character struct {
	ID          int
	Name        string
	Slug        string
	Gender      string
	BirthDate   string
	DeathDate   string
	Description string
	Image       string
}

// Movie represents schema.org/Movie.
// It corresponds to the core details found in the `movies` table in the database.
// This structure is often passed directly to templates to display movie listings or specific movie metadata.
type Movie struct {
	ID                int
	Name              string
	Slug              string // URL-friendly string representing the movie name
	DatePublished     string
	Description       string
	Image             string // URL or path to the poster image
	Trailer           string
	Video             string
	ContentRating     string
	Duration          int     // Runtime length of the movie
	AggregateRating   float64 // Average score given by users
	Genre             string  // Stored as JSON string representation (e.g., '["Action", "Drama"]')
	Budget            string
	BoxOffice         string
	InLanguage        string
	ProductionCompany string
	Keywords          string
}

// TVSeries represents schema.org/TVSeries
type TVSeries struct {
	ID                int
	Name              string
	Slug              string
	StartDate         string
	EndDate           string
	Description       string
	Image             string
	ContentRating     string
	AggregateRating   float64
	NumberOfSeasons   int
	NumberOfEpisodes  int
	Trailer           string
	Genre             string // JSON string representation
	ProductionCompany string
	InLanguage        string
}

// TVEpisode represents schema.org/TVEpisode
type TVEpisode struct {
	ID            int
	SeriesID      int // Foreign Key
	SeasonNumber  int
	EpisodeNumber int
	Name          string
	Slug          string
	DatePublished string
	Description   string
	Image         string
	Duration      int
}

// CastMember represents an actor playing a character in a specific media
type CastMember struct {
	Person       Person
	Character    Character
	BillingOrder int
}

// MovieDetail represents an aggregated view of a movie profile including its core details, cast, and crew.
// It is specifically designed to supply all necessary data to the `movies.html` template in a single object.
// The `database.GetMovieDetail` function constructs this by joining several underlying tables together.
type MovieDetail struct {
	Movie     Movie        // The core movie data
	Cast      []CastMember // List of actors and the characters they played
	Directors []Person     // List of individuals credited as directors
	Writers   []Person     // List of individuals credited as writers
}

// TVSeriesDetail represents a complete TV show profile including cast and episodes
type TVSeriesDetail struct {
	Series    TVSeries
	Cast      []CastMember
	Episodes  []TVEpisode
	Directors []Person
	Writers   []Person
}

// PersonDetail represents a complete person profile
type PersonDetail struct {
	Person Person
	Movies []Movie
	Shows  []TVSeries
}

// Watchlist represents a user's collection of media
type Watchlist struct {
	ID          int
	UserID      int
	Name        string
	Description string
	IsPublic    bool
	CreatedAt   time.Time
}

// WatchlistItem represents an item in a watchlist
type WatchlistItem struct {
	ID          int
	WatchlistID int
	MediaType   string // "movie" or "tv"
	MediaID     int
	AddedAt     time.Time
}

// Achievement represents a gamification goal
type Achievement struct {
	ID          int
	Name        string
	Description string
	Points      int
	BadgeImage  string
}

// UserAchievement maps users to completed achievements
type UserAchievement struct {
	UserID        int
	AchievementID int
	EarnedAt      time.Time
}

// Notification represents a system or social alert
type Notification struct {
	ID        int
	UserID    int
	Message   string
	Link      string
	IsRead    bool
	CreatedAt time.Time
}

// UserNotificationSettings stores user preferences for alerts
type UserNotificationSettings struct {
	UserID      bool // changed from int to avoid error? wait!
	EmailAlerts bool
	SiteAlerts  bool
	Mentions    bool
}

// EditHistory tracks wiki edits
type EditHistory struct {
	ID         int
	UserID     int
	EntityType string // "movie", "tv", "person"
	EntityID   int
	Field      string
	OldValue   string
	NewValue   string
	EditedAt   time.Time
}

// EditSuggestion for lower-reputation users (needs moderation)
type EditSuggestion struct {
	ID          int
	UserID      int
	EntityType  string
	EntityID    int
	Field       string
	NewValue    string
	Status      string // "pending", "approved", "rejected"
	SubmittedAt time.Time
}

// Advertisement represents an ad creative
type Advertisement struct {
	ID          int
	CampaignID  int
	Image       string
	Url         string
	Title       string
	Description string
}

// AdCampaign represents a monetization campaign
type AdCampaign struct {
	ID          int
	CompanyID   int
	Budget      float64
	TargetPages string // JSON array of slugs/categories
	Impressions int
	Clicks      int
	StartDate   time.Time
	EndDate     time.Time
}

// Post represents a social feed post
type Post struct {
	ID        int
	UserID    int
	Content   string
	CreatedAt time.Time
}

// Comment on a post or review
type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time
}

// Like on a post/comment
type Like struct {
	UserID int
	PostID int
}

// Poll for social engagements
type Poll struct {
	ID       int
	PostID   int
	Question string
	Options  string // JSON array
}

// EbayListing represents an affiliate merchandise item
type EbayListing struct {
	Title    string
	Price    string
	Url      string
	ImageUrl string
	IsHot    bool
}
