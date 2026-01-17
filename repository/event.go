package repository

import (
	"database/sql"
	"marianapparitions/model"
)

func GetEventBySlug(db *sql.DB, slug string) (model.Event, error) {
	var e model.Event
	// Query by 'slug' column
	row := db.QueryRow(
		`SELECT
			id,
			category,
			name,
			description,
			wikipedia_section_title,
			COALESCE(image_filename, '') AS image_filename,
			years,
            COALESCE(slug, '') as slug
		FROM events
		WHERE slug = ?`, slug)

	err := row.Scan(&e.ID, &e.Category, &e.Name, &e.Description, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years, &e.SlugDB)
	if err == sql.ErrNoRows {
		return e, sql.ErrNoRows
	} else if err != nil {
		return e, err
	}

	return e, nil
}

func GetAllEvents(db *sql.DB) ([]model.Event, error) {
	var events []model.Event
	rows, err := db.Query(
		`SELECT
			id,
			category,
			name,
			description,
			wikipedia_section_title,
			COALESCE(image_filename, '') AS image_filename,
			years,
			COALESCE(slug, '') as slug
		FROM events
		ORDER BY CAST(years AS INTEGER) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Category, &e.Name, &e.Description, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years, &e.SlugDB); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
