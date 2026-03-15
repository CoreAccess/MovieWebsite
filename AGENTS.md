# AGENTS.md - Enterprise Development Protocol

## 1. Development Lifecycle (Data-First)
Strict adherence to this sequence is mandatory for every feature:
1. **Schema:** Define PostgreSQL-compatible tables using the **Media Supertype** pattern.
2. **Models:** Update `internal/models` to align 1:1 with **Schema.org** specifications.
3. **Metadata:** Implement **JSON-LD** generation within the Service Layer for SEO and AI Agent readability.
4. **Logic:** Implement Repository interfaces and Service coordination.
5. **UI:** Render data provided by the Service. The UI must never influence backend logic.

## 2. N-Tier Architecture
* **HTTP Handlers (`cmd/web/`)**: Thin controllers. Handle only request/response mapping and template execution.
* **Service Layer (`internal/service/`)**: Business logic, JSON-LD payload generation, and coordination between relational and vector data.
* **Repository Interface (`internal/repository/`)**: Abstracted SQL definitions.
* **Repository Implementation (`internal/repository/dbrepo/`)**: SQL driver implementations.

## 3. Database & Scale Standards
* **Scale:** Design for 250M+ connections. Ensure referential integrity via the `media` Supertype table for all `reviews`, `cast`, and `crew` relations.
* **PostgreSQL Migration:** All SQL must be compatible with PostgreSQL. Use `SERIAL`/`IDENTITY` over `AUTOINCREMENT` and `ILIKE` for case-insensitive searches.
* **Vector Integration:** New content must trigger asynchronous jobs to compute embeddings. Semantic search results from the Vector DB must resolve back to PostgreSQL IDs.

## 4. AI & Search Readiness
* **Semantic Truth:** Every entity must serve a valid JSON-LD block in the HTML `<head>`.
* **Vector Readiness:** Repository interfaces must support ID-based resolution for semantic similarity searches.

## 5. Go Engineering Requirements
* **Dependency Injection:** Global variables (e.g., `var DB`) are prohibited. Inject all dependencies (Logger, Repo, Service) via the `application` or `service` structs.
* **Code Reuse:** Ensure all code promotes explicitness, testability, scalability, and maintainability that follow enterprise grade best practices. Abstract repeated logic into the Service Layer or internal helpers.
* **Unit Testing:** Keep `_test.go` files in the same package as the source code. All refactors must pass `go test ./...`.
* **Best Practices:** Always follow the best practices for Go development that promote explicitness, testability, scalability, and maintainability.

## 6. Prohibited Actions
* Do not bypass the Service Layer to call the Repository directly from a Handler.
* Do not introduce SQLite-specific syntax that breaks PostgreSQL compatibility.
* Do not modify database schemas to accommodate UI limitations.

## 7. Logging Requirements
* Always use Go's built-in `slog` package for logging.
* Never log sensitive information such as passwords, emails, or API keys.
* Use JSON format for logging and machine processing.
* Use structured logging to include context like request IDs and user sessions.
* Implement log rotation and retention policies.
* Add context for better request tracking and debugging.

## 8. Log Levels
* Use `slog.Debug` for debugging messages.
* Use `slog.Info` for informational messages such as application flow.
* Use `slog.Warn` for errors that are unexpected but non-critical.
* Use `slog.Error` for errors and exceptions that are considered critical.

## 9. Front-End Design Patterns
* Use HTML templates for rendering UI.
* Use Bootstrap classes as much as possible.
* Use the `ui/static` directory for static assets such as css, images and js.
* Avoid using inline styles on html elements such as div, span, etc.