## 2024-05-23 - Add Authentication Rate Limiting
**Vulnerability:** The `/login` and `/signup` endpoints had no specific rate limits, only relying on a generic IP rate limit of 100 requests per minute. This allowed brute-force password guessing and account enumeration attacks to proceed relatively fast.
**Learning:** Sensitive authentication endpoints should always have stricter specific rate-limiting to prevent brute-force attacks and reduce load from potentially malicious traffic independently from the main app's traffic patterns.
**Prevention:** Implement endpoint-specific strict rate limiting rules (e.g. 5 requests per minute) for all authentication-related routes.
