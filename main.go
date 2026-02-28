package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"marianapparitions/repository"
	"marianapparitions/viewmodel"

	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const DEFAULT_DB_PATH = "./data.sqlite3"
const DEFAULT_PORT = "8080"
const DEFAULT_SORT = "year_desc"

var db *sql.DB

var SupportedSorts = []viewmodel.SupportedSort{
	{Name: "Name", Slug: "name_asc", Orientation: "asc"},
	{Name: "Name", Slug: "name_desc", Orientation: "desc"},
	{Name: "Year", Slug: "year_asc", Orientation: "asc"},
	{Name: "Year", Slug: "year_desc", Orientation: "desc"},
	{Name: "Category", Slug: "category_asc", Orientation: "asc"},
	{Name: "Category", Slug: "category_desc", Orientation: "desc"},
}

func main() {
	ctx := context.Background()

	shutdown, err := initTelemetry(ctx)
	if err != nil {
		log.Printf("Warning: failed to initialize telemetry: %v", err)
	} else {
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Printf("Warning: telemetry shutdown error: %v", err)
			}
		}()
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = DEFAULT_DB_PATH
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", handleIndexOrView)

	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, otelhttp.NewHandler(mux, "marianapparitions")))
}


func handleIndexOrView(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Assume it's a slug if not root
		handleView(w, r)
		return
	}

	// Specific to the index:

	// 1. Parse Filters
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startYear, _ := strconv.Atoi(r.FormValue("start_year"))
	endYear, _ := strconv.Atoi(r.FormValue("end_year"))
	sortBy := r.FormValue("sort_by")
	if sortBy == "" {
		sortBy = DEFAULT_SORT // Default sort (see repository.GetAllEvents()'s SQL query)
	}
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
	var filteredEvents []*viewmodel.EventViewModel
	for i := range allEvents {
		e := &allEvents[i]
		// Category Filter
		if len(selectedCats) > 0 && !selectedCats[e.Category] {
			continue
		}

		// Year Filter
		if !e.MatchesYears(startYear, endYear) {
			continue
		}

		filteredEvents = append(filteredEvents, viewmodel.NewEventVM(e))
	}

	// 5. Apply Sorting
	if sortBy != "" {
		applySorting[*viewmodel.EventViewModel](filteredEvents, sortBy)
	}

	// 6. Render
	viewModel := &viewmodel.IndexViewModel{
		Events:             filteredEvents,
		Categories:         categories,
		SelectedCategories: selectedCats,
		StartYear:          startYear,
		EndYear:            endYear,
		SupportedSorts:     SupportedSorts,
		CurrentSort:        sortBy,
		FilterQuery:        buildQueryMap(r.URL.Query()),
	}
	tmpl, err := template.New("index.html").ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, viewModel)
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
