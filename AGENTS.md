## PART 1: FRONT-END & DESIGN STANDARDS

### 1.1 Architecture & Layout Philosophy

- **Flexible Box-Based Architecture:** The website focuses on box-based layouts where content collapses to the next row and expands seamlessly as the browser window shrinks, preventing the need to "hide" content on smaller devices. We embrace design flexibility: boxes will vary in size, column counts (1 or 2), and styling (some will have borders, others won't).
- **Enterprise-Light Aesthetic:** Crisp white backgrounds, bright high-contrast text, minimalist modern layouts. Explicitly rejecting dark-mode-first or atmospheric designs in favor of maximum text readability.
- **Symmetry & Balance:** Margins and padding (`padding-inline`) must be perfectly symmetrical on both axes to maintain equilibrium across viewports.
- **Fluid Scaling:** Use CSS `clamp()` for continuous, liquid typography scaling rather than relying purely on rigid media queries.

### 1.2 Engineering Standards & Accessibility

- **Semantic HTML5:** Strict hierarchy. Exactly one `<h1>` per view. Never skip heading levels (e.g., `<h2>` to `<h4>`). The DOM structure must tell a logical story without CSS.
- **A11y (Accessibility) Level AA:** Interactive target minimum size must be **24x24 CSS pixels**. Use `aria-live="polite"` for dynamic content updates so users aren't interrupted.
- **Vanilla JS Modularity:** ES6 modules only. Strict separation of concerns via encapsulation. No global variables.
- **Event Delegation:** High-density UI elements must use a single event listener on the parent container, identifying targets via `event.target.closest()`.
- **Performance (LCP Optimized):**
    - Primary "Hero" assets must use `fetchpriority="high"`.
    - All off-screen media must use `loading="lazy"`. AVIF and WebP formats prioritized via `<picture>` fallback tags.
- **Localhost Web Server:** The web server must run on `http://localhost:8080/`. Port `8080` is the default port.
- **Masterplan:** Always consule the `MASTERPLAN_00_INDEX` file in the `tmp` folder and keep it updated as the project evolves.

### 1.3 Design System & Constraints

- **Typography Engine:** The platform relies on the **System Font Stack** (`system-ui, -apple-system, BlinkMacSystemFont, Roboto`) for instantaneous rendering on 90% of UI elements, occasionally paired with a single downloaded hero font.
- **Iconography:** Use **Bootstrap Icons** exclusively (SVG sprite preferred).
- **Color Tokens:** Semantic, intent-based naming only. E.g., `--color-bg-base`, `--color-accent-primary`. Never name a variable after a specific hex color (e.g., no `--gold-accent`).
- **Utility Integration:** Bootstrap 5 utility classes provide the structural grid, but custom CSS custom properties (tokens) dictate all branding.
- **No Inline Styles:** Strictly prohibited. Dynamic values (like progress bars) must be passed as custom inline CSS properties (`style="--progress: 75%;"`).
- **CSS Custom Properties First:** Every repeatable value must be mapped to a root CSS variable. Default to `rem` for spacing constraints, `clamp` for fluidity, and `ms` for duration.
- **No Heavily Rounded Buttons:** All buttons on the site must maintain exactly a 2px rounded border edge (e.g., `border-radius: 2px;`). Never use pill-shaped or heavily rounded buttons.

### 1.4 Security & State Management

- **Zero-Trust Frontend Sandbox:** The Javascript acts solely as a dumb renderer. Absolutely no client-side filtering or calculation logic is to be trusted.
- **Safe Markdown Rendering:** _CRITICAL RULE_: All Markdown parsing into HTML must happen securely on the Go backend. The JS frontend acts as a "dumb renderer" and is strictly forbidden from parsing Markdown to prevent XSS. Formatted reviews/content are delivered as pre-rendered HTML partials.
- **CSP Alignment:** To prepare the static prototype for strict Content Security Policies in the Go integration, all JS must reside in external `.js` modules. **Zero inline `<script>` tags are permitted.**
- **Sanitization Guarantee:** We do not rely on `DOMPurify`. Use `.textContent` for raw data insertion in JS, and Server-Side HTML partials for rich text.

---

## PART 2: BACKEND & GO ENGINEERING PROTOCOL

### 2.1 Development Lifecycle (Data-First)

Strict adherence to this sequence is mandatory for every feature:

1. **Schema:** Define PostgreSQL-compatible tables using the **Media Supertype** pattern.
2. **Models:** Update `internal/models` to align 1:1 with **Schema.org** specifications.
3. **Metadata:** Implement **JSON-LD** generation within the Service Layer for SEO and AI Agent readability.
4. **Logic:** Implement Repository interfaces and Service coordination.
5. **UI:** Render data provided by the Service. The UI must never influence backend logic. Component Slicing: Design HTML conceptually so it can be cleanly sliced into Go `{{define "partial_name"}}` templates.

### 2.2 N-Tier Architecture

- **HTTP Handlers (`cmd/web/`)**: Thin controllers. Handle only request/response mapping and template execution.
- **Service Layer (`internal/service/`)**: Business logic, JSON-LD payload generation, and coordination between relational and vector data.
- **Repository Interface (`internal/repository/`)**: Abstracted SQL definitions.
- **Repository Implementation (`internal/repository/dbrepo/`)**: SQL driver implementations.

### 2.3 Database & Scale Standards

- **Scale:** Design for 250M+ connections. Ensure referential integrity via the `media` Supertype table for all `reviews`, `cast`, and `crew` relations.
- **PostgreSQL Migration:** All SQL must be compatible with PostgreSQL. Use `SERIAL`/`IDENTITY` over `AUTOINCREMENT` and `ILIKE` for case-insensitive searches.
- **Vector Integration:** New content must trigger asynchronous jobs to compute embeddings. Semantic search results from the Vector DB must resolve back to PostgreSQL IDs.

### 2.4 AI & Search Readiness

- **Semantic Truth:** Every entity must serve a valid JSON-LD block in the HTML `<head>`.
- **Vector Readiness:** Repository interfaces must support ID-based resolution for semantic similarity searches.

### 2.5 Go Engineering Requirements

- **Dependency Injection:** Global variables (e.g., `var DB`) are prohibited. Inject all dependencies via `application` or `service` structs.
- **Logging Standards:** Use Go's built-in `slog` package for JSON-formatted logging with contextual data (e.g., request IDs).
- **Prohibited Actions:**
    - Do not bypass the Service Layer to call the Repository directly from a Handler.
    - Do not modify database schemas to accommodate UI limitations.
