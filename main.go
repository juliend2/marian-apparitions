package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	ID                    int
	Category              string
	Name                  string
	Description           string
	WikipediaSectionTitle string
	ImageFilename         string
	Years                 string
}

// Slug returns the identifier used in URLs.
// We use WikipediaSectionTitle as the slug.
func (e *Event) Slug() string {
	return e.WikipediaSectionTitle
}

var db *sql.DB

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

	if count == 0 {
		return seedData()
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

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Assume it's a slug if not root
		handleView(w, r)
		return
	}

	rows, err := db.Query(
		`SELECT
			id,
			category,
			name,
			description,
			wikipedia_section_title,
			COALESCE(image_filename, '') AS image_filename,
			years
		FROM events`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Category, &e.Name, &e.Description, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, events)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/")

	var e Event
	// Assuming WikipediaSectionTitle matches the slug in URL
	row := db.QueryRow(
		`SELECT
			id,
			category,
			name,
			description,
			wikipedia_section_title,
			COALESCE(image_filename, '') AS image_filename,
			years
		FROM events
		WHERE wikipedia_section_title = ?`, slug)

	err := row.Scan(&e.ID, &e.Category, &e.Name, &e.Description, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years)
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
