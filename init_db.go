package main

import (
	"database/sql"
	"log"
	"marianapparitions/model"
	"os"
	"strings"
)

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
