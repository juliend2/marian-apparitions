package viewmodel

import (
	"fmt"
)

type SupportedSort struct {
	Name string
	Orientation string
	Slug string
}

type QueryString struct {
	Key string
	Value string
}

type IndexViewModel struct {
	Events             []*EventViewModel
	Categories         []string
	SelectedCategories map[string]bool
	StartYear          int
	EndYear            int
	SupportedSorts     []SupportedSort
	CurrentSort        string
	FilterQuery        map[string]string
}

// SortHref generates a slice of QueryString's
// It also makes sure you won't get a duplicate sort_by key
// if one was passed in the current querystring (through vm.FilterQuery).
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

// GetSortNameByString return the string of the passed sort's name
// The filtering is based on the SupportedSort's Slug field.
func (vm *IndexViewModel) GetSortNameByString(sort string) string {
	for _, sSort := range vm.SupportedSorts {
		if sSort.Slug == sort {
			return sSort.Name
		}
	}
	return ""
}

