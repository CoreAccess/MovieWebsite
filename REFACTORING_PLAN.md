# **Project Restructuring & Enterprise Refactoring Plan**

## **1\. Objective**

Transform the existing "script-style" Go web application into a **Layered Architecture** utilizing **Dependency Injection** and **Interface-based Repositories**. This will eliminate global variables, enable unit testing, and improve scalability for an IMDb-scale platform.

## **2\. Target Directory Structure**

The project will follow the **Standard Go Project Layout**.

Plaintext  
├── cmd/  
│   └── web/  
│       ├── main.go           \# Wiring and Startup  
│       ├── routes.go         \# Centralized Routing  
│       ├── handlers.go       \# HTTP Handlers (Method Receivers)  
│       ├── middleware.go     \# Global & Route Middleware  
│       └── helpers.go        \# UI Helpers (Rendering, Error handling)  
├── internal/  
│   ├── config/               \# Configuration and Environment  
│   ├── models/               \# Shared Data Structures (DB-agnostic)  
│   ├── repository/           \# Data Access Interfaces  
│   │   └── dbrepo/           \# SQLite Implementation of Repository  
│   ├── service/              \# Business Logic (Monetization, TMDB logic)  
│   └── tmdb/                 \# External API Client  
├── ui/                       \# Unchanged (Static/HTML)  
├── go.mod  
└── .env                      \# Secrets (TMDB Key, Port)

## **3\. Implementation Phases**

### **Phase 1: Dependency Injection Setup**

1. **Define the `application` struct** in `cmd/web/main.go`. This struct must hold all dependencies:  
   * `Logger` (Structured logging)  
   * `Repository` (The DB interface)  
   * `TemplateCache` (Parsed HTML map for performance)  
   * `Config` (Env variables)  
2. **Remove Global Variables:** Delete `var DB *sql.DB` from `internal/database/db.go`.  
3. **Environment Configuration:** Move the hardcoded TMDB API key to a `.env` file and load it using a config package.

### **Phase 2: Repository Pattern (Data Layer)**

1. **Create Repository Interfaces:** In `internal/repository/repository.go`, define an interface (e.g., `DatabaseRepo`) that includes methods like `GetMovieByID(id int)`, `GetAllMovies()`, etc.  
2. **Implement SQLite Repo:** Move the raw SQL logic from `internal/database/` into `internal/repository/dbrepo/sqlite.go`.  
3. **Dependency Injection:** Pass the `*sql.DB` connection into the repository constructor rather than relying on a package-level variable.

### **Phase 3: Routing & Middleware Centralization**

1. **Create `routes.go`:** Move all `mux.HandleFunc` calls from `main()` into an `(app *application) routes() http.Handler` method.  
2. **Clean up Middleware:** Move `logRequest`, `recoverPanic`, and `secureHeaders` into the `cmd/web/middleware.go` file as methods of the `application` struct.

### **Phase 4: Handler Refactoring**

1. **Method Receivers:** Change all handler functions (e.g., `movieView`) to be methods of the `application` struct: `func (app *application) movieView(...)`.  
2. **Centralized Rendering:** Create a `render` helper in `helpers.go` that handles template parsing from a cache rather than calling `template.ParseFiles` on every request.  
3. **Abstract Business Logic:** Move complex logic (like combining movie data with eBay listings) out of the handler and into a dedicated `service` package.

## **4\. Specific Refactoring Tasks for the Agent**

* **Struct Optimization:** Refactor the `templateData` struct in `main.go` to be more modular.  
* **Error Handling:** Implement centralized error helpers (`app.serverError`, `app.notFound`) to ensure consistent HTTP responses.  
* **Template Caching:** Implement a `templateCache` during the application startup to prevent disk I/O on every page load.

