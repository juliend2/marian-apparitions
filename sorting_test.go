package main

import (
	"testing"
)

func TestExtractFirstYear(t *testing.T) {
	extracted := extractFirstYear("2026")
	if extracted != 2026 {
		t.Errorf("Wrong value: %v", extracted)
	}
}
