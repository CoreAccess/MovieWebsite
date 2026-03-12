# Master Database Restructuring Plan
## Schema.org-Aligned + Competitive + Schema Normalization

> [!IMPORTANT]
> Implementation requires deleting `movieweb.db` and restarting. The server re-seeds from TMDB automatically.
> **Authorized implementation only — do not execute until approved.**

---

## Monetization Context

Before touching the schema, here's the affiliate landscape so we build the right hooks:

| Service | Program | Commission | Cookie |
|---|---|---|---|
| **Disney+** | Has affiliate program | ~$16/signup | 30 days |
| **Hulu** | Has affiliate program | $9.60/sub, $1.60/trial | 14 days |
| **Amazon Prime** | Amazon Associates bounty | $3.00/signup | 24 hours |
| **Netflix** | ❌ No affiliate program | — | — |
| **Paramount+** | Available via networks | Varies | Varies |

**Strategy:** The `streaming_providers` table should include an `affiliate_url` column beside the generic `website_url`. Each provider entry can have a tracked link. The `media_availability` table then links titles to providers — every "Watch on Disney+" button becomes an affiliate click.

**Future monetization (no traditional ads):**
- **Sponsored Featured Titles** — Movie studios pay to be in "Featured" or "Coming Soon" homepage slots. Requires a `sponsored_placements` table (flagged as "Promoted" — transparent to users, valuable to studios)
- **Curated Editorial Content** — Paid "Staff Pick" lists for a studio's release (labeled transparently)

These don't go in the DB now, but the `streaming_providers.affiliate_url` column sets the foundation immediately.

---

## Tier 1 — Fix Broken Structures (Normalization)

### 1. Genres
One shared lookup. Two junction tables for referential integrity (FK can't span two parent tables via text column).
```
genres (id, name UNIQUE, slug UNIQUE)
movie_genres  (movie_id → movies, genre_id → genres, PK both)
tv_genres     (series_id → tv_series, genre_id → genres, PK both)
```
**Remove:** `movies.genre JSON`, `tv_series.genre JSON`

### 2. Keywords
```
keywords (id, name UNIQUE, slug UNIQUE)
movie_keywords  (movie_id → movies, keyword_id → keywords, PK both)
tv_keywords     (series_id → tv_series, keyword_id → keywords, PK both)
```
**Remove:** `movies.keywords TEXT`

### 3. Production Companies
Remove the redundant free-text columns. The existing `production_companies` junction table (→ `organizations`) is already correct — just needs to be populated during seeding.
**Remove:** `movies.production_company TEXT`, `tv_series.production_company TEXT`

### 4. Languages
```
languages (code TEXT PK — ISO 639-1, name TEXT)  e.g. "en" → "English"
```
**Replace:** `movies.in_language` and `tv_series.in_language` → `language_code REFERENCES languages(code)`

### 5. Countries
```
countries (code TEXT PK — ISO 3166-1, name TEXT)  e.g. "US" → "United States"
```
**Add:** `movies.country_code`, `tv_series.country_code` both reference `countries(code)`

### 6. External Identifiers
Cross-referencing IMDB, TMDB, EIDR, Wikidata.
```
external_ids
(media_type TEXT, media_id INTEGER, source TEXT, external_id TEXT)
source = "imdb" | "tmdb" | "eidr" | "wikidata" | "rottentomatoes"
UNIQUE (media_type, media_id, source)
```
Enables "View on IMDB" links, future API syncing, deduplication.

---

## Tier 2 — Structural Improvements

### 7. Reviews — Proper structured table
*(schema.org/UserReview + CriticReview)*
```
reviews
(id, user_id → users, media_type, media_id,
 rating REAL NOT NULL,
 title TEXT, body TEXT,
 positive_notes TEXT, negative_notes TEXT,
 contains_spoilers BOOLEAN DEFAULT 0,
 review_type TEXT DEFAULT 'user'  ← "user" | "critic" | "press"
 publication_name TEXT nullable,   ← for critic reviews
 external_review_url TEXT nullable,
 status TEXT DEFAULT 'published',
 created_at, updated_at TIMESTAMP,
 UNIQUE (user_id, media_type, media_id))
```

### 8. AggregateRating — Decomposed
*(schema.org/AggregateRating: ratingValue, ratingCount, reviewCount, bestRating, worstRating)*

**Add to `movies` and `tv_series`:**
`rating_count INTEGER DEFAULT 0`, `review_count INTEGER DEFAULT 0`, `best_rating REAL DEFAULT 10.0`, `worst_rating REAL DEFAULT 1.0`

### 9. Rating Demographics
*(IMDB-style age-segmented ratings — no gender segmentation)*
```
rating_demographics
(media_type, media_id,
 age_group TEXT  ← "under_18" | "18_29" | "30_44" | "45_plus"
 avg_rating REAL, vote_count INTEGER,
 UNIQUE (media_type, media_id, age_group))
```
Populated from `reviews` table based on reviewer's age.

### 10. Media Images — Gallery
```
media_images
(id, media_type, media_id,
 image_type TEXT  ← "poster" | "backdrop" | "still" | "logo" | "banner"
 url TEXT, is_primary BOOLEAN DEFAULT 0,
 source TEXT  ← "tmdb" | "user_upload"
 language_code TEXT nullable  ← for localized posters)
```

### 11. Source Material — isBasedOn
*(schema.org/isBasedOn — "Based on the novel by...")*
```
source_material
(id, title TEXT, source_type TEXT, author_id → people nullable, year INTEGER)
source_type = "book" | "novel" | "comic" | "game" | "play" | "true_story" | "podcast" | "remake"

movie_source_material (movie_id → movies, source_id, PK both)
tv_source_material    (series_id → tv_series, source_id, PK both)
```

### 12. People — Aliases and Awards
```
person_aliases (person_id → people, alias TEXT)

award_bodies      (id, name UNIQUE, slug, website_url)
award_ceremonies  (id, body_id → award_bodies, year INTEGER, ceremony_number INTEGER, date_held DATE)
award_categories  (id, ceremony_id, name TEXT "Best Picture", department TEXT)
award_nominations (id, category_id, media_type TEXT, media_id INTEGER,
                   person_id → people nullable, won BOOLEAN DEFAULT 0, nominee_note TEXT)
```
This replaces the flat `awards` table. Real hierarchy: Body → Ceremony → Category → Nomination.  
**Remove:** `people.also_known_as TEXT`, `people.awards TEXT`

### 13. People — Additional Fields
- `people.nationality_code TEXT REFERENCES countries(code)` — *(schema.org/nationality)*
- `people.known_for_department TEXT` — "Acting" | "Directing" | "Writing" | "Production" (avoids JOIN on every list render)
- `people.popularity_score REAL DEFAULT 0` — Updated periodically, enables "Trending Actors"

### 14. Polls — Proper voting
```
poll_options (id, poll_id → polls, option_text TEXT)
poll_votes   (user_id → users, option_id → poll_options, voted_at TIMESTAMP,
              PK (user_id, option_id)  ← prevents double-voting)
```
**Remove:** `polls.options TEXT` JSON blob

### 15. Ad Campaign Targets
```
campaign_targets (campaign_id → ad_campaigns, page_slug TEXT, PK both)
```
**Remove:** `ad_campaigns.target_pages TEXT`

### 16. Bug Fixes
- `UserNotificationSettings.UserID bool` → `int` in models.go
- `edit_history`: replace `changes TEXT` → `field TEXT`, `old_value TEXT`, `new_value TEXT`
- Add `created_at / updated_at TIMESTAMP` to `movies`, `tv_series`, `people`, `characters`, `organizations` (currently missing — can't answer "recently added/updated")

### 17. Posts — Media link
Add nullable `media_type TEXT`, `media_id INTEGER` to `posts`. Enables per-title review threads.

---

## Tier 3 — New Schema.org Entities

### 18. TVSeason — First-class entity
```
tv_seasons (id, series_id → tv_series, season_number INTEGER,
            name TEXT, description TEXT, image TEXT,
            date_published TEXT, episode_count INTEGER, aggregate_rating REAL,
            UNIQUE (series_id, season_number))
```

### 19. MovieSeries — Franchise grouping
```
movie_series (id, name TEXT UNIQUE, slug, description TEXT, image TEXT)
movie_series_entries (series_id → movie_series, movie_id → movies, position INTEGER, PK both)
```

### 20. Quotations
*(schema.org/Quotation — "Here's looking at you, kid")*
```
quotations (id, media_type, media_id,
            person_id → people nullable, character_id → characters nullable,
            quote_text TEXT NOT NULL, scene_context TEXT,
            submitted_by → users, status TEXT DEFAULT 'published')
```
Powers: "Quote of the Day", Trivia, Gamification.

### 21. Screening Events
*(schema.org/ScreeningEvent)*
```
screening_events (id, media_type, media_id,
                  event_type TEXT  ← "world_premiere"|"festival"|"streaming_debut"|"limited_release"
                  event_name TEXT, location TEXT, event_date DATE, description TEXT)
```

### 22. Television Networks
*(schema.org/BroadcastChannel / TelevisionChannel)*
```
networks (id, name UNIQUE, slug, network_type TEXT, country_code → countries, logo_url, website_url)
network_type = "broadcast" | "cable" | "streaming" | "premium"

tv_networks (series_id → tv_series, network_id → networks, PK both)
```

---

## Tier 4 — Competitive Features (from IMDB/TMDB Analysis)

### 23. Streaming Providers + Affiliate Revenue 🔴
```
streaming_providers
(id, name UNIQUE  ← "Netflix"|"Disney+"|"Hulu"|"Prime Video"
 logo_url TEXT, website_url TEXT,
 affiliate_url TEXT nullable  ← tracked affiliate link for revenue
 provider_type TEXT  ← "svod"|"avod"|"tvod"|"free"
 has_affiliate_program BOOLEAN DEFAULT 0)

media_availability
(media_type, media_id,
 provider_id → streaming_providers,
 country_code → countries,
 availability_type TEXT  ← "subscription"|"rent"|"buy"|"free"
 available_from DATE nullable, available_until DATE nullable,
 UNIQUE (media_type, media_id, provider_id, country_code))
```
Every "Watch on Disney+" button routes through `affiliate_url`. No extra user friction.

### 24. Regional Release Dates 🔴
```
release_dates
(media_type, media_id, country_code → countries,
 release_date DATE NOT NULL,
 release_type TEXT  ← "theatrical"|"streaming"|"digital"|"dvd"|"festival"
 certification TEXT  ← localized rating per country, e.g. "R","15","MA15+"
 notes TEXT nullable,
 UNIQUE (media_type, media_id, country_code, release_type))
```

### 25. Filming Locations 🔴
```
filming_locations
(id, media_type, media_id,
 location_name TEXT NOT NULL  ← "Central Park, New York City"
 country_code → countries,
 latitude REAL nullable, longitude REAL nullable,
 description TEXT nullable  ← "Used as exterior of Wayne Manor"
 is_real_world BOOLEAN DEFAULT 1)
```
Enables: "Films shot in France" discovery, future map feature.

### 26. Taglines 🔴
Simple column adds:
- `movies.tagline TEXT`
- `tv_series.tagline TEXT`

### 27. Technical Specifications 🟡
*(IMDB Technical Specs section)*
```
technical_specs  (one row per title)
(media_type, media_id,
 color_type TEXT  ← "Color"|"Black and White"|"Colorized"
 aspect_ratio TEXT  ← "2.39:1"|"1.85:1"|"1.33:1"
 sound_mix TEXT  ← "Dolby Atmos"|"DTS:X"|"Stereo"
 negative_format TEXT  ← "35mm"|"Digital"|"IMAX"
 camera TEXT  ← "Arri Alexa"|"RED Dragon"
 runtime_minutes INTEGER nullable  ← more precise than duration column)
```

### 28. Content Advisory / Parent's Guide 🟡
*(IMDB's detailed ratings system — much richer than just "R" or "PG-13")*
```
content_advisory
(id, media_type, media_id,
 category TEXT  ← "sex_nudity"|"violence_gore"|"profanity"|"substance_use"|"intense_scenes"
 severity_level TEXT  ← "none"|"mild"|"moderate"|"severe"
 notes TEXT,
 submitted_by → users, status TEXT DEFAULT 'published')
```

### 29. Social Links 🟡
*(TMDB official social media per title/person)*
```
social_links
(entity_type TEXT, entity_id INTEGER,
 platform TEXT  ← "instagram"|"twitter"|"facebook"|"tiktok"|"youtube"|"website"
 url TEXT, username TEXT,
 UNIQUE (entity_type, entity_id, platform))
```

### 30. Popularity Snapshots 🟡
*(IMDB MOVIEmeter / TMDB popularity trends — historical data)*
```
popularity_snapshots
(media_type, media_id,
 snapshot_date DATE NOT NULL,
 popularity_score REAL,
 rank_position INTEGER,
 UNIQUE (media_type, media_id, snapshot_date))
```

### 31. Episode Cast 🟡
*(Guest stars at the episode level — critical for anthology shows)*
```
episode_cast
(episode_id → tv_episodes, person_id → people,
 character_id → characters nullable,
 billing_order INTEGER,
 credit_type TEXT DEFAULT 'regular'  ← "regular"|"guest_star"|"cameo"|"voice"
 PK (episode_id, person_id, character_id))
```

### 32. User Lists 🟡
*(Distinct from watchlists — curated, rankable, shareable collections)*
```
user_lists (id, user_id → users, name TEXT, description TEXT,
            is_ranked BOOLEAN DEFAULT 0, is_public BOOLEAN DEFAULT 1, created_at, updated_at)

user_list_items (list_id → user_lists, media_type, media_id,
                 position INTEGER nullable, note TEXT, added_at TIMESTAMP,
                 UNIQUE (list_id, media_type, media_id))
```

### 33. Watch History 🔵
*(What you've actually watched — distinct from "want to watch" lists)*
```
watch_history
(user_id → users, media_type TEXT, media_id INTEGER,
 watched_at TIMESTAMP, rewatch_count INTEGER DEFAULT 0,
 quick_rating REAL nullable  ← star tap without a full review
 UNIQUE (user_id, media_type, media_id))

episode_watch_history
(user_id → users, episode_id → tv_episodes, watched_at TIMESTAMP, UNIQUE both)
```

### 34. User Follow Graph 🔵
*(Social layer — "reviews from people I follow" — neither IMDB nor TMDB does this well)*
```
user_follows (follower_id → users, followed_id → users, followed_at TIMESTAMP, PK both)
```

### 35. Mood / Vibe Tagging 🔵
*(Innovation beyond TMDB — community-voted emotional fingerprinting of content)*
```
moods (id, name UNIQUE  ← "Edge-of-Seat"|"Feel-Good"|"Heartbreaking"|"So-Bad-It's-Good"
       emoji TEXT, description TEXT)

media_moods    (media_type, media_id, mood_id → moods, vote_count INTEGER, PK first three)
user_mood_votes (user_id → users, media_type, media_id, mood_id, PK all four)
```

### 36. On This Day — First-Class Events 🔵
```
on_this_day_events
(id, month INTEGER, day INTEGER,
 entity_type TEXT  ← "movie"|"person"|"award"
 entity_id INTEGER,
 event_type TEXT  ← "born"|"died"|"released"|"premiered"|"won_award"
 year INTEGER, description TEXT)
```

### 37. Notifications — DB Table 🔵
*(Exists in models.go but not in the DB schema)*
```
notifications
(id, user_id → users,
 type TEXT  ← "review_like"|"new_follower"|"reply"|"new_episode"|"award_announced"
 actor_id → users nullable, entity_type TEXT, entity_id INTEGER,
 message TEXT, link TEXT, is_read BOOLEAN DEFAULT 0, created_at TIMESTAMP)
```

---

## All New Tables Summary (37 items total)

| # | Table | Tier | Priority |
|---|---|---|---|
| 1 | `genres` | 1 | 🔴 |
| 2 | `movie_genres` | 1 | 🔴 |
| 3 | `tv_genres` | 1 | 🔴 |
| 4 | `keywords` | 1 | 🔴 |
| 5 | `movie_keywords` | 1 | 🔴 |
| 6 | `tv_keywords` | 1 | 🔴 |
| 7 | `languages` | 1 | 🔴 |
| 8 | `countries` | 1 | 🔴 |
| 9 | `external_ids` | 1 | 🔴 |
| 10 | `reviews` | 2 | 🔴 |
| 11 | `rating_demographics` | 2 | 🟡 |
| 12 | `media_images` | 2 | 🟡 |
| 13 | `source_material` | 2 | 🟡 |
| 14 | `movie_source_material` | 2 | 🟡 |
| 15 | `tv_source_material` | 2 | 🟡 |
| 16 | `person_aliases` | 2 | 🟡 |
| 17 | `award_bodies` | 2 | 🟡 |
| 18 | `award_ceremonies` | 2 | 🟡 |
| 19 | `award_categories` | 2 | 🟡 |
| 20 | `award_nominations` | 2 | 🟡 |
| 21 | `poll_options` | 2 | 🟡 |
| 22 | `poll_votes` | 2 | 🟡 |
| 23 | `campaign_targets` | 2 | 🟡 |
| 24 | `tv_seasons` | 3 | 🟡 |
| 25 | `movie_series` | 3 | 🟡 |
| 26 | `movie_series_entries` | 3 | 🟡 |
| 27 | `quotations` | 3 | 🔵 |
| 28 | `screening_events` | 3 | 🔵 |
| 29 | `networks` | 3 | 🟡 |
| 30 | `tv_networks` | 3 | 🟡 |
| 31 | `streaming_providers` | 4 | 🔴 |
| 32 | `media_availability` | 4 | 🔴 |
| 33 | `release_dates` | 4 | 🔴 |
| 34 | `filming_locations` | 4 | 🟡 |
| 35 | `technical_specs` | 4 | 🟡 |
| 36 | `content_advisory` | 4 | 🟡 |
| 37 | `social_links` | 4 | 🟡 |
| 38 | `popularity_snapshots` | 4 | 🟡 |
| 39 | `episode_cast` | 4 | 🟡 |
| 40 | `user_lists` | 4 | 🟡 |
| 41 | `user_list_items` | 4 | 🟡 |
| 42 | `watch_history` | 4 | 🔵 |
| 43 | `episode_watch_history` | 4 | 🔵 |
| 44 | `user_follows` | 4 | 🔵 |
| 45 | `moods` | 4 | 🔵 |
| 46 | `media_moods` | 4 | 🔵 |
| 47 | `user_mood_votes` | 4 | 🔵 |
| 48 | `on_this_day_events` | 4 | 🔵 |
| 49 | `notifications` | 4 | 🔵 |

## Columns to Remove
| Table | Column |
|---|---|
| `movies` | `genre`, `keywords`, `production_company`, `in_language` |
| `tv_series` | `genre`, `production_company`, `in_language` |
| `people` | `also_known_as`, `awards` |
| `ad_campaigns` | `target_pages` |
| `polls` | `options` |

## Columns to Add
| Table | New Columns |
|---|---|
| `movies` | `language_code`, `country_code`, `tagline`, `rating_count`, `review_count`, `best_rating`, `worst_rating`, `is_family_friendly`, `subtitle`, `created_at`, `updated_at` |
| `tv_series` | `language_code`, `country_code`, `tagline`, `rating_count`, `review_count`, `best_rating`, `worst_rating`, `subtitle`, `created_at`, `updated_at` |
| `people` | `nationality_code`, `known_for_department`, `popularity_score`, `created_at`, `updated_at` |
| `characters` | `created_at`, `updated_at` |
| `organizations` | `created_at`, `updated_at` |
| `posts` | `media_type` (nullable), `media_id` (nullable) |
| `edit_history` | `field`, `old_value`, `new_value` (replaces `changes`) |
| `reviews` | `review_type`, `publication_name`, `external_review_url` |

## Models to Fix
- `UserNotificationSettings.UserID` → change from `bool` to `int`
- Add `Genres []Genre` and `Keywords []string` to `Movie` and `TVSeries` structs

---

## Verification Plan
- `go run ./cmd/web` starts with no fatal DB errors
- `GET /` — homepage loads, genres render from new junction tables
- Trending section renders with proper genre badges from genre relations
- Spot-check: `external_ids` populated from TMDB seed, `genres` table has data
