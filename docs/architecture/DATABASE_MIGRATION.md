# Phase 5: Production-Ready Migration Path (PostgreSQL & Vector DB)

## 1. N-Tier Isolation
The application is currently utilizing an N-Tier architecture specifically chosen to future-proof the database schema layer:
- **HTTP Handlers (`cmd/web/`)**: Thin controllers that handle only request/response mapping and templating.
- **Service Layer (`internal/service/`)**: Business logic coordination, generating Schema.org JSON-LD payloads for SEO/AI, and preparing to map relational queries to Vector DB similarity searches.
- **Repository Interface (`internal/repository/`)**: Abstracted SQL definitions.
- **Repository Implementation (`internal/repository/dbrepo/`)**: The exact SQL driver implementation (currently SQLite).

## 2. PostgreSQL Migration Strategy
To compete with IMDb-scale datasets, the migration to PostgreSQL is paramount. With the new `DatabaseRepo` interface, the migration steps are non-destructive:
1. Create `internal/repository/dbrepo/postgres.go`
2. Define `type PostgresDBRepo struct { DB *sql.DB }`
3. Implement the `DatabaseRepo` interface.
   - *Key Change*: Swap SQLite `AUTOINCREMENT` for PostgreSQL `SERIAL` or `IDENTITY`.
   - *Key Change*: Swap SQLite `COLLATE NOCASE` for PostgreSQL `ILIKE` in search functions.
4. Update `cmd/web/main.go` to optionally load the PostgreSQL DSN via environment variables and inject `&dbrepo.PostgresDBRepo{}` instead of `SqliteDBRepo`.

## 3. Schema.org Supertype Integrity
The new database schema uses a `media` table as the Supertype for all movies, TV shows, and future episodes.
- This explicitly enforces referential integrity for bridging tables like `reviews`, `media_cast`, and `media_crew`.
- This ensures JSON-LD generation has a single source of truth for URL slugs and Aggregate Ratings.

## 4. Vector Database Integration (AI Agent & Recommendations Readiness)
Because the `AppService` encapsulates all data fetching, adding semantic search is seamless:
1. Introduce a `VectorRepo` interface (e.g., Pinecone, Milvus, pgvector).
2. Add it to the `AppService` struct.
3. On insert operations (`s.Repo.InsertMovie`), trigger an asynchronous job inside the Service to compute the embedding of the movie's `Description` and JSON-LD representation and store it in the Vector DB.
4. Handlers will simply call `app.Service.SearchSimilarMedia(query)`, which will ping the Vector DB for IDs and instantly resolve those IDs against PostgreSQL.
