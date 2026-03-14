# MovieWeb

MovieWeb is a Go-based web application for managing and exploring a media database of movies, TV shows, and people. It features user accounts, watchlists, social feeds, and admin moderation.

## Features

- **Media Database:** Browse, search, and view detailed profiles for movies, TV shows, and cast/crew members.
- **User Accounts:** Sign up, log in, manage profiles, and maintain a personalized watchlist.
- **Social & Engagement:** View a personalized feed, post updates, and engage with community polls and comments.

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

- `internal/tmdb/`: Client for interacting with the TMDB API.
- `ui/`: Contains HTML templates (`ui/html/`) and static assets (`ui/static/`).
- `DATABASE_RESTRUCTURING_PLAN.md`: Documentation for planned future schema updates.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.26 or later.
- A valid TMDB API key (configured via the `TMDB_API_KEY` environment variable).

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

5. (Important) Set your TMDB API Read Access Token as an environment variable to avoid 401 errors:
   - On Windows (PowerShell): `$env:TMDB_API_KEY="your_token_here"`
   - On Windows (CMD): `set TMDB_API_KEY=your_token_here`
   - On Linux/macOS: `export TMDB_API_KEY="your_token_here"`

6. The server will start on `http://localhost:8080` by default. You can override the port by setting the `PORT` environment variable:
   ```bash
   PORT=9000 TMDB_API_KEY="your_token_here" go run ./cmd/web
   ```

## Future Roadmap

A comprehensive restructuring of the database is planned to normalize schemas and adopt schema.org structured data. See the `DATABASE_RESTRUCTURING_PLAN.md` file for full details.
