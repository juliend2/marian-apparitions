package repository

import (
	"database/sql"
	"marianapparitions/model"
)

func GetRequestsByEventID(db *sql.DB, eventID int) ([]model.Request, error) {
	var requests []model.Request
	rows, err := db.Query(
		`SELECT
			id,
			event_id,
			request
		FROM marys_requests
		WHERE event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.Request
		if err := rows.Scan(&r.ID, &r.EventID, &r.Request); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}

	return requests, nil
}

func GetBlocksByEventID(db *sql.DB, eventID int) ([]model.EventBlock, error) {
	var blocks []model.EventBlock
	rows, err := db.Query(
		`SELECT
			id,
			title,
			content,
			event_id,
			ordering,
			COALESCE(church_authority, ''),
			COALESCE(authority_position, '')
		FROM event_blocks
		WHERE event_id = ?
		ORDER BY ordering`, eventID)

	if err != nil {
		return blocks, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.EventBlock
		if err := rows.Scan(&r.ID, &r.Title, &r.Content, &r.EventID, &r.Ordering, &r.ChurchAuthority, &r.AuthorityPosition); err != nil {
			return nil, err
		}
		blocks = append(blocks, r)
	}

	return blocks, nil
}

func GetEventBySlug(db *sql.DB, slug string) (model.Event, error) {
	var e model.Event
	// Query by 'slug' column
	row := db.QueryRow(
		`SELECT
			e.id,
			e.category,
			e.name,
			e.wikipedia_section_title,
			COALESCE(e.image_filename, '') AS image_filename,
			e.years,
			COALESCE(e.slug, '') as slug
		FROM events AS e
		WHERE e.slug = ?`, slug)

	err := row.Scan(&e.ID, &e.Category, &e.Name, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years, &e.SlugDB)
	if err != nil {
		return e, err
	}

	e.Requests, _ = GetRequestsByEventID(db, e.ID)
	e.Blocks, _ = GetBlocksByEventID(db, e.ID)

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
		e.Blocks, err = GetBlocksByEventID(db, e.ID)
		if err != nil {
			panic(err)
		}
		events = append(events, e)
	}
	return events, nil
}
