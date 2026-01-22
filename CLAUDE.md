# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go web application that catalogs and displays Marian apparitions (religious events). It uses SQLite for data storage and serves dynamic HTML pages with filtering and sorting capabilities.

## Build and Run Commands

**Development:**
```bash
make run              # Run the application locally (go run .)
```

**Production Build:**
```bash
make build            # Build Linux binary (GOOS=linux GOARCH=amd64) to './app'
```

**Clean:**
```bash
make clean            # Remove data.sqlite3 and marianapparitions binary
```

**Testing:**
There are currently no tests in this codebase.

## Architecture

### Project Structure

```
marianapparitions/
├── main.go              # HTTP handlers, routing, DB initialization
├── model/
│   └── event.go        # Event struct, slug generation, year filtering logic
├── repository/
│   └── event.go        # Database queries (GetEventBySlug, GetAllEvents)
├── templates/
│   ├── index.html      # List view with filters/sorting
│   └── view.html       # Single event detail view
├── schema.sql          # SQLite schema definition
└── data.sqlite3        # SQLite database (auto-created)
```

### Key Architectural Patterns

**Database Initialization (main.go:50-136):**
- On startup, checks if `events` table exists; creates from `schema.sql` if missing
- Seeds with initial data if table is empty (Our Lady of Guadalupe, Lourdes, Fátima)
- Auto-migrates `slug` column if missing from existing schema
- Runs `ensureSlugs()` to populate any missing slug values using `Event.Slug()` method

**Routing:**
- Single handler function `handleIndex` registered at `/`
- Routes to `handleView` if path ≠ "/" (catches all `/<slug>` patterns)
- No static file server yet configured (templates reference `/static/` but handler not implemented)

**Filtering Logic:**
- All events fetched from DB, then filtered in-memory (main.go:206-222)
- Year filtering uses `Event.MatchesYears()` which handles:
  - Single years: "1531"
  - Ranges: "1981-1983"
  - Comma-separated: "1531, 1858"
  - Unicode dashes (en-dash, em-dash) normalized to hyphen
  - "present" keyword (treated as year 10000)
- Category filtering supports multi-select checkboxes

**Slug Generation (model/event.go:26-47):**
- Uses DB `slug` column if present (SlugDB field)
- Fallback: normalizes Name (removes accents, lowercase, replaces non-alphanumeric with dashes)
- Slugs are generated on-the-fly during `ensureSlugs()` migration

**Wikipedia Integration:**
- Events store `wikipedia_section_title` field
- Not currently used in templates but intended for future Wikipedia API fetching

### Important Implementation Details

**Year Range Parsing:**
The `MatchesYears()` method (model/event.go:49-111) is complex due to handling multiple formats. When modifying year logic, remember:
- Normalize unicode dashes first
- Handle both single years and ranges
- Check for overlap between event ranges and filter ranges using: `start <= filterEnd && end >= filterStart`

**Database Schema Evolution:**
The `slug` column is NOT in `schema.sql` but added dynamically via ALTER TABLE in `initDB()` (main.go:76-91). This means:
- New databases won't have the slug column initially
- It's safe to modify schema.sql to include slug for fresh installs
- Migration code exists for backward compatibility with old databases

**HTML Templates:**
- Use Go's `html/template` package (not text/template)
- ViewModels passed to templates:
  - `IndexViewModel`: Events, Categories, SelectedCategories, StartYear, EndYear, SupportedSorts
  - `Event` (for view.html): Direct model struct

**Sorting:**
Currently unimplemented despite UI present. `SupportedSorts` variable exists (main.go:19) with values like "name_asc", "year_desc", but TODO comment at main.go:224 indicates sorting logic needs implementation.

## Development Notes

**Port Configuration:**
- Defaults to `:8080`
- Override with `PORT` environment variable

**Database Location:**
- `./data.sqlite3` relative to working directory
- NOT in `.gitignore` (appears in git status as tracked file)

**Module Path:**
Import paths use `marianapparitions/model` and `marianapparitions/repository` (go.mod module name)
