package model

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
