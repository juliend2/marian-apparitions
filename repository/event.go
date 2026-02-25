package repository

import (
	"context"
	"database/sql"

	"marianapparitions/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("marianapparitions/repository")

const dbSystem = "sqlite"

func GetRequestsByEventID(db *sql.DB, eventID int) ([]model.Request, error) {
	return GetRequestsByEventIDContext(context.Background(), db, eventID)
}

func GetRequestsByEventIDContext(ctx context.Context, db *sql.DB, eventID int) ([]model.Request, error) {
	const query = `SELECT id, event_id, request FROM marys_requests WHERE event_id = ?`
	ctx, span := tracer.Start(ctx, "GetRequestsByEventID")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.system", dbSystem),
		attribute.String("db.statement", query),
	)

	var requests []model.Request
	rows, err := db.QueryContext(ctx, query, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.Request
		if err := rows.Scan(&r.ID, &r.EventID, &r.Request); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		requests = append(requests, r)
	}

	return requests, nil
}

func GetBlocksByEventID(db *sql.DB, eventID int) ([]model.EventBlock, error) {
	return GetBlocksByEventIDContext(context.Background(), db, eventID)
}

func GetBlocksByEventIDContext(ctx context.Context, db *sql.DB, eventID int) ([]model.EventBlock, error) {
	const query = `SELECT id, title, content, event_id, ordering, COALESCE(church_authority, ''), COALESCE(authority_position, '') FROM event_blocks WHERE event_id = ? ORDER BY ordering`
	ctx, span := tracer.Start(ctx, "GetBlocksByEventID")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.system", dbSystem),
		attribute.String("db.statement", query),
	)

	var blocks []model.EventBlock
	rows, err := db.QueryContext(ctx, query, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return blocks, err
	}
	defer rows.Close()

	for rows.Next() {
		var r model.EventBlock
		if err := rows.Scan(&r.ID, &r.Title, &r.Content, &r.EventID, &r.Ordering, &r.ChurchAuthority, &r.AuthorityPosition); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		blocks = append(blocks, r)
	}

	return blocks, nil
}

func GetEventBySlug(db *sql.DB, slug string) (model.Event, error) {
	return GetEventBySlugContext(context.Background(), db, slug)
}

func GetEventBySlugContext(ctx context.Context, db *sql.DB, slug string) (model.Event, error) {
	const query = `SELECT e.id, e.category, e.name, e.wikipedia_section_title, COALESCE(e.image_filename, '') AS image_filename, e.years, COALESCE(e.slug, '') as slug, COALESCE(e.country, '') as country FROM events AS e WHERE e.slug = ?`
	ctx, span := tracer.Start(ctx, "GetEventBySlug")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.system", dbSystem),
		attribute.String("db.statement", query),
	)

	var e model.Event
	row := db.QueryRowContext(ctx, query, slug)
	err := row.Scan(&e.ID, &e.Category, &e.Name, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years, &e.SlugDB, &e.Country)
	if err != nil {
		if err != sql.ErrNoRows {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return e, err
	}

	e.Requests, _ = GetRequestsByEventIDContext(ctx, db, e.ID)
	e.Blocks, _ = GetBlocksByEventIDContext(ctx, db, e.ID)

	return e, nil
}

func GetAllEvents(db *sql.DB) ([]model.Event, error) {
	return GetAllEventsContext(context.Background(), db)
}

func GetAllEventsContext(ctx context.Context, db *sql.DB) ([]model.Event, error) {
	const query = `SELECT id, category, name, description, wikipedia_section_title, COALESCE(image_filename, '') AS image_filename, years, COALESCE(slug, '') as slug, COALESCE(country, '') as country FROM events ORDER BY CAST(years AS INTEGER) DESC`
	ctx, span := tracer.Start(ctx, "GetAllEvents")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.system", dbSystem),
		attribute.String("db.statement", query),
	)

	var events []model.Event
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Category, &e.Name, &e.Description, &e.WikipediaSectionTitle, &e.ImageFilename, &e.Years, &e.SlugDB, &e.Country); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		var blockErr error
		e.Blocks, blockErr = GetBlocksByEventIDContext(ctx, db, e.ID)
		if blockErr != nil {
			panic(blockErr)
		}
		events = append(events, e)
	}
	return events, nil
}
