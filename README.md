# MovieWeb

MovieWeb is a Go-based web application for managing and exploring a media database of movies, TV shows, and people. It features user accounts, watchlists, social feeds, wiki-style editing, admin moderation, and monetization hooks (such as eBay affiliate listings and ad campaigns).

## Features

- **Media Database:** Browse, search, and view detailed profiles for movies, TV shows, and cast/crew members.
- **User Accounts:** Sign up, log in, manage profiles, and maintain a personalized watchlist.
- **Social & Engagement:** View a personalized feed, post updates, and engage with community polls and comments.
- **Community Wiki:** Users can suggest edits to media entries, which administrators can approve or reject.
- **Monetization:** Built-in support for displaying advertisements, managing ad campaigns, and integrating affiliate links (e.g., eBay listings).
- **Security:** Built-in CSRF protection using `nosurf`.

## Tech Stack

- **Backend:** Go (1.26.1)
- **Database:** SQLite (`modernc.org/sqlite`)
- **Frontend:** HTML Templates, CSS, JavaScript (served from `ui/`)
- **Security:** `github.com/justinas/nosurf` for CSRF protection

## Project Structure

- `cmd/web/`: Contains the main application entry point and HTTP handlers.
- `internal/database/`: Database initialization and query functions.
- `internal/models/`: Go structs defining the application's data models.
- `internal/monetization/`: Logic for handling ads and affiliate listings.
- `internal/tmdb/`: Client for interacting with the TMDB API.
- `ui/`: Contains HTML templates (`ui/html/`) and static assets (`ui/static/`).
- `DATABASE_RESTRUCTURING_PLAN.md`: Documentation for planned future schema updates and monetization strategies.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.26 or later.
- A valid TMDB API key (configured in `cmd/web/main.go` or via environment variables).

### Running the Application

1. Clone the repository and navigate to the project root:
   ```bash
   git clone <repository-url>
   cd movieweb
   ```

2. Download Go module dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run ./cmd/web
   ```

4. The server will start on `http://localhost:8080` by default. You can override the port by setting the `PORT` environment variable:
   ```bash
   PORT=9000 go run ./cmd/web
   ```

## Future Roadmap

A comprehensive restructuring of the database is planned to normalize schemas, integrate deeper monetization (affiliate tracking), and adopt schema.org structured data. See the `DATABASE_RESTRUCTURING_PLAN.md` file for full details.
