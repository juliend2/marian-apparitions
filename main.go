package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"marianapparitions/model"
	"marianapparitions/repository"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var SupportedSorts = []string{"name_asc", "name_desc", "year_asc", "year_desc", "category_asc", "category_desc"}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./data.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handleIndex)
	// We handle /<slug> in a catch-all way or specific pattern.
	// Since handleIndex matches "/", we need to distinguish inside,
	// or register specific paths. But /<slug> is dynamic at root.
	// Standard pattern:
	// "/" -> index (if path is exactly "/")
	// "/<slug>" -> view

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initDB() error {
	// Check if table exists
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='events'").Scan(&name)
	if err == sql.ErrNoRows {
		// Table doesn't exist, create it
		schema, err := os.ReadFile("schema.sql")
		if err != nil {
			return err
		}
		_, err = db.Exec(string(schema))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Seed if empty
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM events")
	if err := row.Scan(&count); err != nil {
		return err
	}

	// Check for 'slug' column and add if missing
	var slugCol string
	err = db.QueryRow("SELECT slug FROM events LIMIT 1").Scan(&slugCol)
	if err != nil {
		// Assume column missing (or table empty) - simplistic check, but works for "add column"
		// If table is empty, this might error, but the seed will run.
		// Better: check specific error or just try to add and ignore error?
		// Safest for sqlite: just try to add, if it fails it fails.
		// BUT better logic: check table info.
		// Let's just run the ALTER and ignore "duplicate column" error or check properly.
		// Actually, db.QueryRow returns error if column doesn't exist.
		if !strings.Contains(err.Error(), "no such column") && err != sql.ErrNoRows {
			// Real error
		} else if strings.Contains(err.Error(), "no such column") {
			_, _ = db.Exec("ALTER TABLE events ADD COLUMN slug TEXT")
		}
	}

	if count == 0 {
		if err := seedData(); err != nil {
			return err
		}
	}

	return ensureSlugs()
}

func ensureSlugs() error {
	rows, err := db.Query("SELECT id, name FROM events WHERE slug IS NULL OR slug = ''")
	if err != nil {
		return err
	}
	defer rows.Close()

	var toUpdate []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return err
		}
		toUpdate = append(toUpdate, e)
	}

	// We need to use the model's Slug() generation on the fly.
	// Since e.SlugDB is empty, e.Slug() will run the generation logic.
	stmt, err := db.Prepare("UPDATE events SET slug = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range toUpdate {
		newSlug := e.Slug()
		_, err := stmt.Exec(newSlug, e.ID)
		if err != nil {
			log.Printf("Failed to update slug for %s: %v", e.Name, err)
		} else {
			log.Printf("Updated slug for event: %s -> %s", e.Name, newSlug)
		}
	}
	return nil
}

func seedData() error {
	_, err := db.Exec(`INSERT INTO events (category, name, description, wikipedia_section_title, image_filename, years) VALUES 
	('Apparition', 'Our Lady of Guadalupe', 'A series of five Marian apparitions in December 1531, and the image on a cloak enshrined within the Basilica of Our Lady of Guadalupe in Mexico City.', 'Our_Lady_of_Guadalupe', 'guadalupe.jpg', '1531'),
	('Apparition', 'Our Lady of Lourdes', 'Apparitions of the Virgin Mary to Saint Bernadette Soubirous in 1858 in the grotto of Massabielle.', 'Our_Lady_of_Lourdes', 'lourdes.jpg', '1858'),
	('Apparition', 'Our Lady of Fátima', 'Reported apparitions to three shepherd children at the Cova da Iria, in Fátima, Portugal.', 'Our_Lady_of_Fátima', 'fatima.jpg', '1917')
	`)
	return err
}

type IndexViewModel struct {
	Events             []*model.Event
	Categories         []string
	SelectedCategories map[string]bool
	StartYear          int
	EndYear            int
	SupportedSorts     []string
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Assume it's a slug if not root
		handleView(w, r)
		return
	}

	// 1. Parse Filters
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startYear, _ := strconv.Atoi(r.FormValue("start_year"))
	endYear, _ := strconv.Atoi(r.FormValue("end_year"))
	sortBy := r.FormValue("sort_by")
	selectedCatsSlice := r.Form["category"] // Multi-value
	selectedCats := make(map[string]bool)
	for _, c := range selectedCatsSlice {
		selectedCats[c] = true
	}

	log.Println("Filters - StartYear:", startYear, "EndYear:", endYear, "SortBy:", sortBy, "Categories:", selectedCatsSlice)

	// 2. Fetch Data (All Events)
	// We fetch all because complex string parsing for years is easier in Go
	allEvents, err := repository.GetAllEvents(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Fetch Categories for Dropdown/Checkboxes
	catRows, err := db.Query("SELECT DISTINCT category FROM events ORDER BY category")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer catRows.Close()
	var categories []string
	for catRows.Next() {
		var c string
		if err := catRows.Scan(&c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}

	// 4. Apply Filters in Memory
	var filteredEvents []*model.Event
	for i := range allEvents {
		e := &allEvents[i]
		// Category Filter
		if len(selectedCats) > 0 {
			if !selectedCats[e.Category] {
				continue
			}
		}

		// Year Filter
		if !e.MatchesYears(startYear, endYear) {
			continue
		}

		filteredEvents = append(filteredEvents, e)
	}

	// 5. Apply Sorting
	if sortBy != "" {
		applySorting(filteredEvents, sortBy)
	}

	// 6. Render
	viewModel := IndexViewModel{
		Events:             filteredEvents,
		Categories:         categories,
		SelectedCategories: selectedCats,
		StartYear:          startYear,
		EndYear:            endYear,
		SupportedSorts:     SupportedSorts,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, viewModel)
}

func applySorting(events []*model.Event, sortBy string) {
	// Split sort_by into field and direction (e.g., "name_asc" -> ["name", "asc"])
	parts := strings.Split(sortBy, "_")
	if len(parts) != 2 {
		return // Invalid format, skip sorting
	}

	field := parts[0]
	direction := parts[1]

	sort.Slice(events, func(i, j int) bool {
		var less bool

		switch field {
		case "name":
			less = strings.ToLower(events[i].Name) < strings.ToLower(events[j].Name)
		case "category":
			less = strings.ToLower(events[i].Category) < strings.ToLower(events[j].Category)
		case "year":
			// Extract first year from years string for comparison
			yearI := extractFirstYear(events[i].Years)
			yearJ := extractFirstYear(events[j].Years)
			less = yearI < yearJ
		default:
			return false // Unknown field
		}

		// Reverse if descending
		if direction == "desc" {
			return !less
		}
		return less
	})
}

func extractFirstYear(years string) int {
	// Normalize dashes
	normalizedYears := strings.ReplaceAll(years, "–", "-")
	normalizedYears = strings.ReplaceAll(normalizedYears, "—", "-")

	// Split by comma and take first part
	parts := strings.Split(normalizedYears, ",")
	if len(parts) == 0 {
		return 0
	}

	firstPart := strings.TrimSpace(parts[0])

	// If it's a range, take the start year
	if strings.Contains(firstPart, "-") {
		rangeParts := strings.Split(firstPart, "-")
		firstPart = strings.TrimSpace(rangeParts[0])
	}

	// Convert to int
	year, err := strconv.Atoi(firstPart)
	if err != nil {
		return 0
	}

	return year
}

func handleView(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/")

	e, err := repository.GetEventBySlug(db, slug)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/view.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, e)
}
