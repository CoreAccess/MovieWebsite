## 2024-05-23 - Add Authentication Rate Limiting
**Vulnerability:** The `/login` and `/signup` endpoints had no specific rate limits, only relying on a generic IP rate limit of 100 requests per minute. This allowed brute-force password guessing and account enumeration attacks to proceed relatively fast.
**Learning:** Sensitive authentication endpoints should always have stricter specific rate-limiting to prevent brute-force attacks and reduce load from potentially malicious traffic independently from the main app's traffic patterns.
**Prevention:** Implement endpoint-specific strict rate limiting rules (e.g. 5 requests per minute) for all authentication-related routes.

## 2026-03-15 - Fix SQL Injection in ORDER BY clauses
**Vulnerability:** The `GetAllMovies` and `GetAllShows` repository methods used `fmt.Sprintf` to directly inject user-provided sorting parameters into the SQL `ORDER BY` clause. This created a potential SQL injection vector as SQL identifiers cannot be parameterized using standard bind variables.
**Learning:** Never use string interpolation to build SQL queries with user-controlled input, especially for components like `ORDER BY` or `LIMIT` that don't support standard parameterization.
**Prevention:** Use an explicit whitelist or a switch statement to map user-provided input to static, hardcoded SQL query strings or fragments.
