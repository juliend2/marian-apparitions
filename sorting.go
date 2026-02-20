package main

import (
	"sort"
	"strconv"
	"strings"
)

type Sortable interface {
	GetName() string
	GetCategory() string
	GetYears() string
}

func applySorting[T Sortable](events []T, sortBy string) {
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
			less = strings.ToLower(events[i].GetName()) < strings.ToLower(events[j].GetName())
		case "category":
			less = strings.ToLower(events[i].GetCategory()) < strings.ToLower(events[j].GetCategory())
		case "year":
			// Extract first year from years string for comparison
			yearI := extractFirstYear(events[i].GetYears())
			yearJ := extractFirstYear(events[j].GetYears())
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
