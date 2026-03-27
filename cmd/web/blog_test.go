package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"filmgap/internal/config"
	"filmgap/internal/models"
	"filmgap/internal/repository/dbrepo"
)

// TestSeedBlogData is a "test" that seeds mock blog posts for development.
// Run with: go test -v -run TestSeedBlogData cmd/web/blog_test.go cmd/web/helpers.go cmd/web/routes.go cmd/web/handlers.go cmd/web/middleware.go cmd/web/ratelimit.go cmd/web/auth.go cmd/web/profile.go cmd/web/list_handlers.go cmd/web/admin.go
// Actually, it's easier to just use the main app's dependencies.
func TestSeedBlogData(t *testing.T) {
	config.LoadEnv("c:/Users/adamd/Downloads/Programming/Design_Web_Experiment/.env")

	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	pgRepo := &dbrepo.PostgresDBRepo{}
	db, err := pgRepo.InitDB(dsn, "")
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	posts := []models.BlogPost{
		{
			Title:      "The Rise of Neo-Noir in Modern Cinema",
			Slug:       "rise-of-neo-noir",
			Content:    "## The Shadows Return\n\nNeo-noir is more than just a style; it's a reflection of our modern anxieties. From the rain-slicked streets of *John Wick* to the atmospheric dread of *Blade Runner 2049*, filmmakers are finding new ways to evolve the classic hardboiled aesthetic.\n\n### Key Characteristics\n- **Ambiguous Morality**: No one is purely a hero.\n- **Visual Contrast**: High-contrast lighting and neon accents.\n- **Urban Decay**: The city as a character.\n\nIn this article, we explore how directors like Denis Villeneuve and Chad Stahelski are redefining the genre for a new generation.",
			Image:      "https://images.unsplash.com/photo-1614850523296-d8c1af93d400?ixlib=rb-1.2.1&auto=format&fit=crop&w=1200&q=80",
			AuthorID:   1, // Assuming admin or first user
			IsFeatured: true,
		},
		{
			Title:      "Why Practical Effects Still Rule the Box Office",
			Slug:       "practical-effects-vs-cgi",
			Content:    "## Beyond the Green Screen\n\nWhile CGI has reached mind-blowing levels of realism, there's an undeniable weight to practical effects. Think of the *Mad Max: Fury Road* car chases or the tangible fear in *The Thing*.\n\n### Why it Matters\nActors react better to real objects. The lighting is more natural. The audience feels the stakes.\n\n> \"I want to see something real on set. It changes the energy entirely.\"\n\nWe sit down with veteran SFX artists to discuss the future of the craft.",
			Image:      "https://images.unsplash.com/photo-1485846234645-a62644f84728?ixlib=rb-1.2.1&auto=format&fit=crop&w=1200&q=80",
			AuthorID:   1,
			IsFeatured: false,
		},
		{
			Title:      "10 Must-Watch Indie Gems from 2025",
			Slug:       "indie-gems-2025",
			Content:    "## Support Independent Cinema\n\nThe blockbusters might get the billboard space, but the soul of cinema lives in the indie circuit. This year has seen some incredible debuts.\n\n1. **The Silent Echo**: A haunting family drama.\n2. **Neon Horizon**: A lo-fi sci-fi masterpiece.\n3. **Waves of Change**: A powerful documentary on climate action.\n\nRead more about where to stream these masterpieces.",
			Image:      "https://images.unsplash.com/photo-1536440136628-849c177e76a1?ixlib=rb-1.2.1&auto=format&fit=crop&w=1200&q=80",
			AuthorID:   1,
			IsFeatured: false,
		},
	}

	for _, p := range posts {
		err := pgRepo.CreateBlogPost(p)
		if err != nil {
			log.Printf("Failed to create post %s: %v", p.Title, err)
		} else {
			fmt.Printf("Created post: %s\n", p.Title)
		}
	}
}
