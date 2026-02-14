package main

import (
	"database/sql"
	"html/template"
	"log"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"marianapparitions/model"
	"marianapparitions/repository"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type SupportedSort struct {
	Name string
	Orientation string
	Slug string
}


var SupportedSorts = []SupportedSort{
	{Name: "Name", Slug: "name_asc", Orientation: "asc"},
	{Name: "Name", Slug: "name_desc", Orientation: "desc"},
	{Name: "Year", Slug: "year_asc", Orientation: "asc"},
	{Name: "Year", Slug: "year_desc", Orientation: "desc"},
	{Name: "Category", Slug: "category_asc", Orientation: "asc"},
	{Name: "Category", Slug: "category_desc", Orientation: "desc"},
}

func main() {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.sqlite3"
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handleIndexOrView)
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

type QueryString struct {
	Key string
	Value string
}

type IndexViewModel struct {
	Events             []*model.Event
	Categories         []string
	SelectedCategories map[string]bool
	StartYear          int
	EndYear            int
	SupportedSorts     []SupportedSort
	CurrentSort        string
	FilterQuery        map[string]string
}

func (vm *IndexViewModel) SortHref(sort SupportedSort) []*QueryString {
	fmt.Printf("SortHref %o \n", sort)
	var query []*QueryString
	if len(vm.FilterQuery) > 0 {
		for key, value := range vm.FilterQuery {
			if key != "sort_by" {
				query = append(query, &QueryString{Key: key, Value: value})
			}
		}
	}
	query = append(query, &QueryString{Key: "sort_by", Value: sort.Slug})

	return query
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
	viewModel := &IndexViewModel{
		Events:             filteredEvents,
		Categories:         categories,
		SelectedCategories: selectedCats,
		StartYear:          startYear,
		EndYear:            endYear,
		SupportedSorts:     SupportedSorts,
		CurrentSort:        sortBy,
		FilterQuery:        buildQueryMap(r.URL.Query()),
	}
	funcMap := template.FuncMap{
		"safeURL": func(u string) template.URL {
			return template.URL(u)
		},
	}
	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
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
