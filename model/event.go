package model

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Event struct {
	ID                    int
	Category              string
	Name                  string
	Description           string
	WikipediaSectionTitle string
	ImageFilename         string
	Years                 string
	SlugDB                string // Maps to 'slug' column
	Requests              []Request
}

// Slug returns the identifier used in URLs.
func (e *Event) Slug() string {
	if e.SlugDB != "" {
		return e.SlugDB
	}

	// Fallback generation (shouldn't be needed if DB is populated, but good for safety)
	// 1. Normalize unicode (remove accents)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ := transform.String(t, e.Name)

	// 2. Lowercase
	s = strings.ToLower(s)

	// 3. Replace non-alphanumeric with dash
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")

	// 4. Trim dashes
	s = strings.Trim(s, "-")

	return s
}

func (e *Event) MatchesYears(filterStart, filterEnd int) bool {
	// If no filter provided, everything matches
	if filterStart == 0 && filterEnd == 0 {
		return true
	}

	// Normalize filter end if 0 (open-ended)
	if filterEnd == 0 {
		filterEnd = 10000 // Far future
	}

	// Normalize input: replace en-dash (–) and em-dash (—) with standard hyphen (-)
	normalizedYears := strings.ReplaceAll(e.Years, "–", "-")
	normalizedYears = strings.ReplaceAll(normalizedYears, "—", "-")

	parts := strings.Split(normalizedYears, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Range: "1981-1983"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 != nil || err2 != nil {
				continue
			}

			// Check overlap: max(start, filterStart) <= min(end, filterEnd)
			if start <= filterEnd && end >= filterStart {
				return true
			}
		} else {
			// Single year: "1531"
			// Check if it's "1981-present" style but normalized
			if strings.Contains(part, "present") {
				// Handle "1981-present"
				subParts := strings.Split(part, "-")
				if len(subParts) >= 1 {
					start, err := strconv.Atoi(strings.TrimSpace(subParts[0]))
					if err == nil {
						// treated as start-10000
						if start <= filterEnd && 10000 >= filterStart {
							return true
						}
					}
				}
				continue
			}

			year, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			if year >= filterStart && year <= filterEnd {
				return true
			}
		}
	}
	return false
}
